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


type GravityInstructionBuilder struct{}

func (port *GravityInstructionBuilder) Init(bft uint8, initRound uint64, consuls []byte) interface{} {
	return InitGravityContractInstruction {
		Instruction: 0,
		Bft:         bft,
		InitRound:   initRound,
		Consuls:     consuls[:],
	}
}