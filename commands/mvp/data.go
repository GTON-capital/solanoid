package mvp

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"runtime/debug"
	"time"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethhexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)



type evmKey struct {
	Address string
	PubKey  string
	PrivKey *ecdsa.PrivateKey
}

func newEVMKey(pk string) (*evmKey, error) {
	// key := ethcrypto.GenerateKey()
	decodedPK, err := ethcrypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}
	
	return &evmKey {
		Address: ethcrypto.PubkeyToAddress(decodedPK.PublicKey).String(),
		PubKey:  ethhexutil.Encode(ethcrypto.CompressPubkey(&decodedPK.PublicKey)),
		PrivKey: decodedPK,
	}, nil
}

type extractorCfg struct {
	originDecimals      int
	destinationDecimals int
	chainID             int64
	originNodeURL       string
	destinationNodeURL  string
	luportAddress       string
	ibportDataAccount   string
	ibportProgramID     string
}

type crossChainTokenCfg struct {
	originDecimals      int
	destinationDecimals int
	originAddress       string
	destinationAddress  string
}

func FloatToBigInt(val float64, decimals uint8) *big.Int {
    bigval := new(big.Float)
    bigval.SetFloat64(val)
    // Set precision if required.
    // bigval.SetPrec(64)

	multiplier := int64(math.Pow(10, float64(decimals)))

    coin := new(big.Float)
    coin.SetInt(big.NewInt(multiplier))

    bigval.Mul(bigval, coin)

    result := new(big.Int)
    bigval.Int(result) // store converted number in result

    return result
}

type tokenAmount struct {
	amount float64
}

func (ta *tokenAmount) Set(amount float64) {
	ta.amount = amount
}

func (ta *tokenAmount) PatchDecimals(decimals uint8) *big.Int {
	return FloatToBigInt(ta.amount, decimals)
}

type crossChainToken struct {
	token *tokenAmount
	cfg    *crossChainTokenCfg
}

func NewCrossChainToken(cfg *crossChainTokenCfg, amount float64) (*crossChainToken, error) {
	ccToken := &crossChainToken {
		token: &tokenAmount{ amount: amount },
		cfg: cfg,
	}

	return ccToken, nil
}

func (cct *crossChainToken) SetTokenCfg(cfg *crossChainTokenCfg) *crossChainToken {
	cct.cfg = cfg
	return cct
}

func (cct *crossChainToken) Set(amount float64) {
	cct.token.Set(amount)
}

func (cct *crossChainToken) Float() float64 {
	return cct.token.amount
}

func (cct *crossChainToken) AsOriginBigInt() *big.Int {
	return cct.token.PatchDecimals(uint8(cct.cfg.originDecimals))
}

func (cct *crossChainToken) AsDestinationBigInt() *big.Int {
	return cct.token.PatchDecimals(uint8(cct.cfg.destinationDecimals))
}

type CrossChainTokenDepositAwaiter interface {
	SetNotifier(func ()) error
}

type EVMDepositAwaiter struct {}

type SolanaDepositAwaiter struct {}


type EVMTransactor struct {
	ethClient  *ethclient.Client
	transactor *ethbind.TransactOpts
}

func NewEVMTransactor(ethClient *ethclient.Client, transactor *ethbind.TransactOpts) *EVMTransactor {
	return &EVMTransactor{
		ethClient:  ethClient,
		transactor: transactor,
	}
}

type EVMTokenTransferEvent struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}
type EVMTokenTransfersResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []EVMTokenTransferEvent `json:"result"`
}

type PolygonExplorerClient struct {
	awaitCheckTimeout time.Duration
	ctx    context.Context
	client http.Client
	apiKey string
}

func (pec *PolygonExplorerClient) DefaultNodeURL() string {
	return "https://api.polygonscan.com"
}

func (pec *PolygonExplorerClient) requestLastDeposits(watchAddress string, startBlock uint64) (*EVMTokenTransfersResult, error) {
	url := fmt.Sprintf(
		"%v/api?module=account&action=tokentx&address=%v&startblock=%v&sort=desc&apikey=%v",
		pec.DefaultNodeURL(),
		watchAddress,
		startBlock,
		pec.apiKey,
	)
	resp, err := http.DefaultClient.Get(
		url,
	)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result EVMTokenTransfersResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (pec *PolygonExplorerClient) awaitTokenDeposit(watchAddress, evmAssetId string, blockStart uint64, amount *big.Int, buf chan<- *EVMTokenTransferEvent) {
	// startOfAwait := time.Now()

	for {
		lastDeposits, err := pec.requestLastDeposits(watchAddress, blockStart)
		if err != nil {
			fmt.Printf("e: %v \n", err.Error())
			debug.PrintStack()
		}
		if len(lastDeposits.Result) > 0 {

			for _, event := range lastDeposits.Result {
				if event.ContractAddress != evmAssetId {
					continue
				}
				// if event.To != watchAddress {
				// 	continue
				// }
				if event.Value != amount.String() {
					continue
				}

				buf <- &event
				close(buf)
			}
		}

		time.Sleep(pec.awaitCheckTimeout)
	}
}

func (pec *PolygonExplorerClient) AwaitTokenDeposit(watchAddress, evmAssetId string, blockStart uint64, amount *big.Int, buf chan<- *EVMTokenTransferEvent) {
	pec.awaitTokenDeposit(watchAddress, evmAssetId, blockStart, amount, buf)
}

func (pec *PolygonExplorerClient) FetchLastTokenEvents(fromBlock int, targetAddress string) (*EVMTokenTransfersResult, error) {
	return nil, nil
}

func NewPolygonExplorerClient(awaitCheckTimeout time.Duration) *PolygonExplorerClient {
	return &PolygonExplorerClient{
		awaitCheckTimeout: awaitCheckTimeout,
		apiKey: "TU1S16Q38IJJA5A6SKZ2G6R84YHVTXITQK",
	}
}

