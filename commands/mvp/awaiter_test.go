package mvp

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Gravity-Tech/solanoid/commands"
	soltoken "github.com/portto/solana-go-sdk/tokenprog"
)

func TestDepositAwaiter(t *testing.T) {
	polygonClient := NewEVMExplorerClient()

	var err error
	var polygonDepositAwaiter, solanaDepositAwaiter CrossChainTokenDepositAwaiter

	// poly client implements awaiter interface
	polygonDepositAwaiter = polygonClient

	solanaEndpoint, _ := commands.InferSystemDefinedRPC()

	solanaDepositAwaiter = NewSolanaDepositAwaiter(solanaEndpoint)

	// 437000000000000
	polygonWatchAmount := big.NewInt(437000000000000)

	polygonDepositAwaiter.SetCfg(
		&CrossChainDepositAwaiterConfig{
			WatchAddress:    "0xbbc3d3f8c70c1a558bd0b5c25662aa3226b863e9",
			WatchAssetID:    "0xf480f38c366daac4305dc484b2ad7a496ff00cea",
			WatchAmount:     polygonWatchAmount,
			BlockStart:      16298558,
			PerAwaitTimeout: time.Second,
		},
	)

	polygonDepositBuffer := make(chan interface{})

	go func() {
		err = polygonDepositAwaiter.AwaitTokenDeposit(polygonDepositBuffer)
		commands.ValidateError(t, err)
	}()

	for event := range polygonDepositBuffer {
		fmt.Printf("EVM - deposit event: %v \n", event)
	}

	solanaDepositBuffer := make(chan interface{})

	solanaDepositAwaiter.SetCfg(
		&CrossChainDepositAwaiterConfig{
			WatchAddress:    "FMtjwGs2V6j3eWvZhLA18tkHuzvBHfpjFcCuuvsweuwC",
			WatchAssetID:    "nVZnRKdr3pmcgnJvYDE8iafgiMiBqxiffQMcyv5ETdA",
			WatchAmount:     big.NewInt(1),
			PerAwaitTimeout: time.Second,
		},
	)

	go func() {
		err = solanaDepositAwaiter.AwaitTokenDeposit(solanaDepositBuffer)
		commands.ValidateError(t, err)
	}()

	for event := range solanaDepositBuffer {
		tokenDataState := event.(soltoken.TokenAccount)
		fmt.Printf("SOL - deposit event: %+v \n", tokenDataState)
	}
}
