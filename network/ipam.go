package network

import (
	"bucket/log"
	"bucket/utils"
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/bucket/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets *map[string]string
}

var ipAllocator = &IPAM {
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (i *IPAM) load() error {
	exist, err := utils.PathExists(i.SubnetAllocatorPath)
	if err != nil{
		return err
	}
	if !exist{
		return nil
	}

	subnetConfigFile, err := os.Open(i.SubnetAllocatorPath)
	if err != nil{
		return err
	}
	defer func() {
		_ = subnetConfigFile.Close()
	}()

	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], i.Subnets)
	if err != nil {
		return err
	}

	return nil
}

func (i *IPAM) dump() error{
	ipamConfigFileDir, _ := path.Split(i.SubnetAllocatorPath)
	exist, _ := utils.PathExists(ipamConfigFileDir)
	if !exist {
		_ = os.MkdirAll(ipamConfigFileDir, 0644)
	}

	subnetConfigFile, err  := os.OpenFile(i.SubnetAllocatorPath, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	ipamConfigJSON, err := json.Marshal(i.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJSON)
	if err != nil {
		return err
	}

	return nil
}

func (i *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	i.Subnets = &map[string]string{}
	err = i.load()
	if err != nil {
		log.ConsoleLog.Error("error load allocation info, %v", err)
	}
	one, size := subnet.Mask.Size()

	if _, exist := (*i.Subnets)[subnet.String()]; !exist{
		(*i.Subnets)[subnet.String()] = strings.Repeat("0", 1 << uint8(size-one))
	}

	for c := range (*i.Subnets)[subnet.String()]{
		if (*i.Subnets)[subnet.String()][c] == '0'{
			ipalloc := []byte((*i.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*i.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t --{
				[]byte(ip)[4-t] += uint8((c >> ((t-1) * 8)))
			}
			ip[3] += 1
			break
		}
	}
	err = i.dump()
	return
}

func (i *IPAM)Release(subnet *net.IPNet, ipaddr *net.IP) error {
	i.Subnets = &map[string]string{}
	err := i.load()
	if err != nil {
		log.ConsoleLog.Error("Error load allocation info, %v", err)
	}
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1{
		c += int(releaseIP[t-1] - subnet.IP[t - 1]) << ((4-t) * 8)
	}
	ipalloc := []byte((*i.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*i.Subnets)[subnet.String()] = string(ipalloc)

	return i.dump()
}