package commands

import (
	"time"

	// "os"

	"testing"

	"github.com/portto/solana-go-sdk/common"
	// "strings"
)

func TestDeployGravity(t *testing.T) {

	deployerPrivateKeyPath := "../private-keys/main-deployer.json"
	// deployerPrivateKeyPath := "/Users/shamil/.config/solana/id.json"
	deployerPrivateKey, err := ReadPKFromPath(t, deployerPrivateKeyPath)
	ValidateError(t, err)

	deployerAddress, err := ReadAccountAddress(deployerPrivateKeyPath)
	ValidateError(t, err)
	
	initialBalance, err := ReadAccountBalance(deployerAddress)
	ValidateError(t, err)

	gravityProgramID, err := DeploySolanaProgram(t, "gravity", "../private-keys/devnet-gravity.json", deployerPrivateKeyPath, "../binaries/gravity.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()
	
	gravityStateAccount, err := GenerateNewAccount(deployerPrivateKey, GravityContractAllocation, gravityProgramID, endpoint)
	ValidateError(t, err)

	gravityMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, gravityProgramID, endpoint)
	ValidateError(t, err)

	time.Sleep(time.Second * 30)

	consuls := make([]byte, 0)
	consulsKeysList := []common.PublicKey {
		common.PublicKeyFromString("EnwGpvfZdCpkjs8jMShjo8evce2LbNfrYvREzdwGh5oc"),
		common.PublicKeyFromString("ESgKDVemBdqDty6WExZ74kV8Re9yepth5tbKcsWTNXC9"),
		common.PublicKeyFromString("5Ng92o7CPPWk5tT2pqrnRMndoD49d51f4QcocgJttGHS"),
	}
	for _, x := range consulsKeysList {
		consuls = append(consuls, x.Bytes()...)
	}
	
	_, err = InitGravity(
		deployerPrivateKey, gravityProgramID, 
		gravityStateAccount.Account.PublicKey.ToBase58(),
		gravityMultisigAccount.Account.PublicKey.ToBase58(),
		endpoint,
		consuls,
	)
	ValidateError(t, err)

	aftermathBalance, err := ReadAccountBalance(deployerAddress)
	ValidateError(t, err)

	t.Log("Deploy result in a success")
	t.Logf("Gravity Program ID: %v \n", gravityProgramID)
	t.Logf("Spent: %v SOL \n", initialBalance - aftermathBalance)
}

// func TestGravityDeployment(t *testing.T) {
// 	// var err error

// 	// gravityProgramID := "BXDqLUQwWGDMQ6tFuca6mDLSZ1PgsS8T3R6oneXUUnoy"

// 	// nebulaProgramID := "CybfUMjVa13jLASS6BD53VvkeWChKHCWWZrs96dv5orN"
// 	// nebulaProgramID, err := DeploySolanaProgram(t, "nebula", "../private-keys/nebula.json", "../binaries/nebula.so")
// 	// ValidateError(t, err)

// 	// deployerPrivateKeyPath := "../private-keys/gravity-deployer.json"
// 	// deployerPrivateKey, err := readPKFromPath(t, deployerPrivateKeyPath)
// 	// ValidateError(t, err)

// 	// nebulaStateAccount, err := GenerateNewAccount(deployerPrivateKey, 2000, nebulaProgramID)
// 	// ValidateError(t, err)
// 	// t.Logf("nebula state account: %v \n", nebulaStateAccount.Account.PublicKey.ToBase58())
	
// 	// nebulaMultisigAccount, err := GenerateNewAccount(deployerPrivateKey, MultisigAllocation, nebulaProgramID)
// 	// ValidateError(t, err)
// 	// t.Logf("nebula multisig state account: %v \n", nebulaMultisigAccount.Account.PublicKey.ToBase58())

// 	// confirmationTimeout := time.Second * 20
// 	// t.Log("timeout 20 seconds - wait for MAX confirmations.")

// 	// time.Sleep(confirmationTimeout)

// 	nebulaExecutor, err := InitNebula(
// 		deployerPrivateKey, 
// 		nebulaProgramID,
// 		nebulaStateAccount.Account.PublicKey.ToBase58(),
// 		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
// 		endpoint.LocalEnvironment,
// 		common.PublicKeyFromString(gravityProgramID),
// 	)
// }