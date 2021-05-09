package commands

import (
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	count int
	// alias for show
	newAccsCmd = &cobra.Command{
		Hidden: false,

		Use:   "new-accs",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   newAccs,
	}
)

// init
func init() {
	newAccsCmd.Flags().IntVarP(&count, "count", "c", 3, "space for data")
	viper.BindPFlag("count", SolanoidCmd.Flags().Lookup("count"))
	newAccsCmd.MarkFlagRequired("count")

	SolanoidCmd.AddCommand(newAccsCmd)
}

func newAccs(ccmd *cobra.Command, args []string) {

	for i := 0; i < count; i++ {
		acc := types.NewAccount()

		fmt.Printf("-------------------- ACCOUNT # %d -------------------------\n", i)
		fmt.Printf("Pubkey: %s\n", acc.PublicKey.ToBase58())
		fmt.Printf("Private key: %s\n\n", base58.Encode(acc.PrivateKey))
	}

}
