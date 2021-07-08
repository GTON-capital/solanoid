package mvp

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	erc20 "github.com/Gravity-Tech/gateway/abi/ethereum/erc20"
	commands "github.com/Gravity-Tech/solanoid/commands"
	"github.com/Gravity-Tech/solanoid/commands/executor"

	luport "github.com/Gravity-Tech/gateway/abi/ethereum/luport"
	"github.com/portto/solana-go-sdk/common"
	solcommon "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)


func waitTransactionConfirmations() {
	// time.Sleep(time.Millisecond * 500)
	// time.Sleep(time.Second * 5)
	// time.Sleep(time.Second * 3) // the most safe timeout
	time.Sleep(time.Second * 30)
	// time.Sleep(time.Second * 25)
	// time.Sleep(time.Second * 45)
}

type MVPConfigMeta struct {
	SolanaGTONTokenRecever string
	PolygonGTONReceiver string
}

type MVPConfig struct { 
	Token *crossChainToken
	Extractor *extractorCfg
	Meta *MVPConfigMeta
}

func BuildMVPConfig() (*MVPConfig, error) {
	gtonToken, err := NewCrossChainToken(&crossChainTokenCfg{
		originDecimals:      18,
		destinationDecimals: 8,
		originAddress:       "0xf480f38c366daac4305dc484b2ad7a496ff00cea",
		destinationAddress:  "nVZnRKdr3pmcgnJvYDE8iafgiMiBqxiffQMcyv5ETdA",
	}, 0)
	if err != nil {
		return nil, err
	}

	extractorCfg := &extractorCfg{
		originDecimals:      18,
		destinationDecimals: 8,
		chainID:             137,
		originNodeURL:       "https://rpc-mainnet.maticvigil.com",
		destinationNodeURL: "https://api.mainnet-beta.solana.com",
		luportAddress:      "0x7725d618122F9A2Ce368dA1624Fbc79ce197c438",
		ibportDataAccount:  "9kwBfNbrQAEmEqkZbvMCKkefuJBj7nuqWrq6dzUhW5fJ",
		ibportProgramID:    "AH3QKaj942UUxDjaRaGh7hvdadsD8yfU9LRTa9KXfJkZ",
	}

	return &MVPConfig{
		Token: gtonToken,
		Extractor: extractorCfg,
		Meta: &MVPConfigMeta {
			SolanaGTONTokenRecever: "FMtjwGs2V6j3eWvZhLA18tkHuzvBHfpjFcCuuvsweuwC",
			PolygonGTONReceiver: "0xBbc3D3F8C70C1A558bD0B5C25662aa3226b863e9",
		},
	}, nil
}

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
	cfg, err := BuildMVPConfig()
	commands.ValidateError(t, err)
	gtonToken, extractorCfg := cfg.Token, cfg.Extractor

	polygonCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	polygonClient, err := ethclient.DialContext(polygonCtx, extractorCfg.originNodeURL)
	commands.ValidateError(t, err)

	polygonGTONHolder, err := newEVMKey("76b77e0673cdf31a4bbfa0f0cdd9ed1fe02f036697d91dbf6293767f630e3b47")
	commands.ValidateError(t, err)

	transactor, err := ethbind.NewKeyedTransactorWithChainID(polygonGTONHolder.PrivKey, big.NewInt(extractorCfg.chainID))
	transactor.GasLimit = 10 * 150000
	transactor.Context = polygonCtx

	commands.ValidateError(t, err)

	polygonTransactor := NewEVMTransactor(polygonClient, transactor)

	solanaGTONHolder, err := commands.ReadOperatingAddress(t, "../../public/from-polygon-gton-recipient.json")
	commands.ValidateError(t, err)
	_ = solanaGTONHolder

	/*
		> spl-token create-account --fee-payer private-keys/_from-polygon-gton-mvp-recipient.json nVZnRKdr3pmcgnJvYDE8iafgiMiBqxiffQMcyv5ETdA
		  Creating account FMtjwGs2V6j3eWvZhLA18tkHuzvBHfpjFcCuuvsweuwC
		  Signature: 3ojYtfDofzBSNWrRPRSdm3Nz9iBZqbbsV3rJBbpjrW56CPpzFYkj8K8XvnZT284Va6VGq9uqEUiv5yHhpY9HERBM
	*/
	// solanaGTONTokenAccount := "FMtjwGs2V6j3eWvZhLA18tkHuzvBHfpjFcCuuvsweuwC"
	solanaGTONTokenAccountCreateResult := commands.CreateTokenAccountWithFeePayer(solanaGTONHolder.PKPath, gtonToken.cfg.destinationAddress)
	solanaGTONTokenAccount := solanaGTONTokenAccountCreateResult.TokenAccount

	fmt.Printf("solanaGTONTokenAccount: %v \n", solanaGTONTokenAccountCreateResult.TokenAccount)
	// fmt.Printf("solanaGTONTokenAccount(err): %v \n", solanaGTONTokenAccountCreateResult.Error)

	luportClient, err := luport.NewLUPort(ethcommon.HexToAddress(extractorCfg.luportAddress), polygonClient)
	commands.ValidateError(t, err)

	transferAmount := float64(int64(rand.Float64()*1000)) / 1e6

	// gtonToken.Set(0.0000227)
	gtonToken.Set(transferAmount)

	fmt.Printf("As Origin: %v GTON \n", gtonToken.AsOriginBigInt())
	fmt.Printf("As Destination: %v GTON \n", gtonToken.AsDestinationBigInt())

	// approve token spend
	gtonERC20, err := erc20.NewToken(ethcommon.HexToAddress(gtonToken.cfg.originAddress), polygonClient)
	commands.ValidateError(t, err)

	fmt.Printf("Approving %v GTON spend \n", gtonToken.Float())
	approveTx, err := gtonERC20.Approve(
		polygonTransactor.transactor,
		ethcommon.HexToAddress(extractorCfg.luportAddress),
		gtonToken.AsOriginBigInt(),
	)
	commands.ValidateError(t, err)

	t.Logf("Approve %v GTON spend tx (Polygon): %v \n", gtonToken.Float(), approveTx.Hash().Hex())

	time.Sleep(time.Second * 3)

	fmt.Printf("Locking %v GTON \n", gtonToken.Float())

	// (1)
	lockFundsTx, err := luportClient.CreateTransferUnwrapRequest(
		polygonTransactor.transactor,
		gtonToken.AsOriginBigInt(),
		solcommon.PublicKeyFromString(solanaGTONTokenAccount),
	)
	commands.ValidateError(t, err)

	t.Logf("Lock %v GTON tx (Polygon): %v \n", gtonToken.Float(), lockFundsTx.Hash().Hex())

	// print
	// (3)

	// burn
	// (4)
	ibportBuilder := executor.IBPortInstructionBuilder{}
	ibportExecutor, err := commands.InitGenericExecutor(
		solanaGTONHolder.PKPath,
		extractorCfg.ibportProgramID,
		extractorCfg.ibportDataAccount,
		"",
		extractorCfg.destinationNodeURL,
		common.PublicKeyFromString(""),
	)
	commands.ValidateError(t, err)

	polygonAddressDecoded, err := hexutil.Decode(polygonGTONHolder.Address)
	commands.ValidateError(t, err)

	var polygonTargetAddress [32]byte
	copy(polygonTargetAddress[:], polygonAddressDecoded)

	burnFundsResponse, err := ibportExecutor.BuildAndInvoke(
		ibportBuilder.CreateTransferUnwrapRequest(
			polygonTargetAddress,
			gtonToken.Float(),
		),
	)
	commands.ValidateError(t, err)

	time.Sleep(time.Second * 15)

	// print
	// (5)
	t.Logf("Burn %v GTON tx (Solana): %v \n", gtonToken.Float(), burnFundsResponse.TxSignature)

	// awaiting unlock

	t.Logf("Awaiting unlock on Polygon... \n")

	// result
}

