

# Solanoid

### Intro

Solanoid is framework for testing and building programs on Solana blockchain.

### Purpose

Solanoid aims to fulfill a gap between building and testing contracts on Solana. There's no yet built-in testing framework so we considered to present Solanoid.

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
7. Provides helper functions for interaction with `solana-cli`. Example [solana.go](commands/solana.go)

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

### Tutorial

To get the most of the Solanoid, follow these steps:

1. Fork the repository
2. If you build your own custom program: declare method signatures in `commands/executor`.
3. Tests can be declared in `commands/` directory.
4. Custom data models - in `models/`.

### Things to consider

1. When writing tests consider awaiting till confirmations reach MAX (via `	waitTransactionConfirmations()` call) - for Mainnet it's about 30 seconds, Devnet - 15 seconds. If you won't wait, state transition is not guaranteed. 
2. Tests require temporary addresses to operate with. For such purpose use `NewOperatingAddress` function.
3. Deployment tests require persisten addresses. For such purpose use `ReadOperatingAddress` function.