package netutil

import (
	"log"
	"net"
	"strings"

	"github.com/net-byte/go-gateway"
)

// Returns the name of interface
func GetInterface() (name string) {
	ifaces := getAllInterfaces()
	if len(ifaces) == 0 {
		return ""
	}
	netAddrs, _ := ifaces[0].Addrs()
	for _, addr := range netAddrs {
		ip, ok := addr.(*net.IPNet)
		if ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
			name = ifaces[0].Name
			break
		}
	}
	return name
}

// Returns all interfaces
func getAllInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return nil
	}

	var outInterfaces []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp == 1 && isPhysicalInterface(iface.Name) {
			netAddrs, _ := iface.Addrs()
			if len(netAddrs) > 0 {
				outInterfaces = append(outInterfaces, iface)
			}
		}
	}
	return outInterfaces
}

// Returns true if the interface is physical
func isPhysicalInterface(addr string) bool {
	prefixArray := []string{"ens", "enp", "enx", "eno", "eth", "en0", "wlan", "wlp", "wlo", "wlx", "wifi0", "lan0"}
	for _, pref := range prefixArray {
		if strings.HasPrefix(strings.ToLower(addr), pref) {
			return true
		}
	}
	return false
}

// Lookup IP address of the given hostname
func LookupIP(domain string) (net.IP, error) {
	ips, err := net.LookupIP(domain)
	if err != nil || len(ips) == 0 {
		return nil, err
	}
	return ips[0], nil
}

// LookupServerAddrIP returns the IP of server address
func LookupServerAddrIP(serverAddr string) (net.IP, error) {
	host, _, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, err
	}
	ip, err := LookupIP(host)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

// DiscoverGateway returns the local gateway IP address
func DiscoverGateway(ipv4 bool) (net.IP, error) {
	var ip net.IP
	var err error
	if ipv4 {
		ip, err = gateway.DiscoverGatewayIPv4()
	} else {
		ip, err = gateway.DiscoverGatewayIPv6()
	}
	if err != nil {
		return nil, err
	}
	return ip, err
}

// IsIPv4 returns true if the packet is IPv4s
func IsIPv4(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 4
}

// IsIPv6 returns true if the packet is IPv6s
func IsIPv6(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 6
}

// GetIPv4Src returns the IPv4 source address of the packet
func GetIPv4Src(packet []byte) net.IP {
	return net.IPv4(packet[12], packet[13], packet[14], packet[15])
}

// GEtIPv4Dst returns the IPv4 destination address of the packet
func GetIPv4Dst(packet []byte) net.IP {
	return net.IPv4(packet[16], packet[17], packet[18], packet[19])
}

// GetIPv6Src returns the IPv6 source address of the packet
func GetIPv6Src(packet []byte) net.IP {
	return net.IP(packet[8:24])
}

// GetIPv6Dst returns the IPv6 destination address of the packet
func GetIPv6Dst(packet []byte) net.IP {
	return net.IP(packet[24:40])
}

// GetSrcKey returns the source key of the packet
func GetSrcKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Src(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Src(packet).To16().String()
	}
	return key
}

// GetdstKey returns the destination key of the packets
func GetDstKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Dst(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Dst(packet).To16().String()
	}
	return key
}
