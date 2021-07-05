package mvp

import (
	"math/big"
	"testing"
	"time"
)










func TestDepositAwaiter(t *testing.T) {
	polygonClient := NewPolygonExplorerClient(time.Second)

	var polygonDepositAwaiter, solanaDepositAwaiter CrossChainTokenDepositAwaiter


	amountTarget := big.NewInt(0)
	amountTarget.SetString("100000000000000", 10)

	type DepositBuffer struct {
		// MintBuffer chan bool
		// Unlock     chan *EVMTokenTransferEvent
		MintBuffer chan interface{}
		Unlock     chan interface{}
	}

	transferBuff := &DepositBuffer{}
	// transferBuff := make(chan *EVMTokenTransferEvent)

	polygonDepositAwaiter = NewGenericDepositAwaiter().
		SetComparator(func(i interface{}) bool {
			return polygonClient.IsAwaitedDeposit(i, "0xbbc3d3f8c70c1a558bd0b5c25662aa3226b863e9", "0xf480f38c366daac4305dc484b2ad7a496ff00cea", amountTarget)
		}).
		SetRetriever(func() (*interface{}, error) {
			var result interface{}
			deposits, err := polygonClient.RequestLastDeposits("0xbbc3d3f8c70c1a558bd0b5c25662aa3226b863e9", 1)
			result = deposits
			return &result, err
		})

	polygonDepositAwaiter.AwaitTokenDeposit(transferBuff.MintBuffer)
		
	// solanaDepositAwaiter = NewGenericDepositAwaiter()
	_ = solanaDepositAwaiter



	// polygonExplorerClient := NewPolygonExplorerClient(time.Second)


	// go func() {
	// 	polygonExplorerClient.AwaitTokenDeposit(
	// 		"0xbbc3d3f8c70c1a558bd0b5c25662aa3226b863e9",
	// 		"0xf480f38c366daac4305dc484b2ad7a496ff00cea", 
	// 		16298558,
	// 		amountTarget,
	// 		transferBuff,
	// 	)	
	// }()

	// for awaitedTransfer := range transferBuff {
	// 	fmt.Printf("transfer is: %v \n", awaitedTransfer.Hash)
	// }
}