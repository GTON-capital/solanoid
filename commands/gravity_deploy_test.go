package commands

import (
	"fmt"
	"solanoid/commands/executor"
	"time"

	// "os"

	"testing"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	// "strings"
)

func TestGravityContract(t *testing.T) {
	var err, errFailing error

	deployerPrivateKeyPath := "../private-keys/main-deployer.json"
	// deployerPrivateKeyPath := "/Users/shamil/.config/solana/id.json"
	deployerPrivateKey, err := ReadPKFromPath(t, deployerPrivateKeyPath)
	ValidateError(t, err)

	deployerAddress, err := ReadAccountAddress(deployerPrivateKeyPath)
	ValidateError(t, err)

	initialBalance, err := ReadAccountBalance(deployerAddress)
	ValidateError(t, err)

	gravityProgramID, err := DeploySolanaProgram(t, "gravity", "../private-keys/gravity3.json", deployerPrivateKeyPath, "../binaries/gravity.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()

	gravityStateAccount, err := GenerateNewAccount(deployerPrivateKey, GravityContractAllocation, gravityProgramID, endpoint)
	ValidateError(t, err)

	gravityMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, gravityProgramID, endpoint)
	ValidateError(t, err)

	bft := uint8(3)
	consulsPKlist := make([]types.Account, bft)

	var consulsKeysList []common.PublicKey

	for i := range consulsPKlist {
		consul := types.NewAccount()
		consulsPKlist[i] = consul

		consulsKeysList = append(consulsKeysList, consul.PublicKey)
	}

	// consulsKeysList := []common.PublicKey {
	// 	common.PublicKeyFromString("EnwGpvfZdCpkjs8jMShjo8evce2LbNfrYvREzdwGh5oc"),
	// 	common.PublicKeyFromString("ESgKDVemBdqDty6WExZ74kV8Re9yepth5tbKcsWTNXC9"),
	// 	common.PublicKeyFromString("5Ng92o7CPPWk5tT2pqrnRMndoD49d51f4QcocgJttGHS"),
	// }

	consuls := make([]byte, 0)
	for _, x := range consulsKeysList {
		consuls = append(consuls, x.Bytes()...)
	}

	// _, err = InitGravity(
	// 	deployerPrivateKey, gravityProgramID,
	// 	gravityStateAccount.Account.PublicKey.ToBase58(),
	// 	gravityMultisigAccount.Account.PublicKey.ToBase58(),
	// 	endpoint,
	// 	consuls,
	// )
	// ValidateError(t, err)
	time.Sleep(time.Second * 20)

	gravityExecutor, err := InitNebula(
		deployerPrivateKey,
		gravityProgramID,
		gravityStateAccount.Account.PublicKey.ToBase58(),
		gravityMultisigAccount.Account.PublicKey.ToBase58(),
		endpoint,
		common.PublicKeyFromString(gravityProgramID),
	)
	ValidateError(t, err)

	// t.Logf("before - Gravity Consuls Update should fail - program account is not initialized: %v \n", errFailing)

	_, errFailing = gravityExecutor.BuildAndInvoke(executor.UpdateConsulsGravityContractInstruction{
		Instruction: 1,
		Bft:         bft,
		LastRound:   10,
		Consuls:     append(consuls[:], consuls[:]...),
	})
	ValidateErrorExistence(t, errFailing)

	t.Logf("Gravity Consuls Update should fail - program account is not initialized: %v \n", errFailing)

	time.Sleep(time.Second * 20)

	gravityInitResponse, err := gravityExecutor.BuildAndInvoke(executor.InitGravityContractInstruction{
		Instruction: 0,
		Bft:         bft,
		InitRound:   1,
		Consuls:     consuls[:],
	})
	ValidateError(t, err)

	t.Logf("Gravity Init: %v \n", gravityInitResponse.TxSignature)

	time.Sleep(time.Second * 20)

	var signers []executor.GravityBftSigner
	// var additionalMeta []types.AccountMeta

	for _, signer := range consulsPKlist {
		signers = append(signers, *executor.NewGravityBftSigner(base58.Encode(signer.PrivateKey)))
		// additionalMeta = append(additionalMeta, types.AccountMeta{
		// 	PubKey: common.PublicKeyFromString(solana.ClockProgram), IsSigner: false, IsWritable: false
		// })
	}

	gravityExecutor.SetAdditionalSigners(signers)
	// gravityExecutor.SetAdditionalMeta(additionalMeta)
	// nebulaExecutor.SetAdditionalMeta([]types.AccountMeta {
	// 	{ PubKey: common.PublicKeyFromString(solana.ClockProgram), IsSigner: false, IsWritable: false },
	// })

	gravityConsulsUpdateResponse, err := gravityExecutor.BuildAndInvoke(executor.UpdateConsulsGravityContractInstruction{
		Instruction: 1,
		Bft:         bft,
		LastRound:   10,
		Consuls:     consuls,
	})
	ValidateError(t, err)

	t.Logf("Gravity Consuls Update: %v \n", gravityConsulsUpdateResponse.TxSignature)

	time.Sleep(time.Second * 20)
	_, errFailing = gravityExecutor.BuildAndInvoke(executor.UpdateConsulsGravityContractInstruction{
		Instruction: 1,
		Bft:         bft,
		LastRound:   0,
		Consuls:     consuls,
	})
	ValidateErrorExistence(t, errFailing)

	t.Logf("Gravity Consuls Update should fail - invalid last round: %v \n", errFailing)

	aftermathBalance, err := ReadAccountBalance(deployerAddress)
	ValidateError(t, err)

	t.Log("Deploy result in a success")
	t.Logf("Gravity Program ID: %v \n", gravityProgramID)
	t.Logf("Spent: %v SOL \n", initialBalance-aftermathBalance)
}
func TestPDA(t *testing.T) {
	tokenPDA, err := common.CreateProgramAddress([][]byte{[]byte("ibporttheminter"), []byte("ibporttheminter2")}, common.PublicKeyFromString("AgR3ZKBx7Ce7vLDBqX33uZAHELvB8z2Uu3exKDVNmVhU"))
	if err != nil {
		fmt.Printf("PDA error: %v\n", err)
		t.FailNow()
	}
	fmt.Printf("PDA: %s\n", tokenPDA.ToBase58())
	t.FailNow()
}
func TestIBPortContract(t *testing.T) {
	var err error
	deployerPrivateKeysPath := "../private-keys/_test_only-port-deployer.json"
	tokenOwnerPath := "../private-keys/_test_only-token-owner.json"
	ibportProgramPath := "../private-keys/_test_only_ibport-owner.json"

	err = CreatePersistedAccount(deployerPrivateKeysPath, true)
	ValidateError(t, err)
	err = CreatePersistedAccount(tokenOwnerPath, true)
	ValidateError(t, err)
	err = CreatePersistedAccount(ibportProgramPath, true)
	ValidateError(t, err)

	// tokenOwnerPrivateKey, err := ReadPKFromPath(t, tokenOwnerPath)
	// ValidateError(t, err)

	deployerAddress, err := ReadAccountAddress(deployerPrivateKeysPath)
	ValidateError(t, err)

	tokenOwnerAddress, err := ReadAccountAddress(tokenOwnerPath)
	ValidateError(t, err)

	ibportAddress, err := ReadAccountAddress(ibportProgramPath)
	ValidateError(t, err)

	waitTransactionConfirmations := func() {
		time.Sleep(time.Second * 30)
	}

	// SystemFaucet(t, deployerAddress, 10)
	// ValidateError(t, err)

	SystemFaucet(t, tokenOwnerAddress, 10)
	ValidateError(t, err)

	// SystemFaucet(t, ibportAddress, 10)
	// ValidateError(t, err)

	tokenDeployResult, err := CreateToken(tokenOwnerPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()
	// associatedTokenAccount, err := CreateTokenAccount(tokenOwnerPath, tokenProgramAddress)
	// ValidateError(t, err)

	associatedDeployerTokenAccount, err := CreateTokenAccount(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	fmt.Println("Generateing PDA")
	var tokenPDA common.PublicKey
	tokenPDA, err = common.CreateProgramAddress([][]byte{[]byte("ibport")}, common.PublicKeyFromString(ibportAddress))
	if err != nil {
		fmt.Printf("PDA error: %v", err)
		t.FailNow()
	}

	fmt.Printf("tokenPDA address: %s\n", tokenPDA.ToBase58())
	fmt.Printf("token program address: %s\n", tokenProgramAddress)

	t.Logf("tokenProgramAddress: %v", tokenProgramAddress)
	t.Logf("deployerAddress: %v", deployerAddress)
	t.Logf("tokenOwnerAddress: %v", tokenOwnerAddress)
	t.Logf("ibportAddress: %v", ibportAddress)
	t.Logf("associated token acc: %v", associatedDeployerTokenAccount)
	// deployerAddress, err := ReadAccountAddress(tokenOwnerPath)
	// ValidateError(t, err)

	// initialBalance, err := ReadAccountBalance(deployerAddress)
	// ValidateError(t, err)

	deployerPrivateKey, err := ReadPKFromPath(t, deployerPrivateKeysPath)
	ValidateError(t, err)

	// ibportPrivateKey, err := ReadPKFromPath(t, ibportProgramPath)
	// ValidateError(t, err)

	SystemFaucet(t, deployerAddress, 10)
	ValidateError(t, err)

	// love this *ucking timeouts
	time.Sleep(time.Second * 15)

	portProgramID, err := DeploySolanaProgram(t, "ibport", ibportProgramPath, deployerPrivateKeysPath, "../binaries/ibport.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()

	portDataAccount, err := GenerateNewAccount(deployerPrivateKey, IBPortAllocation, portProgramID, endpoint)
	ValidateError(t, err)

	ibportExecutor, err := InitNebula(
		deployerPrivateKey,
		portProgramID,
		portDataAccount.Account.PublicKey.ToBase58(),
		"",
		endpoint,
		common.PublicKeyFromString(portProgramID),
	)
	ValidateError(t, err)

	instructionBuilder := executor.NewIBPortInstructionBuilder()

	waitTransactionConfirmations()
	ibportInitResult, err := ibportExecutor.BuildAndInvoke(
		instructionBuilder.Init(common.PublicKeyFromBytes(make([]byte, 32)), common.PublicKeyFromString(tokenProgramAddress)),
	)
	ValidateError(t, err)
	t.Logf("IBPort Init: %v \n", ibportInitResult.TxSignature)

	waitTransactionConfirmations()

	// authorize ib port to mint token to provided account

	// AuthorizeToken
	err = AuthorizeToken(t, tokenOwnerPath, tokenProgramAddress, "mint", ibportAddress)
	ValidateError(t, err)

	time.Sleep(10 * time.Second)
	mintAmount := float64(55.5)

	deployerBeforeMintBalance, err := ReadSPLTokenBalance(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	ibportExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.PublicKeyFromString(associatedDeployerTokenAccount), IsSigner: false, IsWritable: true},
		{PubKey: tokenPDA, IsSigner: false, IsWritable: false},
	})

	ibportTestMintResult, err := ibportExecutor.BuildAndInvoke(
		instructionBuilder.TestMint(common.PublicKeyFromString(associatedDeployerTokenAccount), mintAmount),
	)
	ValidateError(t, err)

	waitTransactionConfirmations()

	deployerAfterMintBalance, err := ReadSPLTokenBalance(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	if deployerAfterMintBalance-deployerBeforeMintBalance != mintAmount {
		t.Log("error: balance mismatch")
		t.Logf("deployerBeforeMintBalance: %v", deployerBeforeMintBalance)
		t.Logf("deployerAfterMintBalance: %v", deployerAfterMintBalance)
	}

	t.Logf("IBPort Test Mint: %v \n", ibportTestMintResult.TxSignature)

	time.Sleep(time.Second * 20)
}
