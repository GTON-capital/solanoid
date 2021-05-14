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
	NebulaDataAccount     string
	initNebulaContractCmd = &cobra.Command{
		Hidden: false,

		Use:   "init-nebula",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   initNebula,
	}
)

// init
func init() {
	initNebulaContractCmd.Flags().StringVarP(&GravityProgramID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", initNebulaContractCmd.Flags().Lookup("program"))
	initNebulaContractCmd.MarkFlagRequired("program")

	initNebulaContractCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", initNebulaContractCmd.Flags().Lookup("data-account"))
	initNebulaContractCmd.MarkFlagRequired("data-account")

	initNebulaContractCmd.Flags().StringVarP(&NebulaDataAccount, "nebula-account", "n", "", "Gravity multisig Account")
	viper.BindPFlag("nebula-account", initNebulaContractCmd.Flags().Lookup("nebula-account"))
	initNebulaContractCmd.MarkFlagRequired("nebula-account")

	initNebulaContractCmd.Flags().StringVarP(&MultisigDataAccount, "multisig-account", "m", "", "Gravity multisig Account")
	viper.BindPFlag("multisig-account", initNebulaContractCmd.Flags().Lookup("multisig-account"))
	initNebulaContractCmd.MarkFlagRequired("multisig-account")

	initNebulaContractCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", initNebulaContractCmd.Flags().Lookup("private-key"))
	initNebulaContractCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(initNebulaContractCmd)
}

func NewInitNebulaContractInstruction(fromAccount, programData, nebulaData, multisigData, targetProgramID common.PublicKey, Bft uint8, Consuls [5][32]byte) types.Instruction {
	consuls := []byte{}
	consuls = append(consuls, fromAccount.Bytes()...)
	// for i := 0; i < 2; i++ {
	// 	acc := types.NewAccount()
	// 	zap.L().Sugar().Infof("consul %d pk %s", i, base58.Encode(acc.PrivateKey))
	// 	consuls = append(consuls, acc.PublicKey.Bytes()...)
	// }
	/*
			InitContract {
		        nebula_data_type: DataType,
		        gravity_contract_program_id: Pubkey,
		        oracles_bft: u8,
		        initial_oracles: Vec<Pubkey>,
		    },
	*/
	fmt.Printf("Nebula programId: %s\n", targetProgramID.ToBase58())
	data, err := common.SerializeData(struct {
		Instruction              uint8
		Bft                      uint8
		NebulaDataType           uint8
		GravityContractProgramID common.PublicKey
		Consuls                  []byte
	}{
		Instruction:              0,
		Bft:                      1,
		NebulaDataType:           2,
		GravityContractProgramID: programData,
		Consuls:                  consuls,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	fmt.Printf("%s\n", hex.EncodeToString(data))
	fmt.Println("------- END RAW INSTRUCTION DATA ---------")
	return types.Instruction{
		Accounts: []types.AccountMeta{
			{PubKey: fromAccount, IsSigner: true, IsWritable: false},
			//{PubKey: programData, IsSigner: false, IsWritable: true},
			{PubKey: nebulaData, IsSigner: false, IsWritable: true},
			{PubKey: multisigData, IsSigner: false, IsWritable: true},
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}

func initNebula(ccmd *cobra.Command, args []string) {
	pk, err := base58.Decode(UpdateConsulsPrivateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	program := common.PublicKeyFromString(GravityProgramID)
	dataAcc := common.PublicKeyFromString(GravityDataAccount)
	nebulaAcc := common.PublicKeyFromString(NebulaDataAccount)
	multisigAcc := common.PublicKeyFromString(MultisigDataAccount)
	c := client.NewClient(client.DevnetRPCEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			NewInitNebulaContractInstruction(
				account.PublicKey, dataAcc, nebulaAcc, multisigAcc, program, 1, [5][32]byte{},
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
