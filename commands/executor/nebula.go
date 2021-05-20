package executor

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"solanoid/models"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	"go.uber.org/zap"
)

type InitNebulaContractInstruction struct {
	Instruction              uint8
	Bft                      uint8
	NebulaDataType           uint8
	GravityContractProgramID common.PublicKey
	InitialOracles           []byte
}

type UpdateOraclesNebulaContractInstruction struct {
	Instruction              uint8
	Bft                      uint8
	Oracles                  []byte
	NewRound                 uint64
}

type NebulaInstructionExecutor struct {
	deployerPrivKey types.Account
	nebulaProgramID string

	nebulaDataAccount string
	nebulaMultisigDataAccount string

	clientEndpoint string
}

func (nexe *NebulaInstructionExecutor) Deployer() common.PublicKey {
	return nexe.deployerPrivKey.PublicKey
}

func (nexe *NebulaInstructionExecutor) invokePureInstruction(instruction interface{}) (*models.CommandResponse, error) {
	account := nexe.deployerPrivKey

	c := client.NewClient(nexe.clientEndpoint)

	res, err := c.GetRecentBlockhash()
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	builtInstruction, err := nexe.BuildInstruction(instruction)
	if err != nil {
		return nil, err
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction { *builtInstruction },
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

func (nexe *NebulaInstructionExecutor) BuildAndInvoke(instruction interface{}) (*models.CommandResponse, error) {
	return nexe.invokePureInstruction(instruction)
}

func (nexe *NebulaInstructionExecutor) BuildInstruction(instruction interface{}) (*types.Instruction, error) {
	data, err := common.SerializeData(instruction)

	if err != nil {
		panic(err)
	}

	fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	fmt.Printf("%s\n", hex.EncodeToString(data))
	fmt.Println("------- END RAW INSTRUCTION DATA ---------")

	return &types.Instruction{
		Accounts: []types.AccountMeta{
			{ PubKey: nexe.deployerPrivKey.PublicKey, IsSigner: true, IsWritable: false },
			{ PubKey: common.PublicKeyFromString(nexe.nebulaDataAccount), IsSigner: false, IsWritable: true },
			{ PubKey: common.PublicKeyFromString(nexe.nebulaMultisigDataAccount), IsSigner: false, IsWritable: true },
		},
		ProgramID: common.PublicKeyFromString(nexe.nebulaProgramID),
		Data:      data,
	}, nil
}

func NewNebulaExecutor(privateKey, nebulaProgramID, nebulaDataAccount, nebulaMultisigDataAccount, clientEndpoint string, gravityProgramID common.PublicKey) (*NebulaInstructionExecutor, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	return &NebulaInstructionExecutor{
		deployerPrivKey: account,
		nebulaProgramID: nebulaProgramID,

		nebulaDataAccount: nebulaDataAccount,
		nebulaMultisigDataAccount: nebulaMultisigDataAccount,
	
		clientEndpoint: clientEndpoint,
	}, nil
}

