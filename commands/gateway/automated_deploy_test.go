package gateway

import (
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/contract"
	"github.com/portto/solana-go-sdk/common"
)

func TestDeploySolanaToEVMGateway(t *testing.T) {
	DeploySolanaGateway_LUPort(t, contract.SolanaGravityConsuls(), common.PublicKeyFromString(contract.RaydiumToken))
}

func TestDeploySolanaGateway_LUPort(t *testing.T) {
	// DeploySolanaGateway_LUPort(t, contract.SolanaGravityConsuls(), common.PublicKeyFromString(contract.RaydiumToken))
	DeploySolanaGateway_LUPort(t, contract.SolanaGravityConsuls(), common.PublicKeyFromString(contract.SerumToken))
}

func TestDeploySolanaGateway_IBPort(t *testing.T) {
	DeploySolanaGateway_IBPort(t, contract.SolanaGravityConsuls())
}
