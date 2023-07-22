package config

type Config struct {
	DeviceName         string
	DeviceId           string
	ServerAddr         string
	ServerIP           string
	LocalGateway       string
	CIDR               string
	Key                string
	Protocol           string
	BufferSize         int
	MTU                int
	GlobalMode         bool
	ServerMode         bool
	GUIMode            bool
	InsecureSkipVerify bool
	Compress           bool
}

var AppConfig Config
