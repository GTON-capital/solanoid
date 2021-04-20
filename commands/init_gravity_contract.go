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
	initGravityContractCmd = &cobra.Command{
		Hidden: false,

		Use:   "init-gravity",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   initGravity,
	}
)

// init
func init() {
	initGravityContractCmd.Flags().StringVarP(&GravityProgramID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", initGravityContractCmd.Flags().Lookup("program"))
	initGravityContractCmd.MarkFlagRequired("program")

	initGravityContractCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", initGravityContractCmd.Flags().Lookup("data-account"))
	initGravityContractCmd.MarkFlagRequired("data-account")

	initGravityContractCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", initGravityContractCmd.Flags().Lookup("private-key"))
	initGravityContractCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(initGravityContractCmd)
}

func NewInitGravityContractInstruction(fromAccount, programData, targetProgramID common.PublicKey, Bft uint8, Round uint64, Consuls [5][32]byte) types.Instruction {
	consuls := []byte{}
	for i := 0; i < 3; i++ {
		acc := types.NewAccount()
		consuls = append(consuls, acc.PublicKey.Bytes()...)
	}
	data, err := common.SerializeData(struct {
		Instruction uint8
		Bft         uint8
		Consuls     []byte
		Round       uint64
	}{
		Instruction: 0,
		Bft:         3,
		Consuls:     consuls,
		Round:       0,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	fmt.Printf("%s\n", hex.EncodeToString(data))
	fmt.Println("------- END RAW INSTRUCTION DATA ---------")
	return types.Instruction{
		Accounts: []types.AccountMeta{
			{PubKey: fromAccount, IsSigner: true, IsWritable: true},
			{PubKey: programData, IsSigner: false, IsWritable: true},
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}

func initGravity(ccmd *cobra.Command, args []string) {
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
			NewInitGravityContractInstruction(
				account.PublicKey, dataAcc, program, 3, 1, [5][32]byte{},
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
