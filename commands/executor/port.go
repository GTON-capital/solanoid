package executor

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/portto/solana-go-sdk/common"
)

const (
	CommonGravityBumpSeeds = "ibport"
)

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
