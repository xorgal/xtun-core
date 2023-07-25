package tun

import (
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/internal"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/netutil"
)

func CreateTunInterface(config config.Config) (*water.Interface, error) {
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
		return nil, err
	}
	setRoute(config, iface)
	return iface, err
}

func setRoute(config config.Config, iface *water.Interface) error {
	ip, _, err := net.ParseCIDR(config.CIDR)
	if err != nil {
		return err
	}
	os := runtime.GOOS
	if os == "linux" {
		internal.Exec("/sbin/ip", "link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(config.MTU))
		internal.Exec("/sbin/ip", "addr", "add", config.CIDR, "dev", iface.Name())
		internal.Exec("/sbin/ip", "link", "set", "dev", iface.Name(), "up")
		if !config.ServerMode && config.GlobalMode {
			physicalIface := netutil.GetInterface()
			serverAddrIP, err := netutil.LookupServerAddrIP(config.ServerAddr)
			if err != nil {
				return err
			}
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
			serverAddrIP, err := netutil.LookupServerAddrIP(config.ServerAddr)
			if err != nil {
				return err
			}
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
			serverAddrIP, err := netutil.LookupServerAddrIP(config.ServerAddr)
			if err != nil {
				return err
			}
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
		return fmt.Errorf("not supported os: %s", os)
	}
	return nil
}

// Resets the system routes
func ResetRoute(config config.Config) error {
	if config.ServerMode || !config.GlobalMode {
		return nil
	}
	os := runtime.GOOS
	if os == "darwin" {
		if config.LocalGateway != "" {
			internal.Exec("route", "add", "default", config.LocalGateway)
			internal.Exec("route", "change", "default", config.LocalGateway)
		}
		return nil
	} else if os == "windows" {
		serverAddrIP, err := netutil.LookupServerAddrIP(config.ServerAddr)
		if err != nil {
			return err
		}
		if serverAddrIP != nil {
			if config.LocalGateway != "" {
				internal.Exec("cmd", "/C", "route", "delete", "0.0.0.0", "mask", "0.0.0.0")
				internal.Exec("cmd", "/C", "route", "add", "0.0.0.0", "mask", "0.0.0.0", config.LocalGateway, "metric", "6")
			}
		}
	}
	return nil
}
