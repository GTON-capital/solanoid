package commands

import (
	"fmt"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
)




func TestLUPortFullFlow(t *testing.T) {
	var err error
	// deployerPrivateKeysPath := "../private-keys/__test_deployer-pk-deployer.json"
	// tokenOwnerPath := "../private-keys/__test_only-token-owner.json"
	// luportProgramPath := "../private-keys/__test_only_luport-owner.json"

	deployer, err := NewOperatingAddress(t, "../private-keys/__test_deployer-pk-deployer.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
	})
	ValidateError(t, err)
	fmt.Printf("Deployer: %v \n", deployer.PublicKey.ToBase58())

	tokenOwner, err := NewOperatingAddress(t, "../private-keys/__test_only-token-owner.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
	})
	ValidateError(t, err)
	fmt.Printf("Token Owner: %v \n", tokenOwner.PublicKey.ToBase58())

	luportProgram, err := NewOperatingAddress(t, "../private-keys/mainnet/ibport.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
		WithPDASeeds: []byte(executor.LUPortPDABumpSeeds),
	})


	tokenOwnerAddress := tokenOwner.PublicKey.ToBase58()
	// tokenOwnerAddress, err := ReadAccountAddress(tokenOwnerPath)
	// ValidateError(t, err)

	// err = SystemFaucet(t, tokenOwnerAddress, 10)
	WrappedFaucet(t, tokenOwner.PKPath, tokenOwnerAddress, 10)
	ValidateError(t, err)
	WrappedFaucet(t, deployer.PKPath, deployer.PublicKey.ToBase58(), 10)
	ValidateError(t, err)

	tokenDeployResult, err := CreateToken(tokenOwner.PKPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()

	deployerTokenAccount, err := CreateTokenAccount(deployer.PKPath, tokenProgramAddress)
	ValidateError(t, err)

	ibportAddress := luportProgram.PublicKey.ToBase58()

	fmt.Printf("token  program address: %s\n", tokenProgramAddress)

	t.Logf("tokenProgramAddress: %v", tokenProgramAddress)
	t.Logf("deployerAddress: %v", deployer.PublicKey.ToBase58())
	t.Logf("tokenOwnerAddress: %v", tokenOwnerAddress)
	t.Logf("LU Port Address: %v", ibportAddress)
	t.Logf("LU Port PDA: %v", luportProgram.PDA.ToBase58())
	t.Logf("deployerTokenAccount: %v", deployerTokenAccount)

	// deployerPrivateKey, err := ReadPKFromPath(t, deployer.PK)
	// ValidateError(t, err)

	// SystemFaucet(t, deployerAddress, 10)
	// ValidateError(t, err)

	// love this *ucking timeouts
	waitTransactionConfirmations()

	_, err = DeploySolanaProgram(t, "luport", luportProgram.PKPath, deployer.PKPath, "../binaries/luport.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()

	portDataAccount, err := GenerateNewAccount(deployer.PrivateKey, LUPortAllocation, ibportAddress, endpoint)
	ValidateError(t, err)

}