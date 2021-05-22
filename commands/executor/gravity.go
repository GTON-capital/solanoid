package executor

type InitGravityContractInstruction struct {
	Instruction   uint8 
	Bft           uint8
	InitRound     uint64 
	Consuls     []byte
}

type UpdateConsulsGravityContractInstruction struct {
	Instruction   uint8 
	Bft           uint8
	LastRound     uint64 
	Consuls     []byte
}