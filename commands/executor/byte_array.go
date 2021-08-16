package executor

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/Gravity-Tech/solanoid/abstract"
	"github.com/portto/solana-go-sdk/common"
)

type BACfg struct {
	OriginDecimals, DestDecimals int
}

type CrossChainBridgeBABuilder interface {
	SetCfg(BACfg)
	BuildForDirect() []byte
	BuildForReverse() []byte
}

type EVMSOLByteArrayData struct {
	SwapID    []byte
	Amount    *big.Int
	Receiver  []byte
	Operation byte
}

type EVMToSolanaBABuilder struct {
	lastSwapID [32]byte
	cfg        BACfg

	// Amount *big.Int
	Amount   *big.Int
	Origin   [20]byte
	Receiver common.PublicKey
}

func (ets *EVMToSolanaBABuilder) SetCfg(cfg BACfg) {
	ets.cfg = cfg
}

func (ets *EVMToSolanaBABuilder) BuildForDirect() []byte {
	ets.lastSwapID = rndSwapID()
	var swapID [16]byte
	copy(swapID[:], ets.lastSwapID[:])

	// mapper := abstract.NewDecimalMapperFromFloat(ets.Amount, uint(ets.cfg.OriginDecimals))
	mapper := abstract.NewDecimalMapperFromBig(ets.Amount)
	amount := mapper.MapThrough(uint(ets.cfg.OriginDecimals), uint(ets.cfg.DestDecimals))

	floatAmount := float64(amount.Int64()) / math.Pow(10, float64(ets.cfg.DestDecimals))

	return buildMintSolana(swapID[:], ets.Receiver, floatAmount)
}

func (ets *EVMToSolanaBABuilder) BuildForReverse() []byte {
	ets.lastSwapID = rndSwapID()
	var swapID [32]byte
	copy(swapID[:], ets.lastSwapID[:])
	amount := ets.Amount

	return buildUnlockEVM(swapID[:], ets.Origin, amount)
}

type SolanaToEVMBABuilder struct {
	lastSwapID [32]byte
	cfg        BACfg

	// Amount *big.Int
	// Amount *big.Int
	Amount float64
	// Origin [32]byte
	Origin   common.PublicKey
	Receiver [20]byte
}

func (ets *SolanaToEVMBABuilder) SetCfg(cfg BACfg) {
	ets.cfg = cfg
}

func (ets *SolanaToEVMBABuilder) BuildForDirect() []byte {
	ets.lastSwapID = rndSwapID()
	var swapID [32]byte
	copy(swapID[:], ets.lastSwapID[:])

	mapper := abstract.NewDecimalMapperFromFloat(ets.Amount, uint(ets.cfg.OriginDecimals))
	amount := mapper.MapThrough(uint(ets.cfg.OriginDecimals), uint(ets.cfg.DestDecimals))

	return buildMintEVM(swapID[:], ets.Receiver, amount)
}

func (ets *SolanaToEVMBABuilder) BuildForReverse() []byte {
	ets.lastSwapID = rndSwapID()
	var swapID [16]byte
	copy(swapID[:], ets.lastSwapID[:])

	return buildUnlockSolana(swapID[:], ets.Origin, ets.Amount)
}

func rndSwapID() [32]byte {
	var subID [32]byte
	rand.Read(subID[:])
	return subID
}

func buildMintEVM(swapId []byte, receiver [20]byte, amount *big.Int) []byte {
	var res []byte

	// action
	res = append(res, 'm')
	// swap id
	res = append(res, swapId[0:32]...)
	// amount
	res = append(res, amount.Bytes()...)
	// receiver
	res = append(res, receiver[:]...)

	return res
}

func buildUnlockEVM(swapId []byte, receiver [20]byte, amount *big.Int) []byte {
	var res []byte

	// action
	res = append(res, 'u')
	// swap id
	res = append(res, swapId[0:32]...)
	// amount
	res = append(res, amount.Bytes()...)
	// receiver
	res = append(res, receiver[:]...)

	return res
}

func buildMintSolana(swapId []byte, receiver common.PublicKey, amount float64) []byte {
	var res []byte

	// action
	res = append(res, 'm')
	// swap id
	res = append(res, swapId[0:16]...)
	// amount
	res = append(res, Float64ToBytes(amount)...)
	// receiver
	res = append(res, receiver[:]...)

	return res
}

func buildUnlockSolana(swapId []byte, receiver common.PublicKey, amount float64) []byte {
	var res []byte

	// action
	res = append(res, 'u')
	// swap id
	res = append(res, swapId[0:16]...)
	// amount
	res = append(res, Float64ToBytes(amount)...)
	// receiver
	res = append(res, receiver[:]...)

	return res
}

func WrapIntoConfirmedRequest(bytearray []byte) []byte {
	bytearray[0] = 'c'
	return bytearray
}
