package gateway

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/Gravity-Tech/solanoid/commands"
	"github.com/Gravity-Tech/solanoid/commands/contract"
	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/Gravity-Tech/solanoid/models/nebula"
	"github.com/portto/solana-go-sdk/common"
)



func WaitTransactionConfirmations() {
	time.Sleep(time.Second * 30)
}

type GatewayDeployResult struct {}

func DeploySolanaGateway_LUPort(t *testing.T, consuls []string, originTokenMint common.PublicKey) *GatewayDeployResult {
	var err error

	deployer, err := commands.ReadOperatingAddress(t, "../../private-keys/mainnet/deployer.json")
	commands.ValidateError(t, err)

	balanceBeforeDeploy, err := commands.ReadAccountBalance(deployer.PublicKey.ToBase58())
	commands.ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL;  \n", balanceBeforeDeploy)

	nebulaProgram, err := commands.NewOperatingBinaryAddressFromString(
		contract.NebulaBinary,
		[]byte(executor.CommonGravityBumpSeeds),
	)
	commands.ValidateError(t, err)

	luportProgram, err := commands.NewOperatingBinaryAddressFromString(
		contract.LUPortBinary,
		[]byte(executor.CommonGravityBumpSeeds),
	)

	commands.ValidateError(t, err)
	fmt.Printf("LU Port Program ID: %v \n", luportProgram.PublicKey.ToBase58())
	fmt.Printf("LU Port PDA: %v \n", luportProgram.PDA.ToBase58())

	gravityDataAccount := contract.GravityDataAccount

	var consulsAsByteList []byte
	for _, consul := range consuls {
		consulsPubKey := common.PublicKeyFromString(consul)

		consulsAsByteList = append(consulsAsByteList, consulsPubKey.Bytes()...)
	}

	const BFT = 3

	RPCEndpoint, _ := commands.InferSystemDefinedRPC()

	WaitTransactionConfirmations()

	nebulaDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.NebulaAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Data Account: %v \n", nebulaDataAccount.Account.PublicKey.ToBase58())

	nebulaMultisigDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.MultisigAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Multisig Account: %v \n", nebulaMultisigDataAccount.Account.PublicKey.ToBase58())

	luportDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.LUPortAllocation, luportProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("LU Port Data Account: %v \n", luportProgram.Account.PublicKey.ToBase58())

	WaitTransactionConfirmations()
	
	nebulaBuilder := executor.NebulaInstructionBuilder{}
	nebulaExecutor, err := commands.InitGenericExecutor(
		deployer.PrivateKey,
		nebulaProgram.PublicKey.ToBase58(),
		nebulaDataAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigDataAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		common.PublicKeyFromString(gravityDataAccount),
	)
	commands.ValidateError(t, err)

	luportExecutor, err := commands.InitGenericExecutor(
		deployer.PrivateKey,
		luportProgram.PublicKey.ToBase58(),
		luportDataAccount.Account.PublicKey.ToBase58(),
		"",
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)
	commands.ValidateError(t, err)

	WaitTransactionConfirmations()

	nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Init(BFT, nebula.Bytes, common.PublicKeyFromString(gravityDataAccount), consulsAsByteList),
	)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Init: %v \n", nebulaInitResponse.TxSignature)

	WaitTransactionConfirmations()

	luportInitResult, err := luportExecutor.BuildAndInvoke(
		executor.LUPortIXBuilder.InitWithOracles(nebulaProgram.PublicKey, common.TokenProgramID, originTokenMint, BFT, consulsAsByteList),
	)

	fmt.Printf("LU Port Init: %v \n", luportInitResult.TxSignature)
	commands.ValidateError(t, err)

	WaitTransactionConfirmations()

	fmt.Println("LU Port Program is being subscribed to Nebula")

	var subID [16]byte
	rand.Read(subID[:])

	fmt.Printf("subID: %v \n", subID)

	// (4)
	nebulaSubscribePortResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(luportProgram.PDA, 1, 1, subID),
	)
	commands.ValidateError(t, err)

	fmt.Printf("Nebula Subscribe: %v \n", nebulaSubscribePortResponse.TxSignature)
	// fmt.Println("Now checking for valid double spend prevent")

	WaitTransactionConfirmations()

	balanceAfterDeploy, err := commands.ReadAccountBalance(deployer.PublicKey.ToBase58())
	commands.ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL; \n", balanceBeforeDeploy)
	fmt.Printf("balanceAfterDeploy: %v SOL; \n", balanceAfterDeploy)
	fmt.Printf("balance diff: %v SOL; \n", balanceBeforeDeploy-balanceAfterDeploy)

	return &GatewayDeployResult{}
}

