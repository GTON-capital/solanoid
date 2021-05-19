package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"solanoid/models/endpoint"

	// "os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/common"
	// "strings"
)


func validateError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error: %v \n", err)
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

	nebulaProgramID, err := DeploySolanaProgram(t, "nebula", "../private-keys/nebula.json", "../binaries/nebula.so")
	validateError(t, err)

	deployerPrivateKeyPath := "../private-keys/gravity-deployer.json"
	deployerPrivateKey, err := readPKFromPath(t, deployerPrivateKeyPath)
	validateError(t, err)

	nebulaStateAccount, err := GenerateNewAccount(deployerPrivateKey, 2000, nebulaProgramID)
	validateError(t, err)
	t.Logf("nebula state account: %v \n", nebulaStateAccount.Account.PublicKey.ToBase58())
	
	nebulaMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, nebulaProgramID)
	validateError(t, err)
	t.Logf("nebula multisig state account: %v \n", nebulaMultisigAccount.Account.PublicKey.ToBase58())

	nebulaDeploymentResponse, err := InitNebula(
		deployerPrivateKey, 
		nebulaProgramID,
		nebulaStateAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		endpoint.LocalEnvironment,
		common.PublicKeyFromString(gravityProgramID),
	)
	validateError(t, err)

	t.Logf("Ser Message: %v \n", nebulaDeploymentResponse.SerializedMessage)

}
