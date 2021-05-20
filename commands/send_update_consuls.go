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
	UpdateConsulsPrivateKey string
	GravityProgramID        string
	GravityDataAccount      string
	Round                   uint64
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
	viper.BindPFlag("program", updateConsulsCmd.Flags().Lookup("program"))
	updateConsulsCmd.MarkFlagRequired("program")

	updateConsulsCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", updateConsulsCmd.Flags().Lookup("data-account"))
	updateConsulsCmd.MarkFlagRequired("data-account")

	updateConsulsCmd.Flags().StringVarP(&MultisigDataAccount, "multisig-account", "m", "", "Gravity multisig Account")
	viper.BindPFlag("multisig-account", updateConsulsCmd.Flags().Lookup("multisig-account"))
	updateConsulsCmd.MarkFlagRequired("multisig-account")

	updateConsulsCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", updateConsulsCmd.Flags().Lookup("private-key"))
	updateConsulsCmd.MarkFlagRequired("private-key")

	updateConsulsCmd.Flags().Uint64VarP(&Round, "round", "r", 4, "space for data")
	viper.BindPFlag("round", updateConsulsCmd.Flags().Lookup("round"))
	updateConsulsCmd.MarkFlagRequired("round")

	SolanoidCmd.AddCommand(updateConsulsCmd)
}

func NewUpdateConsulsInstruction(fromAccount, programData, targetProgramID, multisigId common.PublicKey, Bft uint8, Round uint64, Consuls [5][32]byte) types.Instruction {
	meta := []types.AccountMeta{
		{PubKey: fromAccount, IsSigner: true, IsWritable: true},
		{PubKey: programData, IsSigner: false, IsWritable: true},
		{PubKey: multisigId, IsSigner: false, IsWritable: true},
	}
	consuls := []byte{}
	for i := 0; i < int(Bft); i++ {
		//acc := types.NewAccountxx()
		k := common.PublicKeyFromBytes(Consuls[i][:])
		meta = append(meta, types.AccountMeta{PubKey: k, IsSigner: true, IsWritable: false})
		consuls = append(consuls, Consuls[i][:]...)
	}

	data, err := common.SerializeData(struct {
		Instruction uint8
		Bft         uint8
		Consuls     []byte
		Round       uint64
	}{
		Instruction: 1,
		Bft:         1,
		Round:       Round,
		Consuls:     consuls,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	fmt.Printf("%s\n", hex.EncodeToString(data))
	fmt.Println("------- END RAW INSTRUCTION DATA ---------")
	return types.Instruction{
		Accounts:  meta,
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

	// pks := []string{
	// 	"4X77h6B7dAnz5AyJrdJMP5FBuefb7RgPS5K51xSxjkHeYjn7BdfNGLySrFeyHrf8Lzrwm5479a53Ka4bcYTTdrCB",
	// 	//"44yNH2ub6s3xQwi44zHFsDg3VUTP3ZmmsaVxTXfSgJV9BuFFJ8ZXAaNcSvxysxDFDbhAASvXMSZi4gnxskSsH4Aw",
	// }
	consuls := []types.Account{account}
	// for _, cpk := range pks {
	// 	bpk, _ := base58.Decode(cpk)
	// 	consuls = append(consuls, types.AccountFromPrivateKeyBytes(bpk))
	// }
	// consuls = append(consuls, types.NewAccount()) //fake consul

	consulsAddrs := [5][32]byte{}
	for i, v := range consuls {
		copy(consulsAddrs[i][:], v.PublicKey.Bytes())
	}

	program := common.PublicKeyFromString(GravityProgramID)
	dataAcc := common.PublicKeyFromString(GravityDataAccount)
	multisigAcc := common.PublicKeyFromString(MultisigDataAccount)
	c := client.NewClient(client.TestnetRPCEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			NewUpdateConsulsInstruction(
				account.PublicKey, dataAcc, program, multisigAcc, 1, Round, consulsAddrs,
			),
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
	}

	signs := make(map[common.PublicKey]types.Signature)
	signs[account.PublicKey] = ed25519.Sign(account.PrivateKey, serializedMessage)
	for _, c := range consuls {
		signs[c.PublicKey] = ed25519.Sign(c.PrivateKey, serializedMessage)
	}

	tx, err := types.CreateTransaction(message, signs)
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
