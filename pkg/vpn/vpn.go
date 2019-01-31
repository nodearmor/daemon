package vpn

import (
	"net"
)

// NodeConfig : a network node configuration
type NodeConfig struct {
	ID         string
	Name       string
	PrivateIPs []net.IPNet
	PublicIPs  []net.IP
	PubKey     string
}

// RouteConfig : defines a single network static route
type RouteConfig struct {
	Route   net.IPNet
	Gateway net.IP
}

// NetworkConfig : whole network configuration object
type NetworkConfig struct {
	Nodes  []NodeConfig
	SelfID string
	Routes []RouteConfig
}

// VPN : vpn abstraction layer
type VPN interface {
	ID() string
	Start() error
	Stop() error
	Reload() error
	SetConfig(config NetworkConfig) error
	GetPubKey() (string, error)
}

// VPNManager : controls creation of VPNs
type VPNManager interface {
	Type() string
	CreateNetwork(id string) (VPN, error)
	GetNetwork(id string) (VPN, error)
	DeleteNetwork(id string) error
}
