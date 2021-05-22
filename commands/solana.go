package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/mr-tron/base58"
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

func InferSystemDefinedRPC() (string, error) {
	cmd := exec.Command("solana", "config", "get")
	output, err := cmd.CombinedOutput()
	
	rgx, _ := regexp.Compile("RPC URL: .+")
	result := rgx.Find(output)
	resultStr := strings.Trim(string(result), "\n\r ")
	resultList := strings.Split(resultStr, " ")
	rpcURL := resultList[len(resultList) - 1]
	
	fmt.Println(resultList)
	rpcURL = strings.Trim(rpcURL, "\n\r")

	if err != nil {
		return "", err
	}

	// t.Log(output)
	return rpcURL, nil
}

func DeploySolanaProgram(t *testing.T, tag string, programPrivateKeysPath, deployerPrivateKeysPath, programBinaryPath string) (string, error) {
	t.Log("deploying program")

	cmd := exec.Command("solana", "program", "deploy", "--keypair", deployerPrivateKeysPath, "--program-id", programPrivateKeysPath, programBinaryPath)

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

func ReadPKFromPath(t *testing.T, path string) (string, error) {
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