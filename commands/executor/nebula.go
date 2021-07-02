package executor

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"solanoid/models"

	"github.com/mr-tron/base58"
	solclient "github.com/portto/solana-go-sdk/client"
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
	Instruction uint8
	Bft         uint8
	Oracles     []byte
	NewRound    uint64
}

type SendHashValueNebulaContractInstructionn struct {
	Instruction uint8
	DataHash    []byte
}

type DataType [1]byte
type PulseID [8]byte
type SubscriptionID [16]byte

type SendValueToSubsNebulaContractInstructionn struct {
	Instruction    uint8
	DataHash       []byte
	DataType       DataType
	PulseID        PulseID
	SubscriptionID SubscriptionID
}

type SubscribeNebulaContractInstruction struct {
	Instruction          uint8
	Subscriber       [32]byte
	MinConfirmations      uint8
	Reward               uint64
	SubscriptionID   [16]byte
}

type SendValueToSubsNebulaContractInstruction struct {
	Instruction        uint8
	DataValue      [64]byte
	DataType           uint8
	PulseID            uint64
	SubscriptionID [16]byte
}
type SendHashValueNebulaContractInstruction struct {
	Instruction        uint8
	DataValue      [64]byte
}


type NebulaInstructionBuilder struct{}

func (port *NebulaInstructionBuilder) Init(bft, dataType uint8, gravityProgramID common.PublicKey, oracles []byte) interface{} {
	return InitNebulaContractInstruction {
		Instruction:              0,
		Bft:                      bft,
		NebulaDataType:           dataType,
		GravityContractProgramID: gravityProgramID,
		InitialOracles:           oracles,
	}
}

func (port *NebulaInstructionBuilder) Subscribe(subscriber common.PublicKey, minConfirmations uint8, reward uint64, subscriptionID [16]byte) interface{} {
	return SubscribeNebulaContractInstruction {
		Instruction:     4,
		Subscriber:      subscriber,    
		MinConfirmations: minConfirmations, 
		Reward:          reward,
		SubscriptionID:  subscriptionID,
	}
}

func (port *NebulaInstructionBuilder) SendValueToSubs(data [64]byte, dataType uint8, pulseID uint64, subscriptionID [16]byte) interface{} {
	return SendValueToSubsNebulaContractInstruction {
		Instruction:     3,
		DataValue:       data,
		DataType:        dataType,
		PulseID:         pulseID, 
		SubscriptionID:  subscriptionID,
	}
}

func (port *NebulaInstructionBuilder) SendHashValue(data [64]byte) interface{} {
	return SendHashValueNebulaContractInstruction {
		Instruction:     2,
		DataValue:       data,
	}
}


type ExecutionVisitor interface {
	InvokePureInstruction(interface{}) (*models.CommandResponse, error)
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
	return types.AccountMeta{PubKey: signer.account.PublicKey, IsSigner: true, IsWritable: false}
}

type GenericExecutor struct {
	deployerPrivKey types.Account
	nebulaProgramID     string

	dataAccount         string
	multisigDataAccount string

	clientEndpoint      string

	signers           []GravityBftSigner
	additionalMeta    []types.AccountMeta

	client             *solclient.Client
}

func (ge *GenericExecutor) Deployer() common.PublicKey {
	return ge.deployerPrivKey.PublicKey
}

func (ge *GenericExecutor) SetAdditionalSigners(signers []GravityBftSigner) {
	ge.signers = signers
}
func (ge *GenericExecutor) EraseAdditionalSigners() {
	ge.signers = make([]GravityBftSigner, 0)
}

func (ge *GenericExecutor) SetDeployerPK(pk types.Account) {
	ge.deployerPrivKey = pk
}

func (ge *GenericExecutor) SetAdditionalMeta(meta []types.AccountMeta) {
	ge.additionalMeta = meta
}
func (ge *GenericExecutor) EraseAdditionalMeta() {
	ge.additionalMeta = make([]types.AccountMeta, 0)
}

