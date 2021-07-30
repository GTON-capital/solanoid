package commands

import (
	"fmt"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
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

	/*
	 * The account problem
	 * The goal is to not require a signature
	 */
	//  tokenprog.
	// luportTokenAccountResponse, err := GenerateNewTokenAccount(
	// 	deployer.PrivateKey,
	// 	165, // alloc for token holder
	// 	// luportProgram.PublicKey also FAILS:
	// 	// send tx error, err: Transaction simulation failed: Error processing Instruction 1: instruction modified data of an account it does not own
	// 	luportProgram.PDA,
	// 	tokenMint,
	// 	RPCEndpoint,
	// 	"ibport",
	// )
	// ValidateError(t, err)

	// luportTokenAccount, err := CreateTokenAccount(luportProgram.PKPath, tokenMint.ToBase58())
	// ValidateError(t, err)

	// // luportTokenAccount := luportTokenAccountResponse.Account.PublicKey.ToBase58()

	waitTransactionConfirmations()

	// fmt.Printf("LU Port PDA token account: %v \n", luportTokenAccount)

	// err = TransferSPLTokensAllowUnfunded(deployer.PKPath, tokenMint.ToBase58(), luportProgram.PublicKey.ToBase58(), 1)
	err = TransferSPLTokensAllowUnfunded(deployer.PKPath, tokenMint.ToBase58(), luportProgram.PDA.ToBase58(), 1)
	ValidateError(t, err)

}