package commands

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	// "solanoid/commands/executor"
	"solanoid/models"
	"solanoid/models/endpoint"
	"solanoid/models/nebula"

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

func NewInitNebulaContractInstruction(fromAccount, programData, multisigData, targetProgramID, gravityProgramID common.PublicKey, Bft uint8, Consuls [5][32]byte) types.Instruction {
	consuls := []byte{}
	consuls = append(consuls, fromAccount.Bytes()...)
	
	fmt.Printf("Nebula programId: %s\n", targetProgramID.ToBase58())

	data, err := common.SerializeData(struct {
		Instruction              uint8
		Bft                      uint8
		NebulaDataType           uint8
		GravityContractProgramID common.PublicKey
		Consuls                  []byte
	} {
		Instruction:              0,
		Bft:                      1,
		NebulaDataType:           nebula.Bytes,
		GravityContractProgramID: gravityProgramID,
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
			{ PubKey: fromAccount, IsSigner: true, IsWritable: false },
			{ PubKey: programData, IsSigner: false, IsWritable: true },
			{ PubKey: multisigData, IsSigner: false, IsWritable: true },
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}

func NewNebulaUpdateOraclesContractInstruction(fromAccount, programData, multisigData, targetProgramID, gravityProgramID common.PublicKey, bft uint8, newRound uint64, newOracles [5][32]byte) types.Instruction {
	var oracles []byte

	for _, x := range newOracles {
		oracles = append(oracles, x[:]...)
	}

	// oracles = append(oracles, fromAccount.Bytes()...)
	
	fmt.Printf("Nebula programId: %s\n", targetProgramID.ToBase58())
	data, err := common.SerializeData(struct {
		Instruction              uint8
		Bft                      uint8
		Oracles                  []byte
		NewRound                 uint64
	} {
		Instruction:              1,
		Bft:                      bft,
		Oracles:                  oracles,
		NewRound:                 newRound,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	fmt.Printf("%s\n", hex.EncodeToString(data))
	fmt.Println("------- END RAW INSTRUCTION DATA ---------")
	return types.Instruction{
		Accounts: []types.AccountMeta{
			{ PubKey: fromAccount, IsSigner: true, IsWritable: false },
			{ PubKey: programData, IsSigner: false, IsWritable: true },
			{ PubKey: multisigData, IsSigner: false, IsWritable: true },
		},
		ProgramID: targetProgramID,
		Data:      data,
	}
}



func PerformPureInvocation(privateKey, clientEndpoint string, instructionBuilder func() []types.Instruction) (*models.CommandResponse, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	message := types.NewMessage(
		account.PublicKey,
		instructionBuilder(),
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
		return nil, err
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
		return nil, err
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
		return nil, err
	}

	log.Println("txHash:", txSig)
	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature: txSig,
	}, nil
}

func InitNebula(privateKey, nebulaProgramID, nebulaDataAccount, nebulaMultisigDataAccount, clientEndpoint string, gravityProgramID common.PublicKey) (*models.CommandResponse, error) {
	// pk, err := base58.Decode(privateKey)
	// if err != nil {
	// 	zap.L().Fatal(err.Error())
	// 	return nil, err
	// }
	// account := types.AccountFromPrivateKeyBytes(pk)

	// return PerformPureInvocation(
	// 	privateKey,
	// 	clientEndpoint,
	// 	func() []types.Instruction {
	// 		return []types.Instruction{
	// 			NewInitNebulaContractInstruction(
	// 				account.PublicKey,
	// 				common.PublicKeyFromString(nebulaDataAccount),
	// 				common.PublicKeyFromString(nebulaMultisigDataAccount),
	// 				common.PublicKeyFromString(nebulaProgramID),
	// 				gravityProgramID, 
	// 				1, 
	// 				[5][32]byte{},
	// 			),
	// 		}
	// 	},
	// )
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	program := common.PublicKeyFromString(nebulaProgramID)
	fmt.Printf("program str: %v \n", nebulaProgramID)
	fmt.Printf("program bytes: %v \n", program)
	nebulaAcc := common.PublicKeyFromString(nebulaDataAccount)
	fmt.Printf("nebula: %v \n", nebulaDataAccount)
	fmt.Printf("nebula: %v \n", nebulaAcc)
	multisigAcc := common.PublicKeyFromString(nebulaMultisigDataAccount)
	fmt.Printf("multisigAcc: %v \n", nebulaMultisigDataAccount)
	fmt.Printf("multisigAcc: %v \n", multisigAcc)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			NewInitNebulaContractInstruction(
				account.PublicKey, nebulaAcc, multisigAcc, program, gravityProgramID, 1, [5][32]byte{},
			),
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
		return nil, err
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
		return nil, err
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
		return nil, err
	}

	log.Println("txHash:", txSig)
	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature: txSig,
	}, nil
}



func initNebula(ccmd *cobra.Command, args []string) {
	_, _ = InitNebula(UpdateConsulsPrivateKey, GravityDataAccount, NebulaDataAccount, MultisigDataAccount, endpoint.LocalEnvironment, common.PublicKeyFromString(GravityProgramID))
}
