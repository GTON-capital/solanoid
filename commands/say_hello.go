package commands

import (
	"crypto/ed25519"
	"log"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	helloPrivateKey string
	programID       string
	helloAccount    string
	// alias for show
	sayHelloCmd = &cobra.Command{
		Hidden: false,

		Use:   "hello",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   hello,
	}
)

// init
func init() {
	sayHelloCmd.Flags().StringVarP(&programID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", SolanoidCmd.Flags().Lookup("program"))
	sayHelloCmd.MarkFlagRequired("program")

	sayHelloCmd.Flags().StringVarP(&helloAccount, "to", "t", "", "Say hello to [account]")
	viper.BindPFlag("to", SolanoidCmd.Flags().Lookup("to"))
	sayHelloCmd.MarkFlagRequired("to")

	sayHelloCmd.Flags().StringVarP(&helloPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", SolanoidCmd.Flags().Lookup("private-key"))
	sayHelloCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(sayHelloCmd)
}

func hello(ccmd *cobra.Command, args []string) {
	pk, err := base58.Decode(helloPrivateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	pid := common.PublicKeyFromString(programID)
	to := common.PublicKeyFromString(helloAccount)

	c := client.NewClient(client.TestnetRPCEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			{
				ProgramID: pid,
				Accounts: []types.AccountMeta{
					{
						PubKey:     to,
						IsSigner:   false,
						IsWritable: true,
					},
				},
				Data: []byte{13},
			},
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
	}

	txSig, err := c.SendRawTransaction(rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Printf("upload program txHash : %s", txSig)

}
