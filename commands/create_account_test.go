package commands

import (
	"fmt"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/contract"
	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/portto/solana-go-sdk/types"
)



func TestCreateAccountForPDA(t *testing.T) {

	deployer, err := NewOperatingAddress(t, "../private-keys/_test-pk-deployer.json", nil)
	ValidateError(t, err)

	luportProgram, err := NewOperatingAddress(t, "../private-keys/_test_only_luport-program.json", &OperatingAddressBuilderOptions{
		WithPDASeeds: []byte(executor.CommonGravityBumpSeeds),
	})

	t.Logf("LU Port PDA: %v \n", luportProgram.PDA.ToBase58())
	ValidateError(t, err)

	WrappedFaucet(t, deployer.PKPath, "", 10)

	waitTransactionConfirmations()

	// RPCEndpoint, _ := InferSystemDefinedRPC()

	tokenDeployResult, err := CreateToken(deployer.PKPath)
	ValidateError(t, err)

	tokenMint := tokenDeployResult.Token

	fmt.Printf("Token deployed: %v \n", tokenDeployResult.Signature)

	deployerTokenAccount, err := CreateTokenAccount(deployer.PKPath, tokenMint.ToBase58())
	ValidateError(t, err)

	waitTransactionConfirmations()

	// mint some tokens for deployer
	err = MintToken(deployer.PKPath, tokenMint.ToBase58(), 1_000_000, deployerTokenAccount)
	ValidateError(t, err)
	t.Logf("Minted %v tokens to %v \n", 1_000_000, deployerTokenAccount)

	waitTransactionConfirmations()


	RPCEndpoint, _ := InferSystemDefinedRPC()

	deployerExecutor, err := executor.NewEmptyExecutor(deployer.PrivateKey, RPCEndpoint)
	ValidateError(t, err)

	ix, tokenAccount, err := contract.CreateAssociatedTokenAccountIX(deployer.PublicKey, luportProgram.PDA, tokenMint)
	ValidateError(t, err)

	response, err := deployerExecutor.InvokeIXList(
		[]types.Instruction { *ix },
	)
	ValidateError(t, err)

	fmt.Printf("LU Port PDA token account: %v; Tx: %v; \n", tokenAccount, response.TxSignature)
}

