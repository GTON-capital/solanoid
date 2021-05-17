package commands

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
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
	// UpdateConsulsPrivateKey string
	// GravityProgramID        string
	// GravityDataAccount      string
	// Round                   uint64
	// alias for show
	MessageToCall string
	callMemoCmd   = &cobra.Command{
		Hidden: false,

		Use:   "call-memo",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   callMemo,
	}
)

// init
func init() {
	callMemoCmd.Flags().StringVarP(&GravityProgramID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", callMemoCmd.Flags().Lookup("program"))
	callMemoCmd.MarkFlagRequired("program")

	callMemoCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", callMemoCmd.Flags().Lookup("private-key"))
	callMemoCmd.MarkFlagRequired("private-key")

	callMemoCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", callMemoCmd.Flags().Lookup("data-account"))
	callMemoCmd.MarkFlagRequired("data-account")

	callMemoCmd.Flags().StringVarP(&MessageToCall, "message", "m", "Kavabunga", "Message")
	viper.BindPFlag("message", callMemoCmd.Flags().Lookup("message"))
	//callMemoCmd.MarkFlagRequired("message")

	SolanoidCmd.AddCommand(callMemoCmd)
}

func NewCallMemoInstruction(fromAccount, targetProgramID common.PublicKey, msg string) types.Instruction {

	return types.Instruction{
		Accounts: []types.AccountMeta{
			{PubKey: fromAccount, IsSigner: false, IsWritable: true},
		},
		ProgramID: targetProgramID,
		Data:      []byte(msg),
	}
}

func callMemo(ccmd *cobra.Command, args []string) {
	pk, err := base58.Decode(UpdateConsulsPrivateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	program := common.PublicKeyFromString(GravityProgramID)
	dataAcc := common.PublicKeyFromString(GravityDataAccount)

	c := client.NewClient(client.TestnetRPCEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			NewCallMemoInstruction(
				dataAcc, program, MessageToCall,
			),
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
	fmt.Println("------ RAW TRANSACTION ------------------------")
	fmt.Printf("%s\n", hex.EncodeToString(rawTx))
	fmt.Println("------ END RAW TRANSACTION ------------------------")

	fmt.Println("------ RAW MESSAGE ------------------------")
	fmt.Printf("%s\n", hex.EncodeToString(serializedMessage))
	fmt.Println("------ END RAW MESSAGE ------------------------")

	txSig, err := c.SendRawTransaction(rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Println("txHash:", txSig)
}
