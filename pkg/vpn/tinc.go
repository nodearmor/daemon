package vpn

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
)

const (
	configPath            = "/etc/tinc/"
	networkConfigFile     = "tinc.conf"
	networkUpScriptFile   = "tinc-up"
	networkDownScriptFile = "tinc-down"
	serviceFilePath       = "/etc/systemd/system/"
	servicePrefix         = "tincd_"
	keySize               = 2048 // tinc RSA key size
	privKeyFile           = "rsa_key.priv"
	pubKeyFile            = "rsa_key.pub"
)

// TincVPN : TINC network object that controlls the tinc daemon
type TincVPN struct {
	id string
}

// ID : returns TINC network id
func (n *TincVPN) ID() string {
	return n.id
}

// Start : creates systemd service, enables it, and starts daemon
func (n *TincVPN) Start() error {
	err := n.serviceCreate()
	if err != nil {
		return err
	}

	err = n.serviceStart()
	if err != nil {
		return err
	}

	err = n.serviceEnable()
	if err != nil {
		return err
	}

	return nil
}

// Stop : stops systemd service and removes from startup
func (n *TincVPN) Stop() error {
	err := n.serviceStop()
	if err != nil {
		return err
	}

	err = n.serviceDisable()
	if err != nil {
		return err
	}

	return nil
}

// Reload : reloads TINC network configuration
func (n *TincVPN) Reload() error {
	err := n.serviceReload()
	if err != nil {
		return err
	}

	return nil
}

// SetConfig : sets tinc network configuration
func (n *TincVPN) SetConfig(config NetworkConfig) error {
	// Remove old hosts if exist
	err := os.RemoveAll(n.hostConfigPath())
	if err != nil {
		return fmt.Errorf("error removing old hosts: %s", err)
	}

	// Create new hosts folder
	os.Mkdir(n.hostConfigPath(), 0666)

	for _, node := range config.Nodes {
		// Write own config
		if node.ID == config.SelfID {
			// build list of nodes to connect to
			var connectIds []string
			for _, conNode := range config.Nodes {
				if conNode.ID != config.SelfID {
					connectIds = append(connectIds, conNode.ID)
				}
			}

			n.writeNetworkConfg(node.ID, connectIds)
			n.writeNetworkUpScript(node.PrivateIPs, config.Routes)
			n.writeNetworkUpScript(node.PrivateIPs, config.Routes)
		}

		// Write general host config
		n.writeHostConfig(node.ID, node.PrivateIPs, node.PublicIPs, node.PubKey)
	}

	return nil
}

