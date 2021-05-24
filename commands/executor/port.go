package executor

import "github.com/portto/solana-go-sdk/common"


type InitIBPortInstruction struct {
	Instruction              uint8
	NebulaDataAccount        common.PublicKey
	TokenDataAccount         common.PublicKey
}

type TestCrossMintIBPortInstruction struct {
	Instruction              uint8
	Receiver                 common.PublicKey
	TokenAmount              float64
}

type TestCrossBurnIBPortInstruction struct {
	Instruction              uint8
	Burner                   common.PublicKey
	TokenAmount              float64
}