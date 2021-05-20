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
	// Oracles                  []byte
	NewRound                 uint64
}

type SendHashValueNebulaContractInstructionn struct {
	Instruction              uint8
	DataHash                 []byte
}


type DataType [1]byte
type PulseID [8]byte
type SubscriptionID [16]byte

type SendValueToSubsNebulaContractInstructionn struct {
	Instruction              uint8
	DataHash                 []byte
	DataType                 DataType
	PulseID                  PulseID
	SubscriptionID           SubscriptionID
}

type SubscribeNebulaContractInstructionn struct {
	Instruction              uint8
	Subscriber               [32]byte
	MinConfirmations          uint8
	Reward                   uint64
}

type SignerDelegate interface {
	Sign([]byte) []byte
	Pubkey() string
	Meta() types.AccountMeta
}

type GravityBftSigner struct {
	account types.Account
}

func NewGravityBftSigner(privateKey string) *GravityBftSigner {
	pk, _ := base58.Decode(privateKey)
	account := types.AccountFromPrivateKeyBytes(pk)

	return &GravityBftSigner{
		account,
	}
}

func (signer *GravityBftSigner) Sign(message []byte) []byte {
	return ed25519.Sign(signer.account.PrivateKey, message)
}

func (signer *GravityBftSigner) Pubkey() string {
	return signer.account.PublicKey.ToBase58()
}

func (signer *GravityBftSigner) Meta() types.AccountMeta {
	return types.AccountMeta{ PubKey: signer.account.PublicKey, IsSigner: true, IsWritable: false }
}

type NebulaInstructionExecutor struct {
	deployerPrivKey types.Account
	nebulaProgramID string

	nebulaDataAccount string
	nebulaMultisigDataAccount string

	clientEndpoint string

	signers        []GravityBftSigner
	additionalMeta []types.AccountMeta
}

func (nexe *NebulaInstructionExecutor) Deployer() common.PublicKey {
	return nexe.deployerPrivKey.PublicKey
}

func (nexe *NebulaInstructionExecutor) SetAdditionalSigners(signers []GravityBftSigner) {
	nexe.signers = signers
}
func (nexe *NebulaInstructionExecutor) EraseAdditionalSigners() {
	nexe.signers = make([]GravityBftSigner, 0)
}

func (nexe *NebulaInstructionExecutor) SetAdditionalMeta(meta []types.AccountMeta) {
	nexe.additionalMeta = meta
}
func (nexe *NebulaInstructionExecutor) EraseAdditionalMeta() {
	nexe.additionalMeta = make([]types.AccountMeta, 0)
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

	signatures := map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	}
	for _, signer := range nexe.signers {
		signatures[signer.Meta().PubKey] = signer.Sign(serializedMessage)
	}

	tx, err := types.CreateTransaction(message, signatures)


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

	accountMeta := []types.AccountMeta{
		{ PubKey: nexe.deployerPrivKey.PublicKey, IsSigner: true, IsWritable: false },
		{ PubKey: common.PublicKeyFromString(nexe.nebulaDataAccount), IsSigner: false, IsWritable: true },
		{ PubKey: common.PublicKeyFromString(nexe.nebulaMultisigDataAccount), IsSigner: false, IsWritable: true },	
	}
	
	for _, signer := range nexe.signers {
		accountMeta = append(accountMeta, signer.Meta())
	}
	
	accountMeta = append(accountMeta, nexe.additionalMeta...)
	
	return &types.Instruction{
		Accounts: accountMeta,
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

