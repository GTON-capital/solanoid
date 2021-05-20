package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"solanoid/commands/executor"
	"solanoid/models/endpoint"
	"solanoid/models/nebula"
	"solanoid/models/solana"
	"time"

	// "os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	// "strings"
)


func ValidateError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error: %v \n", err)
		t.FailNow()
	}
}

func ValidateErrorExistence(t *testing.T, err error) {
	if err == nil {
		t.Errorf("No error occured!")
		t.FailNow()
	}
}


func SystemFaucet(t *testing.T, privateKeyPath string, recipient string, amount uint64) (string, error) {
	t.Logf("transfer %v SOL to %v address \n", amount, recipient)

	cmd := exec.Command("solana", "transfer", recipient, fmt.Sprint(amount))

	output, err := cmd.CombinedOutput()
	t.Log(string(output))

	if err != nil {
		t.Log(err.Error())
		log.Fatal(err)
	}

	// t.Log(output)
	
	return programID, nil
}

func DeploySolanaProgram(t *testing.T, tag string, programPrivateKeysPath, programBinaryPath string) (string, error) {
	t.Log("deploying program")

	cmd := exec.Command("solana", "program", "deploy", "--program-id", programPrivateKeysPath, programBinaryPath)

	output, err := cmd.CombinedOutput()
	
	t.Log(string(output))

	outputList := strings.Split(string(output), " ")
	programID := outputList[len(outputList) - 1]
	programID = strings.Trim(programID, "\n\r")

	t.Logf("Program: %v; Deployed Program ID is: %v\n", tag, programID)
	// t.Logf("Program: %v; Deployed Program ID is: %v\n", tag, common.PublicKeyFromString(programID))

	if err != nil {
		t.Log(err.Error())
		log.Fatal(err)
	}

	// t.Log(output)
	
	return programID, nil
}

func readPKFromPath(t *testing.T, path string) (string, error) {
	result, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	var input []byte

	err = json.Unmarshal(result, &input)
	if err != nil {
		return "", err
	}

	encodedPrivKey := base58.Encode(input)
	t.Logf("priv key: %v \n", encodedPrivKey)

	return encodedPrivKey, nil
}

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

	gravityProgramID := "BXDqLUQwWGDMQ6tFuca6mDLSZ1PgsS8T3R6oneXUUnoy"

	// nebulaProgramID := "CybfUMjVa13jLASS6BD53VvkeWChKHCWWZrs96dv5orN"
	nebulaProgramID, err := DeploySolanaProgram(t, "nebula", "../private-keys/nebula.json", "../binaries/nebula.so")
	ValidateError(t, err)

	deployerPrivateKeyPath := "../private-keys/gravity-deployer.json"
	deployerPrivateKey, err := readPKFromPath(t, deployerPrivateKeyPath)
	ValidateError(t, err)

	nebulaStateAccount, err := GenerateNewAccount(deployerPrivateKey, 2000, nebulaProgramID)
	ValidateError(t, err)
	t.Logf("nebula state account: %v \n", nebulaStateAccount.Account.PublicKey.ToBase58())
	
	nebulaMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, nebulaProgramID)
	ValidateError(t, err)
	t.Logf("nebula multisig state account: %v \n", nebulaMultisigAccount.Account.PublicKey.ToBase58())

	confirmationTimeout := time.Second * 20
	t.Log("timeout 20 seconds - wait for MAX confirmations.")

	time.Sleep(confirmationTimeout)

	nebulaExecutor, err := InitNebula(
		deployerPrivateKey, 
		nebulaProgramID,
		nebulaStateAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		endpoint.LocalEnvironment,
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
		// Oracles: nebulaExecutor.Deployer().Bytes(),
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

	mockedSubscriber, err := GenerateNewAccount(deployerPrivateKey, 1024, nebulaProgramID)
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
