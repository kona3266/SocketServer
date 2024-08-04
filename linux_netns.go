package main

import (
	"flag"
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func GetFirstIp(cidr string) (string, string, error) {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	maskLen, _ := net.Mask.Size()
	ip = ip.To4()
	ip[len(ip)-1]++
	return ip.String(), ip.String() + "/" + strconv.Itoa(maskLen), nil
}

func GetPeerName(name string) string {
	return name + "_p"
}

func CreateVethInNewns(origin netns.NsHandle, newns string, name string, ip_cidr string) {

	netns.NewNamed(newns)
	p1 := netlink.NewLinkAttrs()
	p1.Name = name
	peer_name := GetPeerName(name)
	p1Link := &netlink.Veth{
		LinkAttrs: p1,
		PeerName:  peer_name,
	}
	netlink.LinkAdd(p1Link)
	netlink.LinkSetUp(p1Link)

	peer_link, _ := netlink.LinkByName(peer_name)
	netlink.LinkSetUp(peer_link)

	addr, _ := netlink.ParseAddr(ip_cidr)
	netlink.AddrAdd(peer_link, addr)

	addr_list, _ := netlink.AddrList(peer_link, netlink.FAMILY_V4)
	for _, addr := range addr_list {
		fmt.Println(addr.IPNet, addr.Label)
	}
	// set one endpoint in origin ns
	err := netlink.LinkSetNsFd(p1Link, int(origin))
	if err != nil {
		fmt.Println("set ns fd ", err)
	}
	// back to origin namespace
	netns.Set(origin)

}

func main() {
	var (
		network = flag.String("net", "192.168.0.0/24", "network")
		port    = flag.String("port", "p1", "interface name")
	)
	flag.Parse()

	_, ip_cidr, _ := GetFirstIp(*network)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origin, _ := netns.Get()

	fmt.Printf("set up veth pair %s, network %s\n", *port, *network)
	CreateVethInNewns(origin, "test", *port, ip_cidr)
	origin.Close()

	// destroy veth pair and netns
	err := netns.DeleteNamed("test")
	if err != nil {
		fmt.Println("DeleteNamed ", err)
	}

	p1Link, _ := netlink.LinkByName(*port)

	err = netlink.LinkDel(p1Link)
	if err != nil {
		fmt.Println("LinkDel ", err)
	}

	_, err = netns.GetFromName("test")
	if err != nil {
		fmt.Println("GetFromName ", err)
	}

}
