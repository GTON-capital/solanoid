package mvp

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)










func TestDepositAwaiter(t *testing.T) {
	polygonExplorerClient := NewPolygonExplorerClient(time.Second)

	amountTarget := big.NewInt(0)
	amountTarget.SetString("100000000000000", 10)

	transferBuff := make(chan *EVMTokenTransferEvent)

	go func() {
		polygonExplorerClient.AwaitTokenDeposit(
			"0xbbc3d3f8c70c1a558bd0b5c25662aa3226b863e9",
			"0xf480f38c366daac4305dc484b2ad7a496ff00cea", 
			16298558,
			amountTarget,
			transferBuff,
		)	
	}()

	for awaitedTransfer := range transferBuff {
		fmt.Printf("transfer is: %v \n", awaitedTransfer.Hash)
	}
}