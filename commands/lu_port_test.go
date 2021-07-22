package commands

import (
	"fmt"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

/*
 * Test flow:
 * 1. Init + validate
 * 2. Lock tokens.
 * 3. Validate failing cases of lock.
 * 4. Unlock tokens.
 * 5. Validate failing cases of unlock.
 */
func TestLUPortFullFlow(t *testing.T) {
	var err error
	// deployerPrivateKeysPath := "../private-keys/__test_deployer-pk-deployer.json"
	// tokenOwnerPath := "../private-keys/__test_only-token-owner.json"
	// luportProgramPath := "../private-keys/__test_only_luport-owner.json"

	deployer, err := NewOperatingAddress(t, "../private-keys/_test_deployer-pk-deployer.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
	})
	ValidateError(t, err)
	fmt.Printf("Deployer: %v \n", deployer.PublicKey.ToBase58())

	tokenOwner, err := NewOperatingAddress(t, "../private-keys/_test_only-token-owner.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
	})
	ValidateError(t, err)
	fmt.Printf("Token Owner: %v \n", tokenOwner.PublicKey.ToBase58())
	
	luportProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-luport.json", &OperatingAddressBuilderOptions{
		Overwrite:    true,
		WithPDASeeds: []byte(executor.LUPortPDABumpSeeds),
	})
	ValidateError(t, err)

	tokenOwnerAddress := tokenOwner.PublicKey.ToBase58()
	
	const BFT = 3

	// const OmitSendValueFlow = false

	consulsList, err := GenerateConsuls(t, "../private-keys/___test_consul_prefix_", BFT)
	ValidateError(t, err)

	// operatingConsul := consulsList.List[0]

	for i, consul := range append(consulsList.List, *deployer) {
		if i == BFT {
			WrappedFaucet(t, deployer.PKPath, "", 100)
		}

		WrappedFaucet(t, deployer.PKPath, consul.PublicKey.ToBase58(), 100)
	}

	WrappedFaucet(t, tokenOwner.PKPath, tokenOwnerAddress, 100)
	ValidateError(t, err)
	// WrappedFaucet(t, deployer.PKPath, deployer.PublicKey.ToBase58(), 1)
	// ValidateError(t, err)

	tokenDeployResult, err := CreateToken(tokenOwner.PKPath)
	ValidateError(t, err)

	waitTransactionConfirmations()

	tokenMint := tokenDeployResult.Token

	deployerTokenAccount, err := CreateTokenAccount(deployer.PKPath, tokenMint.ToBase58())
	ValidateError(t, err)
	
	luportTokenAccount, err := CreateTokenAccount(luportProgram.PKPath, tokenMint.ToBase58())
	ValidateError(t, err)

	luportAddress := luportProgram.PublicKey.ToBase58()

	fmt.Printf("token  program address: %s\n", tokenMint.ToBase58())

	t.Logf("tokenProgramAddress: %v", tokenMint.ToBase58())
	t.Logf("deployerAddress: %v", deployer.PublicKey.ToBase58())
	t.Logf("tokenOwnerAddress: %v", tokenOwnerAddress)
	t.Logf("LU Port Address: %v", luportAddress)
	t.Logf("LU Port PDA: %v", luportProgram.PDA.ToBase58())
	t.Logf("deployerTokenAccount: %v", deployerTokenAccount)

	err = MintToken(tokenOwner.PKPath, tokenMint.ToBase58(), 1_000_000, deployerTokenAccount)
	ValidateError(t, err)
	// err = MintToken(tokenOwner.PKPath, tokenMint.ToBase58(), 1_000_000, cons)
	// ValidateError(t, err)
	t.Log("Minted some tokens")
	
	waitTransactionConfirmations()

	_, err = DeploySolanaProgram(t, "luport", luportProgram.PKPath, deployer.PKPath, "../binaries/luport.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()

	portDataAccount, err := GenerateNewAccount(deployer.PrivateKey, LUPortAllocation, luportAddress, endpoint)
	ValidateError(t, err)


	luportExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		luportAddress,
		portDataAccount.Account.PublicKey.ToBase58(),
		"",
		endpoint,
		common.PublicKeyFromString(luportAddress),
	)
	ValidateError(t, err)

	mockedNebulaAddress := common.PublicKeyFromString(deployer.PublicKey.ToBase58())

	waitTransactionConfirmations()

	ibportInitResult, err := luportExecutor.BuildAndInvoke(
		executor.LUPortIXBuilder.InitWithOracles(mockedNebulaAddress, common.TokenProgramID, tokenMint, 3, consulsList.ConcatConsuls()),
	)
	ValidateError(t, err)
	t.Logf("LUPort Init: %v \n", ibportInitResult.TxSignature)

	luportExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
		{PubKey: tokenMint, IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(deployerTokenAccount), IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(luportTokenAccount), IsWritable: true, IsSigner: false},
	})

	evmReceiver := executor.RandomEVMAddress()
	lockAmounts := []float64 {
		1.235,
		0.4234,
	}

	waitTransactionConfirmations()

	lockTokens, err := luportExecutor.BuildAndInvoke(
		executor.LUPortIXBuilder.CreateTransferWrapRequest(evmReceiver, lockAmounts[0]),
	)
	ValidateError(t, err)
	t.Logf("LUPort #1 CreateTransferWrapRequest (%v): %v \n", lockAmounts[0], lockTokens.TxSignature)
	
	
	// dataHashForAttach := executor.BuildCrossChainMintByteVector(swapId, common.PublicKeyFromString(deployerTokenAccount), attachedAmount)

	// allow ibport to mint
	// err = AuthorizeToken(t, tokenOwner.PKPath, tokenProgramAddress, "mint", luportProgram.PDA.ToBase58())
	// ValidateError(t, err)
	// t.Log("Authorizing ib port to allow minting")
	// t.Log("Call attach value ")

	// waitTransactionConfirmations()

	// swapId := make([]byte, 16)
	// rand.Read(swapId)

	// t.Logf("Token Swap  Id: %v \n", swapId)

	// attachedAmount := float64(227)

	// t.Logf("15 - Float As Bytes: %v \n", executor.Float64ToBytes(attachedAmount))

	// dataHashForAttach := executor.BuildCrossChainMintByteVector(swapId, common.PublicKeyFromString(deployerTokenAccount), attachedAmount)

	// ibportCreateTransferUnwrapRequestResult, err := ibportExecutor.BuildAndInvoke(
	// 	instructionBuilder.AttachValue(dataHashForAttach),
	// )
	// ValidateError(t, err)

	// t.Logf("#1 AttachValue - Tx: %v \n", ibportCreateTransferUnwrapRequestResult.TxSignature)

	// t.Logf("Checking for double spend problem \n")

	// swapIdSecond := make([]byte, 16)
	// rand.Read(swapIdSecond)

	// dataHashForAttachSecond := executor.BuildCrossChainMintByteVector(swapIdSecond, common.PublicKeyFromString(deployerTokenAccount), attachedAmount)

	// waitTransactionConfirmations()

	// ibportCreateTransferUnwrapRequestResult, err = ibportExecutor.BuildAndInvoke(
	// 	instructionBuilder.AttachValue(dataHashForAttachSecond),
	// )
	// ValidateError(t, err)

	// t.Logf("#2 AttachValue - Tx:  %v \n", ibportCreateTransferUnwrapRequestResult.TxSignature)

	// waitTransactionConfirmations()

	// swapIdThird := make([]byte, 16)
	// rand.Read(swapIdThird)

	// dataHashForAttachThird := executor.BuildCrossChainMintByteVector(swapIdThird, common.PublicKeyFromString(deployerTokenAccount), attachedAmount)

	// waitTransactionConfirmations()

	// ibportCreateTransferUnwrapRequestResult, err = ibportExecutor.BuildAndInvoke(
	// 	instructionBuilder.AttachValue(dataHashForAttachThird),
	// )
	// ValidateError(t, err)

	// t.Logf("#3 AttachValue - Tx:  %v \n", ibportCreateTransferUnwrapRequestResult.TxSignature)

	// ibportCreateTransferUnwrapRequestResult, err = ibportExecutor.BuildAndInvoke(
	// 	instructionBuilder.AttachValue(dataHashForAttachThird),
	// )

	// if err != nil {
	// 	t.Logf("Program must fail with error 0x1 \n")
	// 	t.Logf("If so - double spend has been prevented \n")
	// }
}