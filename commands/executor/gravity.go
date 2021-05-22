package executor

type InitGravityContractInstruction struct {
	Bft       uint8
	Consuls []byte
	InitRound uint64 
}