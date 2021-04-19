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
	UpdateConsulsPrivateKey string
	GravityProgramID        string
	GravityDataAccount      string
	Round                   uint8
	// alias for show
	updateConsulsCmd = &cobra.Command{
		Hidden: false,

		Use:   "update-consuls",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   updateConsuls,
	}
)

// init
func init() {
	updateConsulsCmd.Flags().StringVarP(&GravityProgramID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", SolanoidCmd.Flags().Lookup("program"))
	updateConsulsCmd.MarkFlagRequired("program")

	updateConsulsCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", SolanoidCmd.Flags().Lookup("data-account"))
	updateConsulsCmd.MarkFlagRequired("data-account")

	updateConsulsCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", SolanoidCmd.Flags().Lookup("private-key"))
	updateConsulsCmd.MarkFlagRequired("private-key")

	updateConsulsCmd.Flags().Uint8VarP(&Round, "round", "r", 4, "space for data")
	viper.BindPFlag("round", SolanoidCmd.Flags().Lookup("round"))
	updateConsulsCmd.MarkFlagRequired("round")

	SolanoidCmd.AddCommand(updateConsulsCmd)
}

func NewUpdateConsulsInstruction(fromAccount, programData, targetProgramID common.PublicKey, Bft uint8, Round uint64, Consuls [5][32]byte) types.Instruction {

	data, err := common.SerializeData(struct {
		Instruction uint8
		Bft         uint8
		Consuls     [96]byte
		Round       uint64
	}{
		Instruction: 0,
		Bft:         3,
		Round:       Round,
		Consuls:     [96]byte{},
	})
	if err != nil {
		panic(err)
	}

	return types.Instruction{
		Accounts: []types.AccountMeta{
			{PubKey: fromAccount, IsSigner: true, IsWritable: true},
			{PubKey: programData, IsSigner: false, IsWritable: true},
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}

func updateConsuls(ccmd *cobra.Command, args []string) {
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
			NewUpdateConsulsInstruction(
				account.PublicKey, dataAcc, program, uint8(Round), 1, [5][32]byte{},
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

	txSig, err := c.SendRawTransaction(rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Println("txHash:", txSig)

}