func (ge *GenericExecutor) InvokePureInstruction(instruction interface{}) (*models.CommandResponse, error) {
	account := ge.deployerPrivKey

	if ge.client == nil {
		ge.client = solclient.NewClient(ge.clientEndpoint)
	}

	c := ge.client

	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		fmt.Printf("get recent block hash error, err: %v\n", err)
		return nil, err
	}

	builtInstruction, err := ge.BuildInstruction(instruction)
	if err != nil {
		return nil, err
	}
	for i, v := range builtInstruction.Accounts {
		fmt.Printf("INSTRUCTION ACCOUNT #%d - %s\n", i, v.PubKey.ToBase58())
	}

	message := types.NewMessage(
		account.PublicKey,
		[]types.Instruction{*builtInstruction},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		fmt.Printf("serialize message error, err: %v\n", err)
		return nil, err
	}

	signatures := map[common.PublicKey]types.Signature{
		account.PublicKey: ed25519.Sign(account.PrivateKey, serializedMessage),
	}
	for _, signer := range ge.signers {
		signatures[signer.Meta().PubKey] = signer.Sign(serializedMessage)
	}

	tx, err := types.CreateTransaction(message, signatures)

	if err != nil {
		fmt.Printf("generate tx error, err: %v\n", err)
		return nil, err
	}

	rawTx, err := tx.Serialize()

	logTx := func() {
		fmt.Println("------ RAW TRANSACTION ------------------------")
		// fmt.Printf("%s\n", hex.EncodeToString(rawTx))
		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(rawTx))
		fmt.Println("------ END RAW TRANSACTION ------------------------")

		fmt.Println("------ RAW MESSAGE ------------------------")
		// fmt.Printf("%s\n", hex.EncodeToString(serializedMessage))
		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(serializedMessage))

		fmt.Println("------ END RAW MESSAGE ------------------------")
	}
	logTx()
	
	if err != nil {
		fmt.Printf("serialize tx error, err: %v\n", err)
		// logTx()
		return nil, err
	}

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		fmt.Printf("send tx error, err: %v\n", err)
		// logTx()
		return nil, err
	}

	log.Println("txHash:", txSig)
	return &models.CommandResponse{
		SerializedMessage: hex.EncodeToString(serializedMessage),
		TxSignature:       txSig,
	}, nil
}

func (ge *GenericExecutor) BuildAndInvoke(instruction interface{}) (*models.CommandResponse, error) {
	return ge.InvokePureInstruction(instruction)
}

func (ge *GenericExecutor) BuildInstruction(instruction interface{}) (*types.Instruction, error) {
	data, err := common.SerializeData(instruction)

	if err != nil {
		panic(err)
	}

	// fmt.Println("--------- RAW INSTRUCTION DATA -----------")
	// fmt.Printf("%s\n", hex.EncodeToString(data))
	// fmt.Println("------- END RAW INSTRUCTION DATA ---------")

	accountMeta := []types.AccountMeta {
		{ PubKey: ge.deployerPrivKey.PublicKey, IsSigner: true, IsWritable: false },
		{ PubKey: common.PublicKeyFromString(ge.dataAccount), IsSigner: false, IsWritable: true },
	}

	if ge.multisigDataAccount != "" {
		accountMeta = append(accountMeta, types.AccountMeta{PubKey: common.PublicKeyFromString(ge.multisigDataAccount), IsSigner: false, IsWritable: true})
	}

	for _, signer := range ge.signers {
		accountMeta = append(accountMeta, signer.Meta())
	}

	accountMeta = append(accountMeta, ge.additionalMeta...)

	return &types.Instruction{
		Accounts:  accountMeta,
		ProgramID: common.PublicKeyFromString(ge.nebulaProgramID),
		Data:      data,
	}, nil
}

func NewNebulaExecutor(privateKey, nebulaProgramID, dataAccount, multisigDataAccount, clientEndpoint string, gravityProgramID common.PublicKey) (*GenericExecutor, error) {
	pk, err := base58.Decode(privateKey)
	if err != nil {
		zap.L().Fatal(err.Error())
		return nil, err
	}
	account := types.AccountFromPrivateKeyBytes(pk)

	return &GenericExecutor{
		deployerPrivKey: account,
		nebulaProgramID: nebulaProgramID,

		dataAccount:         dataAccount,
		multisigDataAccount: multisigDataAccount,

		clientEndpoint: clientEndpoint,
	}, nil
}
