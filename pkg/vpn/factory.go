package vpn

import (
	"fmt"
)

// GetVPNManager : returns VPNManager based on string kind
func GetVPNManager(kind string) (VPNManager, error) {
	switch kind {
	case "tinc":
		return &TincVPNManager{}, nil
	default:
		return nil, fmt.Errorf("Invalid VPN kind: %s", kind)
	}
}
