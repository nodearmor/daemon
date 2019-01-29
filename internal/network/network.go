package network

import (
	"io/ioutil"
	"net"

	"github.com/rs/zerolog"
)

var defaultLogger = zerolog.New(ioutil.Discard)
var log = &defaultLogger

// SetLogger : sets library logger
func SetLogger(l *zerolog.Logger) {
	log = l
}

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

// Network : controls network
type Network interface {
	ID() string
	Start() error
	Stop() error
	Reload() error
	SetConfig(config NetworkConfig) error
	GetPubKey() (string, error)
}

// Manager : controls networks
type Manager interface {
	Type() string
	CreateNetwork(id string) (*Network, error)
	GetNetwork(id string) (*Network, error)
	DeleteNetwork(id string) error
}
