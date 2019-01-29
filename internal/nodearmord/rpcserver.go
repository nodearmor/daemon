package nodearmord

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	RPCHost = "localhost"
	RPCPort = 39825
)

type NetworkRPC struct{}

func (t *NetworkRPC) Join(id string, reply *bool) error {
	log.Printf("RPC: Joining network %s", id)

	/*joinNetworkRequest := JoinNetworkRequest{
		NetworkId: id,
	}
	SendMessage("joinNetwork", joinNetworkRequest)*/

	*reply = true
	return nil
}

/*func (t *NetworkRPC) Leave(id string, reply *string) error {
	log.Printf("RPC: Leaving network %s", id)

	leaveNetworkRequest := LeaveNetworkRequest{
		NetworkId: id,
	}
	SendMessage("leaveNetwork", leaveNetworkRequest)

	*reply = "ok"
	return nil
}

func (t *NetworkRPC) List(tmp bool, reply *string) error {
	log.Printf("RPC: List networks")

	for _, network := range config.Networks {
		*reply += fmt.Sprintf("Network %s\n", network.Id)
		*reply += fmt.Sprintf("  Name: %s\n", network.Name)
		*reply += fmt.Sprintf("  Subnet: %s\n", network.Subnet)
		*reply += fmt.Sprintf("  Nodes:\n")
		for _, node := range network.Nodes {
			*reply += fmt.Sprintf("    Node %s\n", node.Id)
			*reply += fmt.Sprintf("      Name: %s\n", node.Name)
			*reply += fmt.Sprintf("    	 PrivateIP: %s\n", node.PrivateIP)
			*reply += fmt.Sprintf("    	 PublicIP: %s\n", node.PublicIP)
		}
	}

	return nil
}*/

func StartRPCServer(wg *sync.WaitGroup, stop signalCh) {
	networkRPC := new(NetworkRPC)

	// Publish the receivers methods
	err := rpc.Register(networkRPC)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register NetworkRPC service")
	}

	// Register a HTTP handler
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", RPCPort))
	if err != nil {
		log.Fatal().Err(err).Msg("RPC Listen error")
	}

	log.Info().Int("port", RPCPort).Msg("Serving RPC server on port")

	go func() {
		<-stop
		listener.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Start accept incoming HTTP connections
		err = http.Serve(listener, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Error serving RPC")
		}
	}()
}
