package mvp

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"time"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethhexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
	solclient "github.com/portto/solana-go-sdk/client"
	solcommon "github.com/portto/solana-go-sdk/common"
	soltoken "github.com/portto/solana-go-sdk/tokenprog"
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

type CrossChainDepositAwaiterConfig struct {
	WatchAddress    string
	WatchAssetID    string
	WatchAmount     *big.Int
	BlockStart      uint64
	PerAwaitTimeout time.Duration
}

type CrossChainTokenDepositAwaiter interface {
	AwaitTokenDeposit(chan <- interface{}) error
	SetCfg(*CrossChainDepositAwaiterConfig)
}

type EVMExplorerClient struct {
	apiKey string
	crossChainCfg *CrossChainDepositAwaiterConfig
}

func (eec *EVMExplorerClient) AwaitTokenDeposit(pipe chan <- interface{}) error {
	if eec.crossChainCfg == nil {
		return fmt.Errorf("cross chain cfg is not set")
	}

	for {
		deposits, err := eec.RequestLastDeposits(eec.crossChainCfg.WatchAddress, eec.crossChainCfg.BlockStart)
		fmt.Printf("deposits(len): %v \n", len(deposits.Result))
		if err != nil {
			return err
		}

		depositEvent := eec.AwaitDeposit(deposits, eec.crossChainCfg.WatchAddress, eec.crossChainCfg.WatchAssetID, eec.crossChainCfg.WatchAmount)

		if depositEvent == nil {
			time.Sleep(eec.crossChainCfg.PerAwaitTimeout)
			continue
		}

		var result interface{}
		result = depositEvent

		pipe <- result
		close(pipe)
		return nil
	}
}

func (eec *EVMExplorerClient) SetCfg(cfg *CrossChainDepositAwaiterConfig) {
	eec.crossChainCfg = cfg
}

func (eec *EVMExplorerClient) DefaultNodeURL() string {
	return "https://api.polygonscan.com"
}

func (eec *EVMExplorerClient) AwaitDeposit(deposits *EVMTokenTransfersResult, watchAddress, evmAssetId string, amount *big.Int) *EVMTokenTransferEvent {
	for _, event := range deposits.Result {
		if event.ContractAddress != evmAssetId {
			continue
		}
		if event.To != watchAddress {
			continue
		}
		if event.Value != amount.String() {
			continue
		}

		return &event
	}

	return nil
}

func (eec *EVMExplorerClient) RequestLastDeposits(watchAddress string, startBlock uint64) (*EVMTokenTransfersResult, error) {
	url := fmt.Sprintf(
		"%v/api?module=account&action=tokentx&address=%v&startblock=%v&sort=desc&apikey=%v",
		eec.DefaultNodeURL(),
		watchAddress,
		startBlock,
		eec.apiKey,
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

func NewEVMExplorerClient() *EVMExplorerClient {
	return &EVMExplorerClient{
		apiKey: "TU1S16Q38IJJA5A6SKZ2G6R84YHVTXITQK",
	}
}


type SolanaDepositAwaiter struct {
	nodeURL        string
	crossChainCfg *CrossChainDepositAwaiterConfig
	client        *solclient.Client
	ctx            context.Context
}

func NewSolanaDepositAwaiter(nodeURL string) *SolanaDepositAwaiter {
	client := solclient.NewClient(nodeURL)

	return &SolanaDepositAwaiter{ nodeURL: nodeURL, client: client, ctx: context.Background() }
}

func (sda *SolanaDepositAwaiter) SetCfg(cfg *CrossChainDepositAwaiterConfig) {
	sda.crossChainCfg = cfg
}

func (sda *SolanaDepositAwaiter) AwaitTokenDeposit(pipe chan <- interface{}) error {
	if sda.crossChainCfg == nil {
		return fmt.Errorf("cross chain cfg is not set")
	}

	var prevTokenAccountState *soltoken.TokenAccount

	for {
		tokenAccountState, err := sda.RequestTokenDataAccount()
		if err != nil {
			return err
		}

		if prevTokenAccountState == nil {
			prevTokenAccountState = tokenAccountState
			time.Sleep(sda.crossChainCfg.PerAwaitTimeout)
			continue
		}

		balanceDiff := tokenAccountState.Amount - prevTokenAccountState.Amount

		prevTokenAccountState = tokenAccountState

		if balanceDiff != sda.crossChainCfg.WatchAmount.Uint64() {
			time.Sleep(sda.crossChainCfg.PerAwaitTimeout)
			continue
		}

		var result interface{}
		result = *tokenAccountState

		pipe <- result
		close(pipe)
		return nil
	}
}

func (sda *SolanaDepositAwaiter) RequestTokenDataAccount() (*soltoken.TokenAccount, error) {
	stateResult, err := sda.client.GetAccountInfo(sda.ctx, sda.crossChainCfg.WatchAddress, solclient.GetAccountInfoConfig{
		Encoding: "base64",
	})
	
	if err != nil {
		return nil, err
	}

	if stateResult.Owner != solcommon.TokenProgramID.ToBase58() {
		return nil, fmt.Errorf("owner is not common.TokenProgramID")
	}

	tokenState := stateResult.Data.([]interface{})[0].(string)
	tokenStateDecoded, err := base64.StdEncoding.DecodeString(tokenState)
	if err != nil {
		return nil, err
	}

	tokenAccountState, err := soltoken.TokenAccountFromData(tokenStateDecoded)
	if err != nil {
		return nil, err
	}

	return tokenAccountState, nil
}