func DeploySolanaGateway_IBPort(t *testing.T, consuls []string) {
	var err error

	deployer, err := commands.ReadOperatingAddress(t, "../../private-keys/mainnet/deployer.json")
	commands.ValidateError(t, err)

	balanceBeforeDeploy, err := commands.ReadAccountBalance(deployer.PublicKey.ToBase58())
	commands.ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL;  \n", balanceBeforeDeploy)

	nebulaProgram, err := commands.NewOperatingBinaryAddressFromString(
		contract.NebulaBinary,
		[]byte(executor.CommonGravityBumpSeeds),
	)
	commands.ValidateError(t, err)

	ibportProgram, err := commands.NewOperatingBinaryAddressFromString(
		contract.IBPortBinary,
		[]byte(executor.CommonGravityBumpSeeds),
	)
	commands.ValidateError(t, err)

	fmt.Printf("IB Port Program ID: %v \n", ibportProgram.PublicKey.ToBase58())
	fmt.Printf("IB Port PDA: %v \n", ibportProgram.PDA.ToBase58())

	// gravityDataAccount := "ErLEJcqRKQdhLpLHLn9zUzx1mu7VfrZbgwsfAL4BG4uQ"
	gravityDataAccount := contract.GravityDataAccount

	var consulsAsByteList []byte
	for _, consul := range consuls {
		consulsPubKey := common.PublicKeyFromString(consul)

		consulsAsByteList = append(consulsAsByteList, consulsPubKey.Bytes()...)
	}

	const BFT = 3

	RPCEndpoint, _ := commands.InferSystemDefinedRPC()

	tokenDeployResult, err := commands.CreateToken(deployer.PKPath)
	commands.ValidateError(t, err)
	
	tokenProgramAddress := tokenDeployResult.Token.ToBase58()
	fmt.Printf("token address: %v \n", tokenProgramAddress)

	WaitTransactionConfirmations()

	nebulaDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.NebulaAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Data Account: %v \n", nebulaDataAccount.Account.PublicKey.ToBase58())

	nebulaMultisigDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.MultisigAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Multisig Account: %v \n", nebulaMultisigDataAccount.Account.PublicKey.ToBase58())

	ibportDataAccount, err := commands.GenerateNewAccount(deployer.PrivateKey, commands.IBPortAllocation, ibportProgram.PublicKey.ToBase58(), RPCEndpoint)
	commands.ValidateError(t, err)
	fmt.Printf("IB Port Data Account: %v \n", ibportDataAccount.Account.PublicKey.ToBase58())

	WaitTransactionConfirmations()

	err = commands.AuthorizeToken(t, deployer.PKPath, tokenProgramAddress, "mint", ibportProgram.PDA.ToBase58())
	commands.ValidateError(t, err)
	t.Log("Authorizing IB Port to allow minting")

	nebulaExecutor, err := commands.InitGenericExecutor(
		deployer.PrivateKey,
		nebulaProgram.PublicKey.ToBase58(),
		nebulaDataAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigDataAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		common.PublicKeyFromString(gravityDataAccount),
	)
	commands.ValidateError(t, err)

	ibportBuilder := executor.IBPortInstructionBuilder{}
	ibportExecutor, err := commands.InitGenericExecutor(
		deployer.PrivateKey,
		ibportProgram.PublicKey.ToBase58(),
		ibportDataAccount.Account.PublicKey.ToBase58(),
		"",
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)
	commands.ValidateError(t, err)

	WaitTransactionConfirmations()

	nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(
		executor.NebulaIXBuilder.Init(BFT, nebula.Bytes, common.PublicKeyFromString(gravityDataAccount), consulsAsByteList),
	)
	commands.ValidateError(t, err)
	fmt.Printf("Nebula Init: %v \n", nebulaInitResponse.TxSignature)

	WaitTransactionConfirmations()

	ibportInitResult, err := ibportExecutor.BuildAndInvoke(
		ibportBuilder.InitWithOracles(nebulaProgram.PublicKey, common.TokenProgramID, tokenDeployResult.Token, BFT, consulsAsByteList),
	)

	fmt.Printf("IB Port Init: %v \n", ibportInitResult.TxSignature)
	commands.ValidateError(t, err)

	WaitTransactionConfirmations()

	fmt.Println("IB Port Program is being subscribed to Nebula")

	var subID [16]byte
	rand.Read(subID[:])

	fmt.Printf("subID: %v \n", subID)

	nebulaSubscribePortResponse, err := nebulaExecutor.BuildAndInvoke(
		executor.NebulaIXBuilder.Subscribe(ibportProgram.PDA, 1, 1, subID),
	)
	commands.ValidateError(t, err)

	fmt.Printf("Nebula Subscribe: %v \n", nebulaSubscribePortResponse.TxSignature)

	WaitTransactionConfirmations()

	balanceAfterDeploy, err := commands.ReadAccountBalance(deployer.PublicKey.ToBase58())
	commands.ValidateError(t, err)

	fmt.Printf("balanceBeforeDeploy: %v SOL; \n", balanceBeforeDeploy)
	fmt.Printf("balanceAfterDeploy: %v SOL; \n", balanceAfterDeploy)
	fmt.Printf("balance diff: %v SOL; \n", balanceBeforeDeploy-balanceAfterDeploy)
}
