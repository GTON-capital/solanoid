package commands

import (
	"solanoid/commands/executor"
	"solanoid/models/nebula"
	"solanoid/models/solana"
	"time"

	// "os"

	"testing"

	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	// "strings"
)

// func deployGravity(t *testing.T) {

// 	gravityProgramID, err := DeploySolanaProgram(t, "gravity", "../private-keys/gravity.json", "../binaries/gravity.so")
// 	validateError(t, err)

// 	gravityStateAccount, err := GenerateNewAccount(deployerPrivateKey, GravityContractAllocation, nebulaProgramID)
// 	validateError(t, err)

// 	gravityMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, nebulaProgramID)
// 	validateError(t, err)

// 	gravityDeploymentResponse, err := InitGravity(
// 		deployerPrivateKey, gravityProgramID,
// 		gravityStateAccount.Account.PublicKey.ToBase58(), gravityMultisigAccount.Account.PublicKey.ToBase58()
// 	)
// 	validateError(t, err)

// 	_, err = SystemFaucet(t, deployerPrivateKeyPath, nebulaProgramID, 1)
// 	validateError(t, err)

// }

func TestNebulaDeployment(t *testing.T) {
	var err error

	endpoint, _ := InferSystemDefinedRPC()

	gravityProgramID := "BXDqLUQwWGDMQ6tFuca6mDLSZ1PgsS8T3R6oneXUUnoy"
	
	deployerPrivateKeyPath := "../private-keys/main-deployer.json"

	// deployerPrivateKeyPath := "../private-keys/gravity-deployer.json"
	deployerPrivateKey, err := ReadPKFromPath(t, deployerPrivateKeyPath)

	// nebulaProgramID := "CybfUMjVa13jLASS6BD53VvkeWChKHCWWZrs96dv5orN"
	nebulaProgramID, err := DeploySolanaProgram(t, "nebula", "../private-keys/nebula2.json", deployerPrivateKeyPath, "../binaries/nebula.so")
	ValidateError(t, err)


	ValidateError(t, err)

	nebulaStateAccount, err := GenerateNewAccount(deployerPrivateKey, 2000, nebulaProgramID, endpoint)
	ValidateError(t, err)
	t.Logf("nebula state account: %v \n", nebulaStateAccount.Account.PublicKey.ToBase58())
	
	nebulaMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, nebulaProgramID, endpoint)
	ValidateError(t, err)
	t.Logf("nebula multisig state account: %v \n", nebulaMultisigAccount.Account.PublicKey.ToBase58())

	confirmationTimeout := time.Second * 20
	t.Log("timeout 20 seconds - wait for MAX confirmations.")

	time.Sleep(confirmationTimeout)

	nebulaExecutor, err := InitGenericExecutor(
		deployerPrivateKey, 
		nebulaProgramID,
		nebulaStateAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		endpoint,
		common.PublicKeyFromString(gravityProgramID),
	)
	ValidateError(t, err)

	nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(executor.InitNebulaContractInstruction {
		Instruction: 0,
		Bft: 1,
		NebulaDataType: nebula.Bytes,
		GravityContractProgramID: common.PublicKeyFromString(gravityProgramID),
		InitialOracles: nebulaExecutor.Deployer().Bytes(),
	})
	ValidateError(t, err)

	t.Logf("Init: %v \n", nebulaInitResponse.SerializedMessage)

	time.Sleep(time.Second * 25)

	// Vital for update oracles (multisig)
	nebulaExecutor.SetAdditionalSigners([]executor.GravityBftSigner {
		*executor.NewGravityBftSigner(deployerPrivateKey),
	})
	
	nebulaUpdateOraclesResponse, err := nebulaExecutor.BuildAndInvoke(executor.UpdateOraclesNebulaContractInstruction {
		Instruction: 1,
		Bft: 1,
		Oracles: nebulaExecutor.Deployer().Bytes(),
		NewRound: 1,
	})
	ValidateError(t, err)

	t.Logf("Update Oracles: %v \n", nebulaUpdateOraclesResponse.SerializedMessage)

	time.Sleep(time.Second * 25)

	nebulaExecutor.SetAdditionalMeta([]types.AccountMeta {
		{ PubKey: common.PublicKeyFromString(solana.ClockProgram), IsSigner: false, IsWritable: false },
	})
	nebulaSendHashValueResponse, err := nebulaExecutor.BuildAndInvoke(executor.SendHashValueNebulaContractInstructionn {
		Instruction: 2,
		DataHash: make([]byte, 16),
	})
	ValidateError(t, err)

	t.Logf("Send Hash Value: %v \n", nebulaSendHashValueResponse.SerializedMessage)

	nebulaExecutor.EraseAdditionalMeta()

	time.Sleep(time.Second * 3)

	nebulaSendValueToSubsResponse, err := nebulaExecutor.BuildAndInvoke(executor.SendValueToSubsNebulaContractInstructionn {
		Instruction: 3,
		DataHash: make([]byte, 32 * 3),
		DataType: [1]byte { nebula.Bytes },
		PulseID: [8]byte{},
		SubscriptionID: [16]byte{},
	})
	ValidateError(t, err)

	t.Logf("Send Value To Subs: %v \n", nebulaSendValueToSubsResponse.SerializedMessage)
	
	time.Sleep(time.Second * 3)

	mockedSubscriber, err := GenerateNewAccount(deployerPrivateKey, 1024, nebulaProgramID, endpoint)
	ValidateError(t, err)
	t.Logf("mocked subscriber state account: %v \n", mockedSubscriber.Account.PublicKey.ToBase58())

	nebulaSubscribeResponse, err := nebulaExecutor.BuildAndInvoke(executor.SubscribeNebulaContractInstructionn {
		Instruction: 4,
		Subscriber: mockedSubscriber.Account.PublicKey,
		MinConfirmations: 1,
		Reward: 1,
	})
	ValidateError(t, err)

	t.Logf("Subscribe: %v \n", nebulaSubscribeResponse.SerializedMessage)

	// time.Sleep(time.Second * 3)

	nebulaExecutor.EraseAdditionalSigners()
}
