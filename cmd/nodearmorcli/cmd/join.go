package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(joinCmd)
}

var joinCmd = &cobra.Command{
	Use:   "join <networkId>",
	Short: "Join a network",
	Long:  `Sends a request to the controller to join a network. Request has to be approved by the controller.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rpcClient()
		if err != nil {
			return err
		}

		var reply bool
		err = client.Call("nodearmord.join", args[0], &reply)
		if err != nil {
			return fmt.Errorf("Error joining network: %s", err)
		}

		return nil
	},
}
