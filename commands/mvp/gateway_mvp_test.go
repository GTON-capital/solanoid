package mvp

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	commands "github.com/Gravity-Tech/solanoid/commands"

	luport "github.com/Gravity-Tech/gateway/abi/ethereum/luport"
	solcommon "github.com/portto/solana-go-sdk/common"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)

/*
 * Flow
 * 1. Lock N of GTON.
 * 2. Wait for GTON.
 * 3. Print TX of lock and issue.
 * 4. Burn GTON.
 * 5. Print TX of burn and unlock.
 *
 */
func TestRunPolygonToSolanaGatewayMVP(t *testing.T) {	
	// gtonToken := &crossChainToken{}

	gtonToken, err := NewCrossChainToken(&crossChainTokenCfg {
		originDecimals: 18,
		destinationDecimals: 8,
		originAddress: "0xf480f38c366daac4305dc484b2ad7a496ff00cea",
		destinationAddress: "FP5MgcQaD3ppWDqfjXouftsWQBSPW2suRzduLAFs712S",
	}, 0)
	commands.ValidateError(t, err)

	extractorCfg := &extractorCfg {
		originDecimals: 18,
		destinationDecimals: 8,
		chainID: 137,
		originNodeURL: "https://rpc-mainnet.matic.network",
		destinationNodeURL: "https://api.mainnet-beta.solana.com",
		luportAddress: "0x7725d618122F9A2Ce368dA1624Fbc79ce197c438",
		ibportAddress: "14dHNRGpDgn4rc27xUyWxp23M72ky33kwPaoYhR6uR7y",
	}

	polygonCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	polygonClient, err := ethclient.DialContext(polygonCtx, extractorCfg.originNodeURL)

	polygonGTONHolder, err := newEVMKey("76b77e0673cdf31a4bbfa0f0cdd9ed1fe02f036697d91dbf6293767f630e3b47")
	commands.ValidateError(t, err)

	transactor, err := ethbind.NewKeyedTransactorWithChainID(polygonGTONHolder.PrivKey, big.NewInt(extractorCfg.chainID))
	polygonTransactor := NewEVMTransactor(polygonClient, transactor)

	solanaGTONHolder, err := commands.ReadOperatingAddress(t, "../../private-keys/_from-polygon-gton-mvp-recipient.json")
	commands.ValidateError(t, err)
	_ = solanaGTONHolder

	solanaGTONTokenAccount, err := commands.CreateTokenAccount(solanaGTONHolder.PKPath, gtonToken.cfg.destinationAddress)
	commands.ValidateError(t, err)

	fmt.Printf("solanaGTONTokenAccount: %v \n", solanaGTONTokenAccount)

	luportClient, err := luport.NewLUPort(ethcommon.HexToAddress(extractorCfg.luportAddress), polygonClient)
	commands.ValidateError(t, err)

	// transferring 0.0001 GTON, 18 decimals (1 * 1e14)
	gtonToken.Set(0.000227)

	fmt.Printf("Locking %v GTON \n", gtonToken.Float())

	fmt.Printf("As Origin: %v GTON \n", gtonToken.AsOriginBigInt())
	fmt.Printf("As Destination: %v GTON \n", gtonToken.AsDestinationBigInt())

	// (1)
	lockFundsTx, err := luportClient.CreateTransferUnwrapRequest(
		polygonTransactor.transactor,
		gtonToken.AsOriginBigInt(), 
		solcommon.PublicKeyFromString(solanaGTONTokenAccount),
	)

	commands.ValidateError(t, err)

	_, err = ethbind.WaitMined(polygonCtx, polygonClient, lockFundsTx)



}