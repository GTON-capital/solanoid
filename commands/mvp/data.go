package mvp

import (
	"crypto/ecdsa"
	"math"
	"math/big"

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

type EVMSolanaDepositAwaiter struct {}


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
