package cmd

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nodearmorcli",
	Short: "Command line interface for nodearmor daemon",
}

var config struct {
	RPCHost string
}

func init() {
	rootCmd.PersistentFlags().StringVar(&config.RPCHost, "rpchost", "localhost:39825", "Host of nodearmord RPC server")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rpcClient() (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", config.RPCHost)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to RPC server '%s': %s", config.RPCHost, err)
	}

	return client, nil
}
