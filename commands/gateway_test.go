package commands

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/Gravity-Tech/solanoid/models/nebula"

	"github.com/portto/solana-go-sdk/common"
)

func TestRunSolanaGatewayDeployment(t *testing.T) {
	var err error

	deployer, err := ReadOperatingAddress(t, "../private-keys/mainnet/deployer.json")
	ValidateError(t, err)

	// deployer, err := NewOperatingAddress(t, "../private-keys/mainnet/deployer.json",  &OperatingAddressBuilderOptions {
	// 	Overwrite: true,
	// })
	// ValidateError(t, err)

	mathWalletUser := "ANRHaW53Z89VWV5ycLr1HFW6dCTiLRj3RSiYNBBF8er1"

	_ = mathWalletUser
	// WrappedFaucet(t, deployer.PKPath, mathWalletUser, 10)
	// WrappedFaucet(t, deployer.PKPath, deployer.PublicKey.ToBase58(), 10)

	// waitTransactionConfirmations()

	balanceBeforeDeploy, err := ReadAccountBalance(deployer.PublicKey.ToBase58())
	ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL; \n", balanceBeforeDeploy)

	nebulaProgram, err := NewOperatingAddress(t, "../private-keys/mainnet/nebula.json", &OperatingAddressBuilderOptions{
		Overwrite: true,
	})
	ValidateError(t, err)
	fmt.Printf("Nebula Program ID: %v \n", nebulaProgram.Account.PublicKey.ToBase58())

	ibportProgram, err := NewOperatingAddress(t, "../private-keys/mainnet/ibport.json", &OperatingAddressBuilderOptions{
		Overwrite: true,
		WithPDASeeds: []byte(executor.IBPortPDABumpSeeds),
	})
	ValidateError(t, err)
	fmt.Printf("IB Port Program ID: %v \n", ibportProgram.Account.PublicKey.ToBase58())
	fmt.Printf("IB Port PDA: %v \n", ibportProgram.PDA.ToBase58())

	gravityProgramID := "3rDUA7AGseQn8VGjtwQ6NxqbrJq6z7Pmy9L8kQ9zXuhc"
	_ = gravityProgramID

	gravityDataAccount := "ErLEJcqRKQdhLpLHLn9zUzx1mu7VfrZbgwsfAL4BG4uQ"

	consuls := []string{
		"EnwGpvfZdCpkjs8jMShjo8evce2LbNfrYvREzdwGh5oc",
		"5Ng92o7CPPWk5tT2pqrnRMndoD49d51f4QcocgJttGHS",
		"ESgKDVemBdqDty6WExZ74kV8Re9yepth5tbKcsWTNXC9",
	}

	var consulsAsByteList []byte
	for _, consul := range consuls {
		consulsPubKey := common.PublicKeyFromString(consul)

		consulsAsByteList = append(consulsAsByteList, consulsPubKey.Bytes()...)
	}

	const BFT = 3

	RPCEndpoint, _ := InferSystemDefinedRPC()

	tokenDeployResult, err := CreateToken(deployer.PKPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()

	waitTransactionConfirmations()

	nebulaDataAccount, err := GenerateNewAccount(deployer.PrivateKey, NebulaAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)
	fmt.Printf("Nebula Data Account: %v \n", nebulaDataAccount.Account.PublicKey.ToBase58())

	nebulaMultisigAccount, err := GenerateNewAccount(deployer.PrivateKey, MultisigAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)
	fmt.Printf("Nebula Multisig Account: %v \n", nebulaMultisigAccount.Account.PublicKey.ToBase58())

	ibportDataAccount, err := GenerateNewAccount(deployer.PrivateKey, IBPortAllocation, ibportProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)
	fmt.Printf("IB Port Data Account: %v \n", ibportDataAccount.Account.PublicKey.ToBase58())

	waitTransactionConfirmations()

	_, err = DeploySolanaProgram(t, "ibport", ibportProgram.PKPath, deployer.PKPath, "../binaries/ibport.so")
	ValidateError(t, err)

	waitTransactionConfirmations()

	_, err = DeploySolanaProgram(t, "nebula", nebulaProgram.PKPath, deployer.PKPath, "../binaries/nebula.so")
	ValidateError(t, err)

	waitTransactionConfirmations()

	err = AuthorizeToken(t, deployer.PKPath, tokenProgramAddress, "mint", ibportProgram.PDA.ToBase58())
	ValidateError(t, err)
	t.Log("Authorizing IB Port to allow minting")

	nebulaBuilder := executor.NebulaInstructionBuilder{}
	nebulaExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		nebulaProgram.PublicKey.ToBase58(),
		nebulaDataAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		common.PublicKeyFromString(gravityDataAccount),
	)
	ValidateError(t, err)

	ibportBuilder := executor.IBPortInstructionBuilder{}
	ibportExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		ibportProgram.PublicKey.ToBase58(),
		ibportDataAccount.Account.PublicKey.ToBase58(),
		"",
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)
	ValidateError(t, err)

	waitTransactionConfirmations()

	nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Init(BFT, nebula.Bytes, common.PublicKeyFromString(gravityDataAccount), consulsAsByteList),
	)
	ValidateError(t, err)
	fmt.Printf("Nebula Init: %v \n", nebulaInitResponse.TxSignature)

	waitTransactionConfirmations()

	ibportInitResult, err := ibportExecutor.BuildAndInvoke(
		ibportBuilder.InitWithOracles(nebulaProgram.PublicKey, common.TokenProgramID, BFT, consulsAsByteList),
	)

	fmt.Printf("IB Port Init: %v \n", ibportInitResult.TxSignature)
	ValidateError(t, err)

	waitTransactionConfirmations()

	fmt.Println("IB Port Program is being subscribed to Nebula")

	var subID [16]byte
	rand.Read(subID[:])

	fmt.Printf("subID: %v \n", subID)

	// (4)
	nebulaSubscribePortResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(ibportProgram.PDA, 1, 1, subID),
	)
	ValidateError(t, err)

	fmt.Printf("Nebula Subscribe: %v \n", nebulaSubscribePortResponse.TxSignature)
	// fmt.Println("Now checking for valid double spend prevent")

	waitTransactionConfirmations()

	balanceAfterDeploy, err := ReadAccountBalance(deployer.PublicKey.ToBase58())
	ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL; \n", balanceBeforeDeploy)
	fmt.Printf("balanceAfterDeploy: %v SOL; \n", balanceAfterDeploy)
	fmt.Printf("balance diff: %v SOL; \n", balanceBeforeDeploy-balanceAfterDeploy)
}
