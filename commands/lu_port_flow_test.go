package commands

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"

	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/Gravity-Tech/solanoid/models/nebula"

	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

/*
 * Test logical steps
 *
 * 1. Deploy Nebula
 * 2. Init Nebula
 * 3. Deploy Port
 * 4. Subscribe Port to Nebula
 * 5. Call mocked attach data.
 *
 * Goals:
 * 1. Validate lock/unlock flow.
 * 2. Validate oracle multisig. (with various bft*)
 * 3. Validate double spend on attach
 * 4. Validate the atomic call: nebula.send_value_to_subs() -> nebula.attach()
 */
func TestNebulaSendValueToLUPortSubscriber(t *testing.T) {
	var err error

	deployer, err := NewOperatingAddress(t, "../private-keys/_test_deployer-pk-deployer.json", nil)
	ValidateError(t, err)

	gravityProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-gravity-program.json", nil)
	ValidateError(t, err)

	nebulaProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-nebula-program.json", &OperatingAddressBuilderOptions{
		WithPDASeeds: []byte(executor.CommonGravityBumpSeeds),
	})
	ValidateError(t, err)

	luportProgram, err := NewOperatingAddress(t, "../private-keys/_test_only_luport-program.json", &OperatingAddressBuilderOptions{
		WithPDASeeds: []byte(executor.CommonGravityBumpSeeds),
	})

	t.Logf("LU Port PDA: %v \n", luportProgram.PDA.ToBase58())
	ValidateError(t, err)

	const BFT = 3

	const OmitSendValueFlow = false

	consulsList, err := GenerateConsuls(t, "../private-keys/_test_consul_prefix_", BFT)
	ValidateError(t, err)

	operatingConsul := consulsList.List[0]

	for i, consul := range append(consulsList.List, *deployer) {
		if i == BFT {
			WrappedFaucet(t, deployer.PKPath, "", 10)
		}

		WrappedFaucet(t, deployer.PKPath, consul.PublicKey.ToBase58(), 10)
	}

	waitTransactionConfirmations()

	RPCEndpoint, _ := InferSystemDefinedRPC()

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
	t.Log("Minted some tokens")

	gravityDataAccount, err := GenerateNewAccount(deployer.PrivateKey, GravityContractAllocation, gravityProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	gravityMultisigAccount, err := GenerateNewAccount(deployer.PrivateKey, MultisigAllocation, gravityProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	nebulaDataAccount, err := GenerateNewAccount(deployer.PrivateKey, NebulaAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	nebulaMultisigAccount, err := GenerateNewAccount(deployer.PrivateKey, MultisigAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	luportDataAccount, err := GenerateNewAccount(deployer.PrivateKey, LUPortAllocation, luportProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	luportTokenAccountResponse, err := GenerateNewTokenAccount(
		deployer.PrivateKey,
		165,
		luportProgram.PublicKey,
		tokenMint,
		RPCEndpoint,
		"ibport",
	)
	ValidateError(t, err)

	luportTokenAccount := luportTokenAccountResponse.Account.PublicKey.ToBase58()

	waitTransactionConfirmations()

	ParallelExecution(
		[]func(){
			func() {
				_, err = DeploySolanaProgram(t, "luport", luportProgram.PKPath, consulsList.List[0].PKPath, "../binaries/luport.so")
				ValidateError(t, err)
			},
			func() {
				_, err = DeploySolanaProgram(t, "gravity", gravityProgram.PKPath, consulsList.List[1].PKPath, "../binaries/gravity.so")
				ValidateError(t, err)
			},
			func() {
				_, err = DeploySolanaProgram(t, "nebula", nebulaProgram.PKPath, consulsList.List[2].PKPath, "../binaries/nebula.so")
				ValidateError(t, err)
			},
		},
	)

	waitTransactionConfirmations()

	gravityBuilder := executor.GravityInstructionBuilder{}
	gravityExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		gravityProgram.PublicKey.ToBase58(),
		gravityDataAccount.Account.PublicKey.ToBase58(),
		gravityMultisigAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)

	nebulaBuilder := executor.NebulaInstructionBuilder{}
	nebulaExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		nebulaProgram.PublicKey.ToBase58(),
		nebulaDataAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		gravityDataAccount.Account.PublicKey,
	)
	ValidateError(t, err)

	luportExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		luportProgram.PublicKey.ToBase58(),
		luportDataAccount.Account.PublicKey.ToBase58(),
		"",
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)
	ValidateError(t, err)

	oracles := consulsList.ConcatConsuls()

	waitTransactionConfirmations()

	ParallelExecution(
		[]func(){
			func() {
				gravityInitResponse, err := gravityExecutor.BuildAndInvoke(
					gravityBuilder.Init(BFT, 1, oracles),
				)
				fmt.Printf("Gravity Init: %v \n", gravityInitResponse.TxSignature)
				ValidateError(t, err)
			},
			func() {
				// (2)
				nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(
					nebulaBuilder.Init(BFT, nebula.Bytes, gravityDataAccount.Account.PublicKey, oracles),
				)
				ValidateError(t, err)
				fmt.Printf("Nebula Init: %v \n", nebulaInitResponse.TxSignature)
			},
			func() {
				luportInitResult, err := luportExecutor.BuildAndInvoke(
					executor.LUPortIXBuilder.InitWithOracles(nebulaProgram.PublicKey, common.TokenProgramID, tokenDeployResult.Token, BFT, consulsList.ConcatConsuls()),
				)

				fmt.Printf("LU Port Init: %v \n", luportInitResult.TxSignature)
				ValidateError(t, err)
			},
		},
	)

	waitTransactionConfirmations()

	fmt.Println("LU Port Program is being subscribed to Nebula")

	var subID [16]byte
	rand.Read(subID[:])

	fmt.Printf("subID: %v \n", subID)

	// (4)
	nebulaSubscribePortResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(luportProgram.PDA, 1, 1, subID),
	)
	ValidateError(t, err)

	fmt.Printf("Nebula Subscribe: %v \n", nebulaSubscribePortResponse.TxSignature)
	fmt.Println("Now checking for valid double spend prevent")

	waitTransactionConfirmations()
	// waitTransactionConfirmations()

	_, err = nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(luportProgram.PDA, 1, 1, subID),
	)
	ValidateErrorExistence(t, err)

	fmt.Printf("Nebula Subscribe with the same subID must have failed: %v \n", err.Error())

	waitTransactionConfirmations()

	// luportTokenAccount, _ := CreateTokenAccount(luportProgram.PKPath, tokenMint.ToBase58())

	
	fmt.Printf("LU Port Token Account: %v \n", luportTokenAccount)
	// luportTokenAccount, err :=  CreateTokenAccount(luportProgram.PKPath, tokenMint.ToBase58())

	luportExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
		{PubKey: tokenMint, IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(deployerTokenAccount), IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(luportTokenAccount), IsWritable: true, IsSigner: false},
	})

	evmReceiver20bytes := executor.RandomEVMAddress()
	var evmReceiver32bytes [32]byte
	copy(evmReceiver32bytes[:], evmReceiver20bytes[:])

	lockAmounts := []float64 {
		1.235,
	}

	waitTransactionConfirmations()

	lockTokens, err := luportExecutor.BuildAndInvoke(
		executor.LUPortIXBuilder.CreateTransferWrapRequest(evmReceiver32bytes, lockAmounts[0]),
	)
	ValidateError(t, err)
	t.Logf("LUPort #1 CreateTransferWrapRequest (%v): %v \n", lockAmounts[0], lockTokens.TxSignature)
	

	attachValue := func(i, pulseID int, nebulaExecutor *executor.GenericExecutor,
		operator types.Account, rawDataValue []byte, additionalSigners []executor.GravityBftSigner,
		tokensReceiverDataAccount common.PublicKey,
	) error {
		swapId := make([]byte, 16)
		rand.Read(swapId)


		var rawDataValue64bytes [64]byte
		copy(rawDataValue64bytes[:], rawDataValue)

		var dataHashForAttach [32]byte

		hashingFunction := func(input []byte) []byte {
			digest := sha256.Sum256(input)
			return digest[:]
		}

		// copy(dataHashForAttach[:], ethcrypto.Keccak256(rawDataValue[:]))
		copy(dataHashForAttach[:], hashingFunction(rawDataValue[:]))
	
		fmt.Printf("Iteration #%v \n", i)
		fmt.Printf("Raw Data Value: %v \n", rawDataValue)
		fmt.Printf("Data Value Hash: %v \n", dataHashForAttach)

		nebulaExecutor.EraseAdditionalMeta()

		// nebulaExecutor.SetAdditionalSigners(consulsList.ToBftSigners())
		if len(additionalSigners) != 0 {
			nebulaExecutor.SetAdditionalSigners(additionalSigners)
		}

		nebulaExecutor.SetDeployerPK(operator)

		waitTransactionConfirmations()

		nebulaSendHashValueResponse, err := nebulaExecutor.BuildAndInvoke(
			executor.NebulaIXBuilder.SendHashValue(dataHashForAttach),
		)
		if err != nil {
			return err
		}

		fmt.Printf("#%v Nebula SendHashValue Call: %v \n", i, nebulaSendHashValueResponse.TxSignature)

		nebulaExecutor.EraseAdditionalSigners()
		nebulaExecutor.SetDeployerPK(operator)

		nebulaExecutor.SetAdditionalMeta([]types.AccountMeta{
			{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
			{PubKey: luportProgram.PublicKey, IsWritable: false, IsSigner: false},
			{PubKey: luportDataAccount.Account.PublicKey, IsWritable: true, IsSigner: false},
			{PubKey: tokenMint, IsWritable: true, IsSigner: false},
			{PubKey: tokensReceiverDataAccount, IsWritable: true, IsSigner: false},
			{PubKey: luportProgram.PDA, IsWritable: false, IsSigner: false},
			{PubKey: common.PublicKeyFromString(luportTokenAccount), IsWritable: true, IsSigner: false},
		})

		waitTransactionConfirmations()

		nebulaAttachResponse, err := nebulaExecutor.BuildAndInvoke(
			nebulaBuilder.SendValueToSubs(rawDataValue64bytes, nebula.Bytes, uint64(pulseID), subID),
		)
		if err != nil {
			return err
		}

		fmt.Printf("#%v Nebula SendValueToSubs Call:  %v \n", i, nebulaAttachResponse.TxSignature)

		waitTransactionConfirmations()
		return nil
	}


	baBuilder := executor.SolanaToEVMBABuilder{
		Amount: lockAmounts[0],
		Origin: common.PublicKeyFromString(deployerTokenAccount),
	}
	baBuilder.SetCfg(executor.BACfg{
		OriginDecimals: 8,
		DestDecimals: 18,
	})

	failingDataHashForAttach := baBuilder.BuildForReverse()

	baBuilder.Receiver = evmReceiver20bytes
	correctDataHashForAttach := baBuilder.BuildForReverse()

	
	hashQueue := [][]byte { failingDataHashForAttach, correctDataHashForAttach }

	// Should fail - caller - is not a gravity oracle
	for i := 0; i < len(hashQueue); i++ {
		err = attachValue(i, i, nebulaExecutor, deployer.Account, hashQueue[i], make([]executor.GravityBftSigner, 0), common.PublicKeyFromString(deployerTokenAccount))
		ValidateErrorExistence(t, err)
	}

	// Should pass valid
	for i := 0; i < len(hashQueue); i++ {
		err = attachValue(i, i, nebulaExecutor, operatingConsul.Account, hashQueue[i], consulsList.ToBftSigners(), common.PublicKeyFromString(deployerTokenAccount),)

		if i == 0 {
			ValidateErrorExistence(t, err)
		}
		ValidateError(t, err)
	}

}
