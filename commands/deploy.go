package commands

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/sysprog"
	"github.com/portto/solana-go-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	BPFLoader2ProgramID common.PublicKey = common.PublicKeyFromString("BPFLoader2111111111111111111111111111111111")
	filename            string
	privateKey          string
	// alias for show
	deployCmd = &cobra.Command{
		Hidden: false,

		Use:   "deploy",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   deploy,
	}
)

func splitArray(in []byte, chunkSize int) [][]byte {
	var divided [][]byte

	for i := 0; i < len(in); i += chunkSize {
		end := i + chunkSize

		if end > len(in) {
			end = len(in)
		}

		divided = append(divided, in[i:end])
	}

	return divided
}

// init
func init() {
	deployCmd.Flags().StringVarP(&filename, "program-file", "p", "program.so", "Path to file for deploy (i.e. program.so)")
	viper.BindPFlag("program-file", SolanoidCmd.Flags().Lookup("program-file"))
	deployCmd.MarkFlagRequired("program-file")

	deployCmd.Flags().StringVarP(&privateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", SolanoidCmd.Flags().Lookup("private-key"))
	deployCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(deployCmd)
}
func createNewAccountForProgram(c *client.Client, account types.Account, space uint64) types.Account {

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	program := types.NewAccount()

	rentBalance, err := c.GetMinimumBalanceForRentExemption(context.Background(), space)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			sysprog.CreateAccount(
				account.PublicKey,
				program.PublicKey,
				BPFLoader2ProgramID,
				rentBalance,
				space,
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
		program.PublicKey: ed25519.Sign(program.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
	}

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Println("txHash:", txSig)
	return program
}
func uploadDataToProgram(c *client.Client, program types.Account, account types.Account, data []byte, chunkSize int) {
	chunks := splitArray(data, chunkSize)
	for i, chunk := range chunks {
		chunkData, err := common.SerializeData(struct {
			Instruction uint32
			Offset      uint32
			Length      uint32
			Padding     uint32
			Data        []byte
		}{
			Instruction: 0,
			Offset:      uint32(i) * uint32(chunkSize),
			Length:      uint32(len(chunk)),
			Data:        chunk,
		})
		if err != nil {
			log.Fatalf("Serialize error, err: %v\n", err)
		}
		//zap.L().Sugar().Debug("data: ", chunkData)
		res, err := c.GetRecentBlockhash(context.Background())
		if err != nil {
			log.Fatalf("get recent block hash error, err: %v\n", err)
		}

		message2 := types.NewMessage(
			account.PublicKey,
			[]types.Instruction{
				{
					ProgramID: BPFLoader2ProgramID,
					Accounts: []types.AccountMeta{
						{
							PubKey:     program.PublicKey,
							IsSigner:   true,
							IsWritable: true,
						},
					},
					Data: chunkData,
				},
			},
			res.Blockhash,
		)

		serializedMessage2, err := message2.Serialize()
		if err != nil {
			log.Fatalf("serialize message error, err: %v\n", err)
		}

		tx2, err := types.CreateTransaction(message2, map[common.PublicKey]types.Signature{
			program.PublicKey: ed25519.Sign(program.PrivateKey, serializedMessage2),
			account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage2),
		})
		if err != nil {
			log.Fatalf("generate tx error, err: %v\n", err)
		}

		rawTx2, err := tx2.Serialize()
		if err != nil {
			log.Fatalf("serialize tx error, err: %v\n", err)
		}

		tx2Sig, err := c.SendRawTransaction(context.Background(), rawTx2)
		if err != nil {
			log.Fatalf("send tx error, err: %v\n", err)
		}

		log.Printf("upload program txHash [chunk %d]: %s", i, tx2Sig)
		time.Sleep(time.Second * 1)

	}
}

func finalizeProgramDeployment(c *client.Client, program types.Account, account types.Account) {
	finalizeData, err := common.SerializeData(uint32(1))
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message3 := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			{
				ProgramID: BPFLoader2ProgramID,
				Accounts: []types.AccountMeta{
					{
						PubKey:     program.PublicKey,
						IsSigner:   true,
						IsWritable: true,
					},
				},
				Data: finalizeData,
			},
		},
		res.Blockhash,
	)

	serializedMessage3, err := message3.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
	}

	tx2, err := types.CreateTransaction(message3, map[common.PublicKey]types.Signature{
		program.PublicKey: ed25519.Sign(program.PrivateKey, serializedMessage3),
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage3),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
	}

	rawTx3, err := tx2.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
	}

	tx3Sig, err := c.SendRawTransaction(context.Background(), rawTx3)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}
	log.Printf("Finalize txHash: %s", tx3Sig)
}

func createAttachedAccountToProgram(c *client.Client, program types.Account, account types.Account, space uint64) types.Account {
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	newAcc := types.NewAccount()

	rentBalance, err := c.GetMinimumBalanceForRentExemption(context.Background(), space)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{
			sysprog.CreateAccount(
				account.PublicKey,
				newAcc.PublicKey,
				program.PublicKey,
				rentBalance,
				space,
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
		newAcc.PublicKey:  ed25519.Sign(newAcc.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
	}

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Println("txHash:", txSig)
	return newAcc
}

// show utilizes the api to show data associated to key
func deploy(ccmd *cobra.Command, args []string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	c := client.NewClient(client.TestnetRPCEndpoint)

	program := createNewAccountForProgram(c, account, uint64(len(data)))

	fmt.Print("Waiting  ")
	for i := 0; i < 25; i++ {
		time.Sleep(time.Second * 1)
		fmt.Print(".")
	}
	fmt.Print("\n")
	//deploy start
	uploadDataToProgram(c, program, account, data, 940)

	fmt.Print("Waiting  ")
	for i := 0; i < 25; i++ {
		time.Sleep(time.Second * 1)
		fmt.Print(".")
	}
	fmt.Print("\n")

	//finalize
	finalizeProgramDeployment(c, program, account)

	newAcc := createAttachedAccountToProgram(c, program, account, uint64(4))
	fmt.Printf("Program PubKey(Id): %s\n", program.PublicKey.ToBase58())
	fmt.Printf("Program PrivateKey: %s\n", base58.Encode(program.PrivateKey))
	fmt.Printf("Data account PubKey(Id): %s\n", newAcc.PublicKey.ToBase58())
	fmt.Printf("Data account PrivateKey: %s\n", base58.Encode(newAcc.PrivateKey))
}