func (n *TincVPN) writeNetworkConfg(selfID string, connectIds []string) error {
	filePath := path.Join(n.networkConfigPath(), networkConfigFile)

	file, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		return fmt.Errorf("error writing network configuration file %s: %s", filePath, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintf(w, "Name = %s\n", selfID)

	// Write nodes to connect to
	for _, ip := range connectIds {
		fmt.Fprintf(w, "ConnectTo = %s\n", ip)
	}

	w.Flush()

	return nil
}

func (n *TincVPN) writeNetworkUpScript(ips []net.IPNet, routes []RouteConfig) error {
	filePath := path.Join(n.networkConfigPath(), networkUpScriptFile)

	file, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0777,
	)
	if err != nil {
		return fmt.Errorf("error writing network up script %s: %s", filePath, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintf(w, "#!/bin/sh\n")

	// Enable interface
	fmt.Fprintf(w, "ip link set $INTERFACE up\n")

	// Add ip addresses
	for _, ip := range ips {
		fmt.Fprintf(w, "ip addr add %s dev $INTERFACE\n", ip.String())
	}

	// Add routes
	for _, route := range routes {
		if route.Gateway.IsUnspecified() {
			fmt.Fprintf(w, "ip route add %s dev $INTERFACE\n", route.Route.String())
		} else {
			fmt.Fprintf(w, "ip route add %s via %s\n", route.Route.String(), route.Gateway.String())
		}
	}

	w.Flush()

	return nil
}

func (n *TincVPN) writeNetworkDownScript(ips []net.IPNet, routes []RouteConfig) error {
	filePath := path.Join(n.networkConfigPath(), networkDownScriptFile)

	file, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0777,
	)
	if err != nil {
		return fmt.Errorf("error writing network down script %s: %s", filePath, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintf(w, "#!/bin/sh\n")

	// Removes routes
	for _, route := range routes {
		if route.Gateway.IsUnspecified() {
			fmt.Fprintf(w, "ip route del %s dev $INTERFACE\n", route.Route.String())
		} else {
			fmt.Fprintf(w, "ip route del %s via %s\n", route.Route.String(), route.Gateway.String())
		}
	}

	// Add ip addresses
	for _, ip := range ips {
		fmt.Fprintf(w, "ip addr del %s dev $INTERFACE\n", ip.String())
	}

	// Disable interface
	fmt.Fprintf(w, "ip link set $INTERFACE down\n")

	w.Flush()

	return nil
}

func (n *TincVPN) writeHostConfig(id string, privateIPs []net.IPNet, publicIPs []net.IP, pubkey string) error {
	filePath := path.Join(n.hostConfigPath(), id)

	// Open a new file for writing only
	file, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		return fmt.Errorf("error writing host config file %s: %s", filePath, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	// Add public ip addresses
	for _, publicIP := range publicIPs {
		fmt.Fprintf(w, "Address = %s\n", publicIP)
	}

	// Add private ip addresses
	for _, privateIP := range privateIPs {
		fmt.Fprintf(w, "Subnet = %s\n", privateIP.String())
	}

	// Write pubkey
	fmt.Fprint(w, string(pubkey))

	w.Flush()

	return nil
}

func (n *TincVPN) GetPubKey() (string, error) {
	var pubKeyPath = path.Join(n.networkConfigPath(), pubKeyFile)

	// Check if key exists
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		// Generate new keys
		err = n.generateKeys()
		if err != nil {
			return "", fmt.Errorf("Error generating keys: %s", err)
		}
	}

	// Read file
	buf, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("Error reading key file: %s", err)
	}

	return string(buf), nil
}

func (n *TincVPN) generateKeys() error {
	var privKeyPath = path.Join(n.networkConfigPath(), privKeyFile)
	var pubKeyPath = path.Join(n.networkConfigPath(), pubKeyFile)

	// Remove old keys if present
	err := os.Remove(privKeyPath)
	if err != nil {
		return fmt.Errorf("Error removing old private key file: %s", err)
	}
	err = os.Remove(pubKeyPath)
	if err != nil {
		return fmt.Errorf("Error removing old public key file: %s", err)
	}

	// Private Key generation
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("Error generating private key: %s", err)
	}

	// Validate Private Key
	err = key.Validate()
	if err != nil {
		return fmt.Errorf("Private key validation failed: %s", err)
	}

	// Save private key in PEM file
	privFile, err := os.Create(privKeyPath)
	if err != nil {
		return fmt.Errorf("Error creating private key file: %s", err)
	}
	defer privFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(privFile, privateKey)
	if err != nil {
		return fmt.Errorf("Error encoding private key file: %s", err)
	}

	// Save public key in PEM file
	asn1Bytes, err := asn1.Marshal(key.PublicKey)
	if err != nil {
		return fmt.Errorf("Error marshaling public key: %s", err)
	}

	var publicKey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pubFile, err := os.Create(pubKeyPath)
	if err != nil {
		return fmt.Errorf("Error creating public key file: %s", err)
	}
	defer pubFile.Close()

	err = pem.Encode(pubFile, publicKey)
	if err != nil {
		return fmt.Errorf("Error encoding public key: %s", err)
	}

	return nil
}

func (n *TincVPN) serviceName() string {
	return fmt.Sprintf("%s%s", servicePrefix, n.id)
}

func (n *TincVPN) serviceFile() string {
	return path.Join(serviceFilePath, fmt.Sprintf("%s.service", n.serviceName()))
}

func (n *TincVPN) networkConfigPath() string {
	return path.Join(configPath, n.id)
}

func (n *TincVPN) hostConfigPath() string {
	return path.Join(n.networkConfigPath(), "hosts")
}

func (n *TincVPN) serviceCreate() error {
	// Open a new file for writing only
	file, err := os.OpenFile(
		n.serviceFile(),
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		return fmt.Errorf("Error creating service file %s: %s", n.serviceFile(), err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintf(w, "[Unit]\n")
	fmt.Fprintf(w, "Description=Tinc Daemon %s\n", n.id)
	fmt.Fprintf(w, "After=network.target\n")
	fmt.Fprintf(w, "Requires=network.target\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "[Service]\n")
	fmt.Fprintf(w, "Type=simple\n")
	fmt.Fprintf(w, "ExecStart=/usr/sbin/tincd -D -n %s -L -R\n", n.id)
	fmt.Fprintf(w, "Restart=always\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "[Install]\n")
	fmt.Fprintf(w, "WantedBy=multi-user.target\n")

	w.Flush()

	return nil
}

func (n *TincVPN) serviceRemove() error {
	err := os.Remove(n.serviceFile())
	if err != nil {
		return fmt.Errorf("Error removing service file %s: %s", n.serviceFile(), err)
	}

	return nil
}

func (n *TincVPN) serviceStart() error {
	cmd := exec.Command("systemctl", "start", n.serviceName())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error starting service: %s", err)
	}

	return nil
}

func (n *TincVPN) serviceStop() error {
	cmd := exec.Command("systemctl", "stop", n.serviceName())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error stopping service: %s", err)
	}

	return nil
}

func (n *TincVPN) serviceReload() error {
	cmd := exec.Command("systemctl", "kill", "-s", "HUP", n.serviceName())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error reloading service: %s", err)
	}

	return nil
}

func (n *TincVPN) serviceEnable() error {
	cmd := exec.Command("systemctl", "enable", n.serviceName())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error enabling service: %s", err)
	}

	return nil
}

func (n *TincVPN) serviceDisable() error {
	cmd := exec.Command("systemctl", "disable", n.serviceName())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error disabling service: %s", err)
	}

	return nil
}

