package network

import (
	"bucket/container"
	"bucket/log"
	"bucket/utils"
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/run/bucket/network/network/"
	drivers = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

type Network struct {
	Name string
	IPRange *net.IPNet
	Driver string
}

type Endpoint struct {
	ID string `json:"id"`
	Device netlink.Veth `json:"dev"`
	IPAddress net.IP `json:"ip"`
	MacAddress net.HardwareAddr `json:"mac"`
	PortMapping []string `json:"portmapping"`
	Network *Network
}

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

func (nw *Network) dump(dumpPath string) error{
	exist, _ := utils.PathExists(dumpPath)
	if !exist {
		_ = os.MkdirAll(dumpPath, 0644)
	}

	nwPath := path.Join(dumpPath, nw.Name)
	newFile, err := os.OpenFile(nwPath, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = newFile.Close()
	}()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = newFile.Write(nwJson)
	if err != nil {
		return err
	}

	return nil
}

func (nw *Network) load(dumpPath string) error {
	nwCfgFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}

	nwJson := make([]byte, 2000)
	n, err := nwCfgFile.Read(nwJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nwJson[:n], nw)

	if err != nil {
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	exist, err := utils.PathExists(dumpPath)
	if err != nil {
		return err
	}
	if exist {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
	return nil
}

func CreateNetwork(driver, subnet, name string) error{
	_, cidr, _ := net.ParseCIDR(subnet)
	gateWayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}

	cidr.IP = gateWayIP
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo)func(){
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.ConsoleLog.Error("get container net namespace error: %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	if err = netlink.LinkSetNsFd(*enLink,  int(nsFD)); err != nil {
		log.ConsoleLog.Error("error set link netns: %v", err)
	}

	origns, err := netns.Get()
	if err != nil {
		log.ConsoleLog.Error("error set link netns, %v", err)
	}

	if err := netns.Set(netns.NsHandle(nsFD)); err != nil{
		log.ConsoleLog.Error("error set netns, %v", err)
	}

	return func() {
		_ = netns.Set(origns)
		_ = origns.Close()
		runtime.UnlockOSThread()
		_ = f.Close()
	}
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	defer enterContainerNetns(&peerLink, cinfo)()

	interfaceIP := *ep.Network.IPRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil{
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw: ep.Network.IPRange.IP,
		Dst: cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error{
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2{
			log.ConsoleLog.Error("port mapping format error, %v", pm)
		}

		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])

		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.ConsoleLog.Error("iptables output, %v", output)
			continue
		}
	}

	return nil
}

func Connect(networkName string, cinfo *container.ContainerInfo) error{
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No such Network: %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID: fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress: ip,
		Network: network,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network, ep); err != nil{
		return err
	}

	if err = configEndpointIpAddressAndRoute(ep, cinfo); err!=nil{
		return err
	}

	return configPortMapping(ep, cinfo)
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprintf(w, "NAME\tTpRange\tDriver\n")

	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
				nw.Name,
				nw.IPRange.String(),
				nw.Driver,
			)
	}
	if err := w.Flush(); err != nil {
		log.ConsoleLog.Error("Flush error: %v", err)
	}
}

func DeleteNetwork(networkName string) error{
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No Such Network: %s", networkName)
	}

	if err := ipAllocator.Release(nw.IPRange, &nw.IPRange.IP); err != nil{
		return fmt.Errorf("Error remove network driver error: %v", err)
	}

	return nw.remove(defaultNetworkPath)
}

func Init() error{
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	exist, _ := utils.PathExists(defaultNetworkPath)
	if !exist {
		_ = os.MkdirAll(defaultNetworkPath, 0644)
	}

	_ = filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwPath,
		}

		if err := nw.load(nwPath); err != nil {
			log.ConsoleLog.Error("error load network: %s", err)
		}

		networks[nwName] = nw

		return nil
	})

	return nil
}