package commands

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/Gravity-Tech/solanoid/models"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/sysprog"
	soltoken "github.com/portto/solana-go-sdk/tokenprog"
	"github.com/portto/solana-go-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	GravityContractAllocation = 299
	MultisigAllocation        = 355
	IBPortAllocation          = 20000
	LUPortAllocation          = 20000
	NebulaAllocation          = 1500
)

var (
	newDataAccPrivateKey string
	programPrivateKey    string

	space uint64
	// alias for show
	newDataAccCmd = &cobra.Command{
		Hidden: false,

		Use:   "attach",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   newAccCommand,
	}
)

// init
func init() {
	newDataAccCmd.Flags().StringVarP(&programID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", SolanoidCmd.Flags().Lookup("program"))
	newDataAccCmd.MarkFlagRequired("program")

	newDataAccCmd.Flags().StringVarP(&newDataAccPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", SolanoidCmd.Flags().Lookup("private-key"))
	newDataAccCmd.MarkFlagRequired("private-key")

	newDataAccCmd.Flags().Uint64VarP(&space, "space", "s", 4, "space for data")
	viper.BindPFlag("space", SolanoidCmd.Flags().Lookup("space"))
	newDataAccCmd.MarkFlagRequired("space")

	SolanoidCmd.AddCommand(newDataAccCmd)
}

func newAccCommand(ccmd *cobra.Command, args []string) {
	endpoint, _ := InferSystemDefinedRPC()
	_, _ = GenerateNewAccount(newDataAccPrivateKey, space, programID, endpoint)
	// if err != nil {
	// 	return
	// }
}

func AllocateAccount(deployerPrivateKey string, existingAccount types.Account, space uint64, programID, clientEndpoint string) (*models.CommandResponse, error) {
	pk, err := base58.Decode(deployerPrivateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	instruction := sysprog.Allocate(
		existingAccount.PublicKey,
		space,
	)
	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			instruction,
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey:         ed25519.Sign(account.PrivateKey, serializedMessage),
		existingAccount.PublicKey: ed25519.Sign(existingAccount.PrivateKey, serializedMessage),
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

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
		return nil, err
	}

	log.Println("txHash:", txSig)
	// fmt.Printf("Data Acc privake key: %s\n", base58.Encode(newAcc.PrivateKey))
	fmt.Printf("Data account address: %s\n", existingAccount.PublicKey.ToBase58())

	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
		Message:           &message,
		Account:           &existingAccount,
	}, nil
}

func GenerateNewAccountWithSeed(privateKey string, newAcc types.Account, space uint64, programID, clientEndpoint string) (*models.CommandResponse, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	program := common.PublicKeyFromString(programID)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	rentBalance, err := c.GetMinimumBalanceForRentExemption(context.Background(), space)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	instruction := sysprog.CreateAccount(
		account.PublicKey,
		newAcc.PublicKey,
		program,
		rentBalance,
		space,
	)
	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			instruction,
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	// fmt.Println("------- begin message --------")
	// fmt.Println(hex.EncodeToString(serializedMessage))
	// fmt.Println("-------- end message ---------")

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
		newAcc.PublicKey:  ed25519.Sign(newAcc.PrivateKey, serializedMessage),
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

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
		return nil, err
	}

	log.Println("txHash:", txSig)
	// fmt.Printf("Data Acc privake key: %s\n", base58.Encode(newAcc.PrivateKey))
	fmt.Printf("Data account address: %s\n", newAcc.PublicKey.ToBase58())

	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
		Message:           &message,
		Account:           &newAcc,
	}, nil
}

func GenerateNewTokenAccount(privateKey string, space uint64, owner, tokenMint common.PublicKey, clientEndpoint string, seeds string) (*models.CommandResponse, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	newAcc := types.NewAccount()

	rentBalance, err := c.GetMinimumBalanceForRentExemption(context.Background(), space)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}

	instruction := sysprog.CreateAccount(
		account.PublicKey,
		newAcc.PublicKey,
		owner,
		rentBalance,
		space,
	)
	initializeAccountIx := soltoken.InitializeAccount(
		newAcc.PublicKey,
		tokenMint,
		owner,
	)

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			instruction,
			initializeAccountIx,
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	fmt.Println("------- begin message --------")
	fmt.Println(base64.StdEncoding.EncodeToString(serializedMessage))
	fmt.Println("-------- end message ---------")

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
		newAcc.PublicKey:  ed25519.Sign(newAcc.PrivateKey, serializedMessage),
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

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
		return nil, err
	}

	log.Println("txHash:", txSig)
	// fmt.Printf("Data Acc privake key: %s\n", base58.Encode(newAcc.PrivateKey))
	fmt.Printf("Data account address: %s\n", newAcc.PublicKey.ToBase58())

	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
		Message:           &message,
		Account:           &newAcc,
	}, nil
}

func GenerateNewAccount(privateKey string, space uint64, programID, clientEndpoint string) (*models.CommandResponse, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	program := common.PublicKeyFromString(programID)

	c := client.NewClient(clientEndpoint)

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	newAcc := types.NewAccount()

	rentBalance, err := c.GetMinimumBalanceForRentExemption(context.Background(), space)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	instruction := sysprog.CreateAccount(
		account.PublicKey,
		newAcc.PublicKey,
		program,
		rentBalance,
		space,
	)
	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			instruction,
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
		return nil, err
	}

	// fmt.Println("------- begin message --------")
	// fmt.Println(hex.EncodeToString(serializedMessage))
	// fmt.Println("-------- end message ---------")

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
		newAcc.PublicKey:  ed25519.Sign(newAcc.PrivateKey, serializedMessage),
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

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
		return nil, err
	}

	log.Println("txHash:", txSig)
	// fmt.Printf("Data Acc privake key: %s\n", base58.Encode(newAcc.PrivateKey))
	fmt.Printf("Data account address: %s\n", newAcc.PublicKey.ToBase58())

	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
		Message:           &message,
		Account:           &newAcc,
	}, nil
}
