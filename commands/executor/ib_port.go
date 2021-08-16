package executor

import (
	"crypto/rand"
	"fmt"

	"github.com/portto/solana-go-sdk/common"
)

const (
	IBPortPDABumpSeeds = "ibport"
)

func NewIBPortInstructionBuilder() *IBPortInstructionBuilder {
	return &IBPortInstructionBuilder{}
}

var IBPortIXBuilder = &IBPortInstructionBuilder{}

type IBPortInstructionBuilder struct{}

func (port *IBPortInstructionBuilder) Init(nebula, token common.PublicKey) interface{} {
	return struct {
		Instruction       uint8
		NebulaDataAccount common.PublicKey
		TokenDataAccount  common.PublicKey
	}{
		Instruction:       0,
		NebulaDataAccount: nebula,
		TokenDataAccount:  token,
	}
}


func (port *IBPortInstructionBuilder) InitWithOracles(nebula, token, tokenMint common.PublicKey, bft uint8, oracles []byte) interface{} {
	return struct {
		Instruction       uint8
		NebulaDataAccount common.PublicKey
		TokenDataAccount  common.PublicKey
		TokenMint         common.PublicKey
		Bft               uint8
		Oracles           []byte
	}{
		Instruction:       0,
		NebulaDataAccount: nebula,
		TokenDataAccount:  token,
		TokenMint:         tokenMint,
		Bft:               bft,
		Oracles:           oracles,
	}
}


type CreateTransferUnwrapRequestInstruction struct {
	Instruction uint8
	TokenAmount []byte
	Receiver    [32]byte
	RequestID   [16]byte
}

func (ix *CreateTransferUnwrapRequestInstruction) Pack() []byte {
	var res []byte

	res = append(res, 'm')
	res = append(res, ix.RequestID[:]...)
	res = append(res, ix.TokenAmount[:]...)
	res = append(res, ix.Receiver[:]...)

	return res
}

func (port *IBPortInstructionBuilder) CreateTransferUnwrapRequest(receiver [32]byte, amount float64) interface{} {
	var requestID [16]byte
	rand.Read(requestID[:])

	fmt.Printf("CreateTransferUnwrapRequest - rq_id: %v amount: %v \n", requestID, amount)
	amountBytes := float64ToByte(amount)

	return CreateTransferUnwrapRequestInstruction {
		Instruction: 1,
		TokenAmount: amountBytes,
		Receiver:    receiver,
		RequestID:   requestID,
	}
}
func (port *IBPortInstructionBuilder) ConfirmProcessedRequest(requestID []byte) interface{} {
	return struct {
		Instruction     uint8
		RequestID     []byte
	}{
		Instruction: 3,
		RequestID:   requestID,
	}
}

func (port *IBPortInstructionBuilder) AttachValue(byte_vector []byte) interface{} {
	fmt.Printf("AttachValue - byte_vector: %v", byte_vector)

	return struct {
		Instruction uint8
		ByteVector  []byte
	}{
		Instruction: 2,
		ByteVector:  byte_vector,
	}
}

func (port *IBPortInstructionBuilder) TransferTokenOwnership(newOwner, newToken common.PublicKey) interface{} {
	fmt.Printf("TransferOwnership - newOwner: %v, newToken: %v \n", newOwner, newToken)

	return struct {
		Instruction   uint8
		NewAuthority  common.PublicKey
		NewToken      common.PublicKey
	}{
		Instruction:  4,
		NewAuthority: newOwner,
		NewToken:     newToken,
	}
}

