package commands

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

type OperatingAddressBuilderOptions struct {
	WithPDASeeds []byte
	Overwrite    bool
}

type OperatingAddress struct {
	Account    types.Account
	PublicKey  common.PublicKey
	PDA        common.PublicKey
	PrivateKey string
	PKPath     string
}

func ReadOperatingAddress(t *testing.T, path string) (*OperatingAddress, error) {
	pubkey, err := ReadAccountAddress(path)
	if err != nil {
		return nil, err
	}

	privateKey, err := ReadPKFromPath(t, path)
	if err != nil {
		return nil, err
	}

	decodedPrivKey, err := base58.Decode(privateKey)
	if err != nil {
		return nil, err
	}

	address := &OperatingAddress{
		Account:    types.AccountFromPrivateKeyBytes(decodedPrivKey),
		PublicKey:  common.PublicKeyFromString(pubkey),
		PrivateKey: privateKey,
		PKPath:     path,
	}

	return address, nil
}

func NewOperatingBinaryAddressFromString(binary string, seeds []byte) (*OperatingAddress, error) {
	var err error
	var targetAddressPDA common.PublicKey

	if len(seeds) > 0 {
		targetAddressPDA, err = common.CreateProgramAddress([][]byte{ seeds[:] }, common.PublicKeyFromString(binary))
		if err != nil {
			return nil, err
		}
	}

	return &OperatingAddress{
		PublicKey: common.PublicKeyFromString(binary),
		PDA: targetAddressPDA,
	}, nil
}

func NewOperatingAddress(t *testing.T, path string, options *OperatingAddressBuilderOptions) (*OperatingAddress, error) {
	var err error

	if options != nil && len(options.WithPDASeeds) > 0 {
		publicKey, pda, err := CreatePersistentAccountWithPDA(path, true, [][]byte{options.WithPDASeeds})
		if err != nil {
			return nil, err
		}

		privateKey, err := ReadPKFromPath(t, path)
		if err != nil {
			return nil, err
		}

		return &OperatingAddress{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
			PKPath:     path,
			PDA:        pda,
		}, nil
	}

	if options != nil && !options.Overwrite {
		err = CreatePersistedAccount(path, false)
		if err != nil {
			return nil, err
		}
	} else {
		err = CreatePersistedAccount(path, true)
		if err != nil {
			return nil, err
		}
	}

	pubkey, err := ReadAccountAddress(path)
	if err != nil {
		return nil, err
	}

	privateKey, err := ReadPKFromPath(t, path)
	if err != nil {
		return nil, err
	}

	decodedPrivKey, err := base58.Decode(privateKey)
	if err != nil {
		return nil, err
	}

	address := &OperatingAddress{
		Account:    types.AccountFromPrivateKeyBytes(decodedPrivKey),
		PublicKey:  common.PublicKeyFromString(pubkey),
		PrivateKey: privateKey,
		PKPath:     path,
	}

	return address, nil
}

type ConsulsHandler struct {
	BFT  uint8
	List []OperatingAddress
}

func (ch *ConsulsHandler) ConcatConsuls() []byte {
	var oracles []byte
	for _, consul := range ch.List {
		oracles = append(oracles, consul.PublicKey.Bytes()...)
	}

	return oracles
}

func (ch *ConsulsHandler) ToBftSigners() []executor.GravityBftSigner {
	var signers []executor.GravityBftSigner

	for _, signer := range ch.List {
		signers = append(signers, *executor.NewGravityBftSigner(signer.PrivateKey))
	}

	return signers
}

func GenerateConsuls(t *testing.T, consulPathPrefix string, count uint8) (*ConsulsHandler, error) {
	result := make([]OperatingAddress, count)

	var i uint8

	for i < count {
		path := fmt.Sprintf("%v_%v.json", consulPathPrefix, i)

		address, err := NewOperatingAddress(t, path, nil)

		if err != nil {
			return nil, err
		}
		result[i] = *address

		i++
	}

	return &ConsulsHandler{
		BFT:  count,
		List: result,
	}, nil
}

func ParallelExecution(callbacks []func()) {
	var wg sync.WaitGroup

	wg.Add(len(callbacks))
	for _, fn := range callbacks {
		// aliasing
		fn := fn
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	wg.Wait()
}
