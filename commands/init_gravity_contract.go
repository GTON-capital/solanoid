package commands

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"solanoid/commands/executor"
	"solanoid/models"

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
	MultisigDataAccount    string
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

	initGravityContractCmd.Flags().StringVarP(&MultisigDataAccount, "multisig-account", "m", "", "Gravity multisig Account")
	viper.BindPFlag("multisig-account", initGravityContractCmd.Flags().Lookup("multisig-account"))
	initGravityContractCmd.MarkFlagRequired("multisig-account")

	initGravityContractCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", initGravityContractCmd.Flags().Lookup("private-key"))
	initGravityContractCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(initGravityContractCmd)
}

func NewInitGravityContractInstruction(fromAccount, programData, multisigData, targetProgramID common.PublicKey, bft uint8, round uint64, consuls []byte) types.Instruction {
	data, err := common.SerializeData(executor.InitGravityContractInstruction{
		Instruction: 0,
		Bft:         bft,
		InitRound:   round,
		Consuls:     consuls[:],
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
			{PubKey: programData, IsSigner: false, IsWritable: true},
			{PubKey: multisigData, IsSigner: false, IsWritable: true},
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}

func InitGravity(privateKey, programID, stateID, multisigID, clientEndpoint string, consuls []byte) (*models.CommandResponse, error) {
	// pk, err := base58.Decode(UpdateConsulsPrivateKey)
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}

	account := types.AccountFromPrivateKeyBytes(pk)

	// program := common.PublicKeyFromString(GravityProgramID)
	// dataAcc := common.PublicKeyFromString(GravityDataAccount)
	// multisigAcc := common.PublicKeyFromString(MultisigDataAccount)
	program := common.PublicKeyFromString(programID)
	dataAcc := common.PublicKeyFromString(stateID)
	multisigAcc := common.PublicKeyFromString(multisigID)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		fmt.Printf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			NewInitGravityContractInstruction(
				account.PublicKey, dataAcc, multisigAcc, program, 3, 1, consuls,
			),
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		fmt.Printf("serialize message error, err: %v\n", err)
		return nil, err
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	})

	if err != nil {
		fmt.Printf("generate tx error, err: %v\n", err)
		return nil, err
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		fmt.Printf("serialize tx error, err: %v\n", err)
		return nil, err
	}
	fmt.Println("------ RAW TRANSACTION ------------------------")
	fmt.Printf("%s\n", hex.EncodeToString(rawTx))
	fmt.Println("------ END RAW TRANSACTION ------------------------")

	fmt.Println("------ RAW MESSAGE ------------------------")
	fmt.Printf("%s\n", hex.EncodeToString(serializedMessage))
	fmt.Println("------ END RAW MESSAGE ------------------------")

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		fmt.Printf("send tx error, err: %v\n", err)
		return nil, err
	}

	log.Println("txHash:", txSig)

	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
		Message:           &message,
	}, nil
}

func initGravity(ccmd *cobra.Command, args []string) {
	endpoint, _ := InferSystemDefinedRPC()
	_, err := InitGravity(UpdateConsulsPrivateKey, GravityProgramID, GravityDataAccount, MultisigDataAccount, endpoint, make([]byte, 0))
	if err != nil {
		log.Fatalf("Error on 'InitGravity': %v\n", err)
	}
}
