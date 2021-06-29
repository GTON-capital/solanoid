package mvp

import (
	"crypto/ecdsa"
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
	ibportAddress       string
}

type crossChainTokenCfg struct {
	originDecimals      int
	destinationDecimals int
	originAddress       string
	destinationAddress  string
}


type crossChainToken struct {
	amount *big.Int	
	cfg    *crossChainTokenCfg
}

func (cct *crossChainToken) SetTokenCfg(cfg *crossChainTokenCfg) *crossChainToken {
	cct.cfg = cfg
	return cct
}

func (cct *crossChainToken) SetAsBigInt(input *big.Int) *crossChainToken {
	cct.amount = input
	return cct
}

func (cct *crossChainToken) SetAsFloat(input float64, decimals uint8) *crossChainToken {
	cct.amount = big.NewInt(0).Mul(
		big.NewInt(int64(input)),
		big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil),
	)
	return cct
}

func (cct *crossChainToken) BigInt() *big.Int {
	return cct.amount
}

type CrossChainTokenDepositAwaiter interface {
	SetNotifier(func ()) error
}

type BalanceAwaiter struct {

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
