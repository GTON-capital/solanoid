package instructions

import (
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

func InitNebulaInstruction(fromAccount, gravityProgramData, programID, nebulaAccount common.PublicKey, Bft uint8, DataType uint8, Oracles []common.PublicKey) (*types.Instruction, error) {
	/*
	   InitContract {
	       nebula_data_type: DataType,
	       gravity_contract_program_id: Pubkey,
	       initial_oracles: Vec<Pubkey>,
	       oracles_bft: u8,
	   },
	*/
	data, err := common.SerializeData(struct {
		Instruction        uint8
		DataType           uint8
		GravityProgramData common.PublicKey
		InitialOracles     []common.PublicKey
		Bft                uint8
	}{
		Instruction:        0,
		DataType:           DataType,
		GravityProgramData: gravityProgramData,
		InitialOracles:     Oracles,
		Bft:                Bft,
	})
	if err != nil {
		return nil, err
	}

	return &types.Instruction{
		Accounts: []types.AccountMeta{
			{PubKey: fromAccount, IsSigner: true, IsWritable: false},
			{PubKey: nebulaAccount, IsSigner: false, IsWritable: true},
			{PubKey: gravityProgramData, IsSigner: false, IsWritable: true},
		},
		ProgramID: programID,
		Data:      data,
	}, nil
}

func UpdateOraclesInstruction(fromAccount, programID, nebulaAccount common.PublicKey, CurrentOracles, NewOracles []common.PublicKey, pulseID uint64) (*types.Instruction, error) {
	/*UpdateOracles {
	  	new_oracles: Vec<Pubkey>,
	  	new_round: PulseID,
	  },
	*/
	data, err := common.SerializeData(struct {
		Instruction uint8
		NewOracles  []common.PublicKey
		PulseID     uint8
	}{
		Instruction: 1,
		NewOracles:  NewOracles,
		PulseID:     uint8(pulseID),
	})
	if err != nil {
		return nil, err
	}

	accounts := []types.AccountMeta{
		{PubKey: fromAccount, IsSigner: true, IsWritable: false},
		{PubKey: nebulaAccount, IsSigner: false, IsWritable: true},
	}
	for _, oracle := range CurrentOracles {
		accounts = append(accounts, types.AccountMeta{
			PubKey:     oracle,
			IsSigner:   false,
			IsWritable: false,
		})
	}
	return &types.Instruction{
		Accounts:  accounts,
		ProgramID: programID,
		Data:      data,
	}, nil

}

func SendHashValueInstruction(fromAccount, programID, nebulaAccount common.PublicKey, currentOracles []common.PublicKey, dataHash [16]byte) (*types.Instruction, error) {
	/*
	   SendHashValue {
	   	data_hash: UUID,
	   },
	*/
	data, err := common.SerializeData(struct {
		Instruction uint8
		DataHash    [16]byte
	}{
		Instruction: 2,
		DataHash:    dataHash,
	})
	if err != nil {
		return nil, err
	}

	accounts := []types.AccountMeta{
		{PubKey: fromAccount, IsSigner: true, IsWritable: false},
		{PubKey: nebulaAccount, IsSigner: false, IsWritable: true},
	}
	for _, oracle := range currentOracles {
		accounts = append(accounts, types.AccountMeta{
			PubKey:     oracle,
			IsSigner:   false,
			IsWritable: false,
		})
	}
	return &types.Instruction{
		Accounts:  accounts,
		ProgramID: programID,
		Data:      data,
	}, nil
}

func SendValueToSubsInstruction(fromAccount, programID, nebulaAccount common.PublicKey, currentOracles []common.PublicKey, pulseID uint64, subscriptionID [16]byte) (*types.Instruction, error) {

	/*
	   SendValueToSubs {
	   	data_type: DataType,
	   	pulse_id: PulseID,
	   	subscription_id: UUID,
	   },
	*/
	data, err := common.SerializeData(struct {
		Instruction   uint8
		PulseID       uint64
		SubsriptionID [16]byte
	}{
		Instruction:   3,
		PulseID:       pulseID,
		SubsriptionID: subscriptionID,
	})
	if err != nil {
		return nil, err
	}

	accounts := []types.AccountMeta{
		{PubKey: fromAccount, IsSigner: true, IsWritable: false},
		{PubKey: nebulaAccount, IsSigner: false, IsWritable: true},
	}
	for _, oracle := range currentOracles {
		accounts = append(accounts, types.AccountMeta{
			PubKey:     oracle,
			IsSigner:   false,
			IsWritable: false,
		})
	}
	return &types.Instruction{
		Accounts:  accounts,
		ProgramID: programID,
		Data:      data,
	}, nil

}

func SubscribeInstructions(fromAccount, programID, nebulaAccount, subscriberAddress common.PublicKey, currentOracles []common.PublicKey, minConfirmations uint8, reward uint64) (*types.Instruction, error) {

	/*
	   Subscribe {
	          address: Pubkey,
	          min_confirmations: u8,
	          reward: u64,
	      },
	*/
	data, err := common.SerializeData(struct {
		Instruction      uint8
		Address          common.PublicKey
		MinConfirmations uint8
		Reward           uint64
	}{
		Instruction:      4,
		Address:          subscriberAddress,
		MinConfirmations: minConfirmations,
		Reward:           reward,
	})
	if err != nil {
		return nil, err
	}

	accounts := []types.AccountMeta{
		{PubKey: fromAccount, IsSigner: true, IsWritable: false},
		{PubKey: nebulaAccount, IsSigner: false, IsWritable: true},
	}
	for _, oracle := range currentOracles {
		accounts = append(accounts, types.AccountMeta{
			PubKey:     oracle,
			IsSigner:   false,
			IsWritable: false,
		})
	}
	return &types.Instruction{
		Accounts:  accounts,
		ProgramID: programID,
		Data:      data,
	}, nil
}
