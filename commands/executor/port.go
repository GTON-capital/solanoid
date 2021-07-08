package executor

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
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

func (port *IBPortInstructionBuilder) InitWithOracles(nebula, token common.PublicKey, bft uint8, oracles []byte) interface{} {
	return struct {
		Instruction       uint8
		NebulaDataAccount common.PublicKey
		TokenDataAccount  common.PublicKey
		Bft               uint8
		Oracles           []byte
	}{
		Instruction:       0,
		NebulaDataAccount: nebula,
		TokenDataAccount:  token,
		Bft:               bft,
		Oracles:           oracles,
	}
}

func Float64ToBytes(f float64) []byte {
	return float64ToByte(f)
}

func float64ToByte(f float64) []byte {
	//bits := math.Float64bits(f)
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
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

	// uint id = uint(keccak256(abi.encodePacked(msg.sender, receiver, block.number, amount)));

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

func BuildCrossChainMintByteVector(swapId []byte, receiver common.PublicKey, amount float64) []byte {
	var res []byte

	// action
	res = append(res, 'm')
	// swap id
	res = append(res, swapId[0:16]...)
	// amount
	res = append(res, Float64ToBytes(amount)...)
	// receiver
	res = append(res, receiver[:]...)

	fmt.Printf("byte array len: %v \n", len(res))
	fmt.Printf("byte array cap: %v \n", len(res))

	return res
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



// pub struct PortOperation<'a> {
//     pub action: u8,
//     pub swap_id: &'a [u8; 16],
//     pub amount: &'a [u8; 8],
//     // receiver: &'a [u8; 32],
//     pub receiver: &'a ForeignAddress,
// }

type PortOperation struct {
	Action        uint8
	SwapID    [16]byte
	Amount     [8]byte
	Receiver  [32]byte
}

const DefaultDecimals = 8

func (po *PortOperation) Pack() []byte {
	var res []byte

	res = append(res, 'm')
	res = append(res, po.SwapID[:]...)
	res = append(res, po.Amount[:]...)
	res = append(res, po.Receiver[:]...)

	return res
}

func UnpackByteArray(encoded []byte) (*PortOperation, error) {
	if len(encoded) < 57 {
		return nil, fmt.Errorf("invalid byte array length")
	}
	pos := 0
	action := encoded[0]
	pos += 1

	// swapId := encoded[pos:pos + 16]
	var swapId [16]byte
	copy(swapId[:], encoded[pos:pos + 16])
	pos += 16;
	
	var rawAmount [8]byte
	copy(rawAmount[:], encoded[pos:pos + 8])
	pos += 8;

	var receiver [32]byte
	copy(receiver[:], encoded[pos:pos + 32])

	return &PortOperation{
		action,
		swapId,
		rawAmount,
		receiver,
	}, nil
}
