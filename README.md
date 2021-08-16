

# Solanoid

### Intro

Solanoid is framework for testing and building programs on Solana blockchain written in Go.

### Purpose

Solanoid aims to fill the gap between writing and testing contracts on Solana. There's no yet built-in testing framework so we considered to present Solanoid.

### Dependencies

1. Go >= 1.15.
2. `solana-cli`.

### Features

1. Allows binding compiled smart contracts via symlinks for deployment via framework. Example: [Details](bind-symlink.sh)

```bash
...

bind-symlink "../solana-adapter/src/gravity-core-adapter" "nebula.so" "nebula/target/deploy/solana_nebula_contract.so"

## Example

## bind-symlink $project_root $binary_name $path_to_binary
...
```

2. Provides examples for end-to-end testing. Example: [full program test](commands/flow_test.go)
3. Provides abstractions for signing/sending transactions and calling programs. Example: [caller abstraction, based on Nebula](commands/executor/nebula.go#L361) 
4. Provides helper functions for Solana. Example [new data account](commands/new_data_account.go#L224)

```go
func GenerateNewTokenAccount(privateKey string, space uint64, owner, tokenMint common.PublicKey, clientEndpoint string, seeds string) (*models.CommandResponse, error) 
```
5. Provides instruction building approach. [Example](commands/executor/ib_port.go)

```go

type NebulaInstructionBuilder struct{}

var NebulaIXBuilder = &NebulaInstructionBuilder{}

func (port *NebulaInstructionBuilder) Init(bft, dataType uint8, gravityProgramID common.PublicKey, oracles []byte) interface{} {
	return InitNebulaContractInstruction{
		Instruction:              0,
		Bft:                      bft,
		NebulaDataType:           dataType,
		GravityContractProgramID: gravityProgramID,
		InitialOracles:           oracles,
	}
}

```
6. Provides management of temporary and storage persistent private keys. Example [Operational](commands/operational.go)
7. Provides helper functions for interaction with `solana-cli`, `spl-token`. Example [solana.go](commands/solana.go)

8. Offers parallel deployment of programs. Example: [Solana gateway deployment](commands/flow_test.go#L114)
```go

	ParallelExecution(
		[]func(){
			func() {
				_, err = DeploySolanaProgram(t, "ibport", ibportProgram.PKPath, consulsList.List[0].PKPath, "../binaries/ibport.so")
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
```

9. Deployment via tests. Example [commands/gateway_test.go](commands/gateway_test.go#L14)
10. Facility for writing MVPs between Solana and EVM [Solana and EVM Gateway MVP](commands/mvp/gateway_mvp_test.go) (Polygon is disabled atm)

### Tutorial on Deployment/Testing with/without Multisig.

To get the most of the Solanoid, follow these steps:

1. Fork the repository. Clone it.

```bash
git clone <your_profile_or_org>/solanoid
```

2. If you build your own custom program: declare method signatures in `commands/executor`.

3. Check this invocation without Multisig.

```go
// commands/executor/helloworld.go

package executor

type SayHelloWorldIX struct {
	Instruction uint8
	Message     string
}

type HelloWorldProgramIXBuilder struct{}

func (builder *HelloWorldProgramIXBuilder) SayHello(message string) interface{} {
	return SayHelloWorldIX{
		Instruction: 0,
		Message: message,
	}
}

// Then call your method via abstraction
// commands/helloworld_test.go

func TestHelloWorld(t *testing.T) {
	deployer, err := NewOperatingAddress(t, "path_to_deployer", nil)
	ValidateError(t, err)

	helloWorldProgram, err := NewOperatingAddress(t, "path_to_program_address", nil)
	ValidateError(t, err)


	// RPC is inferred via `solana config get`
	RPCEndpoint, _ := InferSystemDefinedRPC()

	// deployment
	_, err = DeploySolanaProgram(t, "helloworld", helloWorldProgram.PKPath, deployer.PKPath, "path_to_program_binary")
	ValidateError(t, err)

	// contract bytes allocation
	HelloWorldContractAllocation := 750

	helloWorldDataAccount, err := GenerateNewAccount(deployer.PrivateKey, HelloWorldContractAllocation, helloWorldProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	helloWorldExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		helloWorldProgram.PublicKey.ToBase58(),
		helloWorldDataAccount.Account.PublicKey.ToBase58(),
		"", // Empty if we want NO MULTISIG
		RPCEndpoint,
		common.PublicKeyFromString(""), // can be omitted always
	)
	ValidateError(t, err)

	// we wait right before every method call
	waitTransactionConfirmations()

	ixbuilder := &executor.HelloWorldProgramIXBuilder{}

	sayHelloResponse, err := helloWorldExecutor.BuildAndInvoke(
		ixbuilder.SayHello("HELLO WORLD"),
	)
	fmt.Printf("Hello World Call Result: %v \n", sayHelloResponse.TxSignature)
	ValidateError(t, err)
}
```
4. For methods requiring multisig, just do this.


```go
// inside test
... 
	// if we want additional accounts
	// those are concated to the end of the list

	// Like this:
	// signer + <n> of multisigs + <n> of additional meta
	currentExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
	})


	multisigMember, err := ReadOperatingAddress(t, "path_to_multisig_member")
	ValidateError(t, err)

	var signers []executor.GravityBftSigner
	signers = append(signers, *executor.NewGravityBftSigner(multisigMember.PrivateKey))

	// For multisigs
	currentExecutor.SetAdditionalSigners(signers)


	// multisigs and metas can be erased
	currentExecutor.EraseAdditionalMeta()
	currentExecutor.EraseAdditionalSigners()
...
```
5. Tests can be declared in `commands/` directory.
6. Custom data models - in `models/`.


# MVPs between Solana <-> EVM and vice versa

<!-- 1. You need to repeat same steps mentioned in Testing/Deployment tutorial. -->
Solanoid provides an example on how to write MVPs for dApps between EVM and Solana. Please consider check it here [this gateway example between Solana and EVM](commands/mvp/gateway_mvp_test.go).

### Things to consider

1. When writing tests consider awaiting till confirmations reach MAX (via `	waitTransactionConfirmations()` call) - for Mainnet it's about 30 seconds, Devnet - 15 seconds. If you won't wait, state transition is not guaranteed. 
2. Tests require temporary addresses to operate with. For such purpose use `NewOperatingAddress` function.
3. Deployment tests require persisten addresses. For such purpose use `ReadOperatingAddress` function.