// TincVPNManager : manages TINC networks
type TincVPNManager struct{}

// Type : returns network type "tinc"
func (d *TincVPNManager) Type() string {
	return "tinc"
}

// CreateNetwork : creates configuration folder for TINC network and returns network pointer
func (d *TincVPNManager) CreateNetwork(id string) (VPN, error) {
	networkConfigPath := path.Join(configPath, id)

	if _, err := os.Stat(networkConfigPath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("Network %s already exists", id)
	}

	// Create directory
	os.MkdirAll(networkConfigPath, 0755)

	return &TincVPN{
		id: id,
	}, nil
}

// GetNetwork : finds TINC network pointer based on network id
func (d *TincVPNManager) GetNetwork(id string) (VPN, error) {
	networkConfigPath := path.Join(configPath, id)

	if _, err := os.Stat(networkConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Network %s does not exist", id)
	}

	return &TincVPN{
		id: id,
	}, nil
}

// DeleteNetwork : stops and removes TINC network
func (d *TincVPNManager) DeleteNetwork(id string) error {
	network, err := d.GetNetwork(id)
	if err != nil {
		return fmt.Errorf("Error deleting network %s: %s", id, err)
	}

	// Stop network
	err = network.Stop()
	if err != nil {
		return fmt.Errorf("Error stopping network %s: %s", id, err)
	}

	// Remove files
	networkConfigPath := path.Join(configPath, id)
	err = os.RemoveAll(networkConfigPath)
	if err != nil {
		return fmt.Errorf("Error removing network configuration %s: %s", id, err)
	}

	return nil
}
