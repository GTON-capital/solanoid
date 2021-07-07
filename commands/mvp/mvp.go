package mvp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"

	erc20 "github.com/Gravity-Tech/gateway/abi/ethereum/erc20"
	commands "github.com/Gravity-Tech/solanoid/commands"
	"github.com/Gravity-Tech/solanoid/commands/executor"

	luport "github.com/Gravity-Tech/gateway/abi/ethereum/luport"
	solcommon "github.com/portto/solana-go-sdk/common"
	soltoken "github.com/portto/solana-go-sdk/tokenprog"
	soltypes "github.com/portto/solana-go-sdk/types"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/mr-tron/base58"
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
func ProcessMVP_PolygonSolana() error {
	gtonToken, err := NewCrossChainToken(&crossChainTokenCfg{
		originDecimals:      18,
		destinationDecimals: 8,
		originAddress:       "0xf480f38c366daac4305dc484b2ad7a496ff00cea",
		destinationAddress:  "nVZnRKdr3pmcgnJvYDE8iafgiMiBqxiffQMcyv5ETdA",
	}, 0)

	if err != nil {
		return err
	}

	IBPortProgramPDA := solcommon.PublicKeyFromString("CYEnZhJdYaUjgFtGQ2FgXe4vp4zMiqY8RsdqwNFduxdm")

	extractorCfg := &extractorCfg{
		originDecimals:      18,
		destinationDecimals: 8,
		chainID:             137,
		originNodeURL:       "https://rpc-mainnet.maticvigil.com",
		// originNodeURL: "https://rpc-mainnet.matic.quiknode.pro",
		// originNodeURL: "https://matic-mainnet.chainstacklabs.com",
		destinationNodeURL: "https://api.mainnet-beta.solana.com",
		luportAddress:      "0x7725d618122F9A2Ce368dA1624Fbc79ce197c438",
		ibportDataAccount:  "9kwBfNbrQAEmEqkZbvMCKkefuJBj7nuqWrq6dzUhW5fJ",
		ibportProgramID:    "AH3QKaj942UUxDjaRaGh7hvdadsD8yfU9LRTa9KXfJkZ",
	}

	polygonCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	polygonClient, err := ethclient.DialContext(polygonCtx, extractorCfg.originNodeURL)
	if err != nil {
		return err
	}

	polygonGTONHolder, err := newEVMKey("76b77e0673cdf31a4bbfa0f0cdd9ed1fe02f036697d91dbf6293767f630e3b47")
	if err != nil {
		return err
	}

	transactor, err := ethbind.NewKeyedTransactorWithChainID(polygonGTONHolder.PrivKey, big.NewInt(extractorCfg.chainID))
	transactor.GasLimit = 10 * 150000
	transactor.Context = polygonCtx

	if err != nil {
		return err
	}

	polygonTransactor := NewEVMTransactor(polygonClient, transactor)

	solanaGTONHolderPrivKeyEncoded := "v2THMqDAX3SrM0o0GqzlaEDVdG2zMuB5fYKj/qCrUml44x+mNvktmjXQlP7AzuhuM3k+EJvHBtNgx/OcCU4UNw=="
	decodedSolanaGTONHolderPrivKey, err := base64.StdEncoding.DecodeString(solanaGTONHolderPrivKeyEncoded)
	if err != nil {
		return err
	}

	// solanaGTONHolderPrivKey := ed25519.NewKeyFromSeed(solanaGTONHolderPrivKey)
	solanaGTONHolderAccount := soltypes.AccountFromPrivateKeyBytes(decodedSolanaGTONHolderPrivKey)

	solanaPKPath := "./mvp-output/public_from-polygon-gton-recipient.json"

	/*
	 * Workaround specifically for Solana PK
	 */
	asNumberDecodedSolanaGTONHolderPrivKey := make([]int64, len(decodedSolanaGTONHolderPrivKey))
	for i, b := range decodedSolanaGTONHolderPrivKey {
		asNumberDecodedSolanaGTONHolderPrivKey[i] = int64(b)
	}

	marshalledSolanaPK, err := json.Marshal(asNumberDecodedSolanaGTONHolderPrivKey)
	if err != nil {
		return err
	}

	os.Mkdir("mvp-output", 0777)
	os.WriteFile(solanaPKPath, marshalledSolanaPK, 0777)

	// PublicKey  common.PublicKey
	// PrivateKey ed25519.PrivateKey
	solanaGTONHolder := &commands.OperatingAddress{
		Account:    solanaGTONHolderAccount,
		PublicKey:  solanaGTONHolderAccount.PublicKey,
		PrivateKey: base58.Encode(decodedSolanaGTONHolderPrivKey),
		PKPath:     solanaPKPath,
	}

	solanaGTONTokenAccountCreateResult := commands.CreateTokenAccountWithFeePayer(solanaGTONHolder.PKPath, gtonToken.cfg.destinationAddress)
	solanaGTONTokenAccount := solanaGTONTokenAccountCreateResult.TokenAccount

	luportClient, err := luport.NewLUPort(ethcommon.HexToAddress(extractorCfg.luportAddress), polygonClient)
	if err != nil {
		return err
	}

	randomFloat := func() float64 {
		return (rand.NormFloat64() + (float64(time.Now().Second()) / 60)) / 10
	}

	transferAmount := float64(int64(randomFloat()*1000)) / 1e6
	
	// gtonToken.Set(0.0000227)
	gtonToken.Set(transferAmount)

	fmt.Printf("As Origin: %v GTON \n", gtonToken.AsOriginBigInt())
	fmt.Printf("As Destination: %v GTON \n", gtonToken.AsDestinationBigInt())

	// // approve token spend
	gtonERC20, err := erc20.NewToken(ethcommon.HexToAddress(gtonToken.cfg.originAddress), polygonClient)
	if err != nil {
		return err
	}

	fmt.Printf("Approving %v GTON spend (Polygon) \n", gtonToken.Float())

	approveTx, err := gtonERC20.Approve(
		polygonTransactor.transactor,
		ethcommon.HexToAddress(extractorCfg.luportAddress),
		gtonToken.AsOriginBigInt(),
	)
	if err != nil {
		return err
	}

	fmt.Printf("Approve %v GTON spend tx (Polygon): %v \n", gtonToken.Float(), approveTx.Hash().Hex())

	time.Sleep(time.Second * 10)

	fmt.Printf("Locking %v GTON \n", gtonToken.Float())

	// (1)
	lockFundsTx, err := luportClient.CreateTransferUnwrapRequest(
		polygonTransactor.transactor,
		gtonToken.AsOriginBigInt(),
		solcommon.PublicKeyFromString(solanaGTONTokenAccount),
	)
	if err != nil {
		return err
	}

	fmt.Printf("Lock %v GTON tx (Polygon): %v \n", gtonToken.Float(), lockFundsTx.Hash().Hex())

	var solanaDepositAwaiter, polygonDepositAwaiter CrossChainTokenDepositAwaiter

	solanaEndpoint, _ := commands.InferSystemDefinedRPC()
	solanaDepositAwaiter = NewSolanaDepositAwaiter(solanaEndpoint)

	solanaDepositAwaiter.SetCfg(
		&CrossChainDepositAwaiterConfig{
			WatchAddress:    solanaGTONTokenAccount,
			WatchAssetID:    gtonToken.cfg.destinationAddress,
			WatchAmount:     gtonToken.AsDestinationBigInt(),
			PerAwaitTimeout: time.Second * 10,
		},
	)

	polygonDepositAwaiter = NewEVMExplorerClient()
	polygonDepositAwaiter.SetCfg(
		&CrossChainDepositAwaiterConfig{
			WatchAddress:    polygonGTONHolder.Address,
			WatchAssetID:    gtonToken.cfg.originAddress,
			WatchAmount:     gtonToken.AsOriginBigInt(),
			PerAwaitTimeout: time.Second * 10,
		},
	)

	polygonDepositBuffer := make(chan interface{})
	solanaDepositBuffer := make(chan interface{})

	go func() {
		err = solanaDepositAwaiter.AwaitTokenDeposit(solanaDepositBuffer)
	}()

	if err != nil {
		return err
	}

	for event := range solanaDepositBuffer {
		tokenDataState := event.(soltoken.TokenAccount)
		fmt.Printf("Deposit event (Solana): %+v \n", tokenDataState)
		break
	}

	// (2)
	ibportExecutor, err := commands.InitGenericExecutor(
		solanaGTONHolder.PKPath,
		extractorCfg.ibportProgramID,
		extractorCfg.ibportDataAccount,
		"",
		extractorCfg.destinationNodeURL,
		solcommon.PublicKeyFromString(""),
	)
	if err != nil {
		return err
	}

	err = commands.DelegateSPLTokenAmount(solanaGTONHolder.PKPath, gtonToken.cfg.destinationAddress, IBPortProgramPDA.ToBase58(), gtonToken.Float())
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 20)

	fmt.Printf("Approved %v GTON spend (Solana) \n", gtonToken.Float())

	polygonAddressDecoded, err := hexutil.Decode(polygonGTONHolder.Address)
	if err != nil {
		return err
	}

	var polygonTargetAddress [32]byte
	copy(polygonTargetAddress[:], polygonAddressDecoded)

	burnFundsResponse, err := ibportExecutor.BuildAndInvoke(
		executor.IBPortIXBuilder.CreateTransferUnwrapRequest(
			polygonTargetAddress,
			gtonToken.Float(),
		),
	)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 25)

	fmt.Printf("Lock %v GTON tx (Solana): %v \n", gtonToken.Float(), burnFundsResponse.TxSignature)

	// print
	// (3)
	go func() {
		err = polygonDepositAwaiter.AwaitTokenDeposit(solanaDepositBuffer)
	}()

	if err != nil {
		return err
	}

	for event := range polygonDepositBuffer {
		depositEvent := event.(*EVMTokenTransferEvent)
		fmt.Printf("Polygon - deposit event (burn from Solana) - TX: %+v \n", depositEvent.Hash)
		break
	}

	return nil
}
