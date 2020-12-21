package network

import (
	"bucket/log"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error){
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip

	n := &Network{
		Name: name,
		IPRange: ipRange,
	}
	err := b.initBridge(n)

	if err != nil {
		log.ConsoleLog.Error("error init bridge: %v", err)
	}

	return n, err
}

func (b *BridgeNetworkDriver) Delete(network Network) error{
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error{
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName: "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil{
		return fmt.Errorf("Error Add Endpoint Device: %v", err)
	}

	return nil
}

func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error{
	return nil
}

func (b *BridgeNetworkDriver) initBridge(n *Network) error{
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil{
		return fmt.Errorf("Error add bridge: %s, Error: %v", bridgeName, err)
	}

	gatewayIP := *n.IPRange
	gatewayIP.IP = n.IPRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil{
		return fmt.Errorf("Error assigning address: %s on bridge: %s with an error of: %v", gatewayIP, bridgeName, err)
	}

	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("Error set bridge up: %s, Error: %v", bridgeName, err)
	}

	if err := setupIPTables(bridgeName, n.IPRange); err != nil {
		return fmt.Errorf("Error setting ipatbles for %s: %v", bridgeName, err)
	}

	return nil
}

func createBridgeInterface(bridgeName string) error{
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{la, nil, nil, nil}
	if err := netlink.LinkAdd(br); err != nil{
		return fmt.Errorf("Bridge creation faild for bridge %s: %v", bridgeName, err)
	}

	return nil
}

func setInterfaceIP(name string, rawIP string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error get interface: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil{
		return err
	}

	addr := &netlink.Addr{ipNet, "", 0, 0, nil, nil, 0 ,0}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("Error retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("Error enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	//err := cmd.Run()
	output, err := cmd.Output()
	if err != nil {
		log.ConsoleLog.Error("iptables Output, %v", output)
	}
	return err
}