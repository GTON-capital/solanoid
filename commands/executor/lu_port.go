package executor

import (
	"crypto/rand"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/portto/solana-go-sdk/common"
)

// const (
// 	LUPortPDABumpSeeds = "ibport"
// )

var LUPortIXBuilder = &LUPortInstructionBuilder{}

type LUPortInstructionBuilder struct{}

func (port *LUPortInstructionBuilder) InitWithOracles(nebula, token, tokenMint common.PublicKey, bft uint8, oracles []byte) interface{} {
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

type CreateTransferWrapRequestInstruction struct {
	Instruction uint8
	TokenAmount []byte
	Receiver    [32]byte
	RequestID   [16]byte
}

func (ix *CreateTransferWrapRequestInstruction) Pack() []byte {
	var res []byte

	res = append(res, 'u')
	res = append(res, ix.RequestID[:]...)
	res = append(res, ix.TokenAmount[:]...)
	res = append(res, ix.Receiver[:]...)

	return res
}


func RandomEVMAddress() [32]byte {
	ethReceiverPK, _ := ethcrypto.GenerateKey()

	var ethReceiverAddress [32]byte
	copy(ethReceiverAddress[:], ethcrypto.PubkeyToAddress(ethReceiverPK.PublicKey).Bytes())
	return ethReceiverAddress
}

func (port *LUPortInstructionBuilder) CreateTransferWrapRequest(receiver [32]byte, amount float64) interface{} {
	var requestID [16]byte

	// uint id = uint(keccak256(abi.encodePacked(msg.sender, receiver, block.number, amount)));

	rand.Read(requestID[:])

	fmt.Printf("CreateTransferUnwrapRequest - rq_id: %v amount: %v \n", requestID, amount)
	amountBytes := float64ToByte(amount)

	return CreateTransferWrapRequestInstruction {
		Instruction: 1,
		TokenAmount: amountBytes,
		Receiver:    receiver,
		RequestID:   requestID,
	}
}
func (port *LUPortInstructionBuilder) ConfirmProcessedRequest(requestID []byte) interface{} {
	return struct {
		Instruction     uint8
		RequestID     []byte
	}{
		Instruction: 3,
		RequestID:   requestID,
	}
}

func (port *LUPortInstructionBuilder) AttachValue(byte_vector []byte) interface{} {
	fmt.Printf("AttachValue - byte_vector: %v", byte_vector)

	return struct {
		Instruction uint8
		ByteVector  []byte
	}{
		Instruction: 2,
		ByteVector:  byte_vector,
	}
}

func (port *LUPortInstructionBuilder) TransferTokenOwnership(newOwner, newToken common.PublicKey) interface{} {
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

