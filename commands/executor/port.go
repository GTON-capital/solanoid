package executor

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/portto/solana-go-sdk/common"
)

func NewIBPortInstructionBuilder() *IBPortInstructionBuilder {
	return &IBPortInstructionBuilder{}
}

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

func float64ToByte(f float64) []byte {
	//bits := math.Float64bits(f)
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func (port *IBPortInstructionBuilder) TestMint(receiver common.PublicKey, amount float64) interface{} {
	amountBytes := float64ToByte(amount)
	fmt.Printf("TestMint - amountBytes: %v", amountBytes)

	// binary.LittleEndian.

	return struct {
		Instruction uint8
		Receiver    common.PublicKey
		TokenAmount []byte
	}{
		Instruction: 4,
		Receiver:    receiver,
		TokenAmount: amountBytes,
	}
}
func (port *IBPortInstructionBuilder) TestBurn(burner common.PublicKey, amount float64) interface{} {
	amountBytes := float64ToByte(amount)
	fmt.Printf("TestBurn - amountBytes: %v", amountBytes)

	return struct {
		Instruction uint8
		Burner      common.PublicKey
		TokenAmount []byte
	}{
		Instruction: 5,
		Burner:      burner,
		TokenAmount: amountBytes,
	}
}
