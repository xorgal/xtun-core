package tun

import (
	"log"
	"net"
	"runtime"
	"strconv"

	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/internal"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/netutil"
)

func CreateTunInterface(config config.Config) (iface *water.Interface) {
	c := water.Config{DeviceType: water.TUN}
	c.PlatformSpecificParams = water.PlatformSpecificParams{}
	os := runtime.GOOS
	if os == "windows" {
		c.PlatformSpecificParams.Name = "wintun"
		c.PlatformSpecificParams.Network = []string{config.CIDR}
	}
	if config.DeviceName != "" {
		c.PlatformSpecificParams.Name = config.DeviceName
	}
	iface, err := water.New(c)
	if err != nil {
		internal.Fatalln("failed to create tun interface:", err)
	}
	setRoute(config, iface)
	return iface
}

func setRoute(config config.Config, iface *water.Interface) {
	ip, _, err := net.ParseCIDR(config.CIDR)
	if err != nil {
		internal.Fatalf("failed to parse IPv4 from CIDR: %v", config.CIDR)
	}
	os := runtime.GOOS
	if os == "linux" {
		internal.Exec("/sbin/ip", "link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(config.MTU))
		internal.Exec("/sbin/ip", "addr", "add", config.CIDR, "dev", iface.Name())
		internal.Exec("/sbin/ip", "link", "set", "dev", iface.Name(), "up")
		if !config.ServerMode && config.GlobalMode {
			physicalIface := netutil.GetInterface()
			serverAddrIP := netutil.LookupServerAddrIP(config.ServerAddr)
			if physicalIface != "" && serverAddrIP != nil {
				if config.LocalGateway != "" {
					internal.Exec("/sbin/ip", "route", "add", "0.0.0.0/1", "dev", iface.Name())
					internal.Exec("/sbin/ip", "route", "add", "128.0.0.0/1", "dev", iface.Name())
					if serverAddrIP.To4() != nil {
						internal.Exec("/sbin/ip", "route", "add", serverAddrIP.To4().String()+"/32", "via", config.LocalGateway, "dev", physicalIface)
					}
				}
			}
		}
	} else if os == "darwin" {
		internal.Exec("ifconfig", iface.Name(), "inet", ip.String(), config.ServerIP, "up")
		if !config.ServerMode && config.GlobalMode {
			physicalIface := netutil.GetInterface()
			serverAddrIP := netutil.LookupServerAddrIP(config.ServerAddr)
			if physicalIface != "" && serverAddrIP != nil {
				if config.LocalGateway != "" {
					internal.Exec("route", "add", "default", config.ServerIP)
					internal.Exec("route", "change", "default", config.ServerIP)
					internal.Exec("route", "add", "0.0.0.0/1", "-interface", iface.Name())
					internal.Exec("route", "add", "128.0.0.0/1", "-interface", iface.Name())
					if serverAddrIP.To4() != nil {
						internal.Exec("route", "add", serverAddrIP.To4().String(), config.LocalGateway)
					}
				}
			}
		}
	} else if os == "windows" {
		if !config.ServerMode && config.GlobalMode {
			serverAddrIP := netutil.LookupServerAddrIP(config.ServerAddr)
			if serverAddrIP != nil {
				if config.LocalGateway != "" {
					internal.Exec("cmd", "/C", "route", "delete", "0.0.0.0", "mask", "0.0.0.0")
					internal.Exec("cmd", "/C", "route", "add", "0.0.0.0", "mask", "0.0.0.0", config.ServerIP, "metric", "6")
					if serverAddrIP.To4() != nil {
						internal.Exec("cmd", "/C", "route", "add", serverAddrIP.To4().String()+"/32", config.LocalGateway, "metric", "5")
					}
				}
			}
		}
	} else {
		internal.Fatalf("not support os: %s", os)
	}
	log.Printf("tun device configured %s", iface.Name())
}

// Resets the system routes
func ResetRoute(config config.Config) {
	if config.ServerMode || !config.GlobalMode {
		return
	}
	os := runtime.GOOS
	if os == "darwin" {
		if config.LocalGateway != "" {
			internal.Exec("route", "add", "default", config.LocalGateway)
			internal.Exec("route", "change", "default", config.LocalGateway)
		}
	} else if os == "windows" {
		serverAddrIP := netutil.LookupServerAddrIP(config.ServerAddr)
		if serverAddrIP != nil {
			if config.LocalGateway != "" {
				internal.Exec("cmd", "/C", "route", "delete", "0.0.0.0", "mask", "0.0.0.0")
				internal.Exec("cmd", "/C", "route", "add", "0.0.0.0", "mask", "0.0.0.0", config.LocalGateway, "metric", "6")
			}
		}
	}
}