func TestBurn_PolygonSolana(t *testing.T) {
	cfg, err := BuildMVPConfig()
	commands.ValidateError(t, err)
	gtonToken, extractorCfg, meta := cfg.Token, cfg.Extractor, cfg.Meta

	ibPortPDA, err := common.CreateProgramAddress([][]byte{[]byte(executor.IBPortPDABumpSeeds)}, common.PublicKeyFromString(extractorCfg.ibportProgramID))
	commands.ValidateError(t, err)

	polygonGTONReceiver := ethcommon.HexToAddress(meta.PolygonGTONReceiver)

	solanaGTONHolder, err := commands.ReadOperatingAddress(t, "../../public/from-polygon-gton-recipient.json")
	commands.ValidateError(t, err)
	_ = solanaGTONHolder

	// solanaGTONTokenAccount := "FMtjwGs2V6j3eWvZhLA18tkHuzvBHfpjFcCuuvsweuwC"
	solanaGTONTokenAccount := meta.SolanaGTONTokenRecever

	burnAmount := 0.0001

	_ = solanaGTONTokenAccount
	// delegate amount to port BINARY for burning and request creation
	err = commands.DelegateSPLTokenAmount(solanaGTONHolder.PKPath, solanaGTONTokenAccount, ibPortPDA.ToBase58(), burnAmount)
	commands.ValidateError(t, err)

	t.Log("Delegated some tokens to ibport from  deployer")
	t.Log("Creating cross chain transfer tx")

	// executor.waitTransactionConfirmations()
	waitTransactionConfirmations()

	// ethReceiverPK, err := ethcrypto.GenerateKey()
	// commands.ValidateError(t, err)

	var ethReceiverAddress [32]byte
	copy(ethReceiverAddress[:], polygonGTONReceiver.Bytes())

	t.Logf("#1 EVM Receiver: %v \n", polygonGTONReceiver.String())
	// t.Logf("#1 EVM Receiver (bytes): %v \n", ethReceiverAddress[:])

	// ibportBuilder := executor.IBPortInstructionBuilder{}

	ibportExecutor, err := commands.InitGenericExecutor(
		solanaGTONHolder.PrivateKey,
		extractorCfg.ibportProgramID,
		extractorCfg.ibportDataAccount,
		"",
		extractorCfg.destinationNodeURL,
		common.PublicKeyFromString(""),
	)
	commands.ValidateError(t, err)


	ibportExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
		{PubKey: common.PublicKeyFromString(gtonToken.cfg.destinationAddress), IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(solanaGTONTokenAccount), IsWritable: true, IsSigner: false},
		{PubKey: ibPortPDA, IsWritable: false, IsSigner: false},
	})

	ibportCreateTransferUnwrapRequestResult, err := ibportExecutor.BuildAndInvoke(
		executor.IBPortIXBuilder.CreateTransferUnwrapRequest(ethReceiverAddress, burnAmount),
	)
	commands.ValidateError(t, err)

	t.Logf("#1 CreateTransferUnwrapRequest - Tx: %v \n", ibportCreateTransferUnwrapRequestResult.TxSignature)

	// commands.ValidateError(t, err)
}
