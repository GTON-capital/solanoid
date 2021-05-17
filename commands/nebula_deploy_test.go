package commands

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"testing"
	// "strings"
)


func validateError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error: %v \n", err)
		t.FailNow()
	}
}



func DeploySolanaProgram(programPrivateKeysPath, programBinaryPath string, t *testing.T) (string, error) {
	// cmd := exec.Command("cat", "../binaries/nebula.so")
	// cmd := exec.Command("cat", "../private-keys/nebula.json")
	t.Log("deploying program")
	// cmd := exec.Command("solana", "program", "deploy", "--program-id", "../private-keys/nebula.json", "../binaries/nebula.so")
	cmd := exec.Command("solana", "program", "deploy", "--program-id", programPrivateKeysPath, programBinaryPath)
	var out bytes.Buffer
	
	output, err := cmd.CombinedOutput()
	cmd.Stdout = &out

	t.Log(string(output))

	outputList := strings.Split(string(output), " ")
	programID := outputList[len(outputList) - 1]

	t.Logf("Program ID is: %v\n", programID)

	if err != nil {
		t.Log(err.Error())
		log.Fatal(err)
	}

	t.Log(&out)
	
	return programID, nil
}



func TestNebulaDeployment(t *testing.T) {
	var err error

	gravityProgramID, err := DeploySolanaProgram("../private-keys/gravity.json", "../binaries/gravity.so", t)
	validateError(t, err)

	nebulaProgramID, err := DeploySolanaProgram("../private-keys/nebula.json", "../binaries/nebula.so", t)
	validateError(t, err)

	deployerPrivateKey := ""

	gravityStateAccount, err := GenerateNewAccount(deployerPrivateKey, space, gravityProgramID)
	validateError(t, err)

	nebulaStateAccount, err := GenerateNewAccount(deployerPrivateKey, space, nebulaProgramID)
	validateError(t, err)

	nebulaMultisigAccount, err :=GenerateNewAccount(deployerPrivateKey, space, nebulaProgramID)
	validateError(t, err)

	_, _, _ = gravityStateAccount, nebulaStateAccount, nebulaMultisigAccount
}

