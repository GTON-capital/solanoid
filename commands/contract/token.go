package contract

import (
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

const (
	RaydiumToken = "4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"
	SerumToken   = "SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"
)

const (
	AssociatedTokenProgram = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
)

// pub(crate) fn get_associated_token_address_and_bump_seed(
//     wallet_address: &Pubkey,
//     spl_token_mint_address: &Pubkey,
//     program_id: &Pubkey,
// ) -> (Pubkey, u8) {
//     get_associated_token_address_and_bump_seed_internal(
//         wallet_address,
//         spl_token_mint_address,
//         program_id,
//         &spl_token::id(),
//     )
// }
func GetAssociatedTokenAddressAndBumpSeed(
	walletAddress, splTokenMint, programId common.PublicKey,
) (common.PublicKey, int, error) {
	return GetAssociatedTokenAddressAndBumpSeedInternal(walletAddress, splTokenMint, programId, common.TokenProgramID)
}

// fn get_associated_token_address_and_bump_seed_internal(
//     wallet_address: &Pubkey,
//     spl_token_mint_address: &Pubkey,
//     program_id: &Pubkey,
//     token_program_id: &Pubkey,
// ) -> (Pubkey, u8) {
//     Pubkey::find_program_address(
//         &[
//             &wallet_address.to_bytes(),
//             &token_program_id.to_bytes(),
//             &spl_token_mint_address.to_bytes(),
//         ],
//         program_id,
//     )
// }
func GetAssociatedTokenAddressAndBumpSeedInternal(
	walletAddress, tokenMint, programId, tokenProgramId common.PublicKey,
) (common.PublicKey, int, error) {
	programAddress, bt, err := common.FindProgramAddress([][]byte{
		walletAddress.Bytes(),
		tokenProgramId.Bytes(),
		tokenMint.Bytes(),
	}, programId)
	return programAddress, bt, err
	// return common.FindAssociatedTokenAddress(walletAddress, tokenMint)
}

// /// Derives the associated token account address for the given wallet address and token mint
// pub fn get_associated_token_address(
//     wallet_address: &Pubkey,
//     spl_token_mint_address: &Pubkey,
// ) -> Pubkey {
//     get_associated_token_address_and_bump_seed(wallet_address, spl_token_mint_address, &id()).0
// }
func GetAssociatedTokenAddress(
	walletAddress, splTokenMint common.PublicKey,
) (common.PublicKey, error) {
	address, _, err := GetAssociatedTokenAddressAndBumpSeed(walletAddress, splTokenMint, common.PublicKeyFromString(AssociatedTokenProgram))
	return address, err
}

/// Create an associated token account for the given wallet address and token mint
///
/// Accounts expected by this instruction:
///
///   0. `[writeable,signer]` Funding account (must be a system account)
///   1. `[writeable]` Associated token account address to be created
///   2. `[]` Wallet address for the new associated token account
///   3. `[]` The token mint for the new associated token account
///   4. `[]` System program
///   5. `[]` SPL Token program
///   6. `[]` Rent sysvar
///
// pub fn create_associated_token_account(
//     funding_address: &Pubkey,
//     wallet_address: &Pubkey,
//     spl_token_mint_address: &Pubkey,
// ) -> Instruction {
//     let associated_account_address =
//         get_associated_token_address(wallet_address, spl_token_mint_address);

//     Instruction {
//         program_id: id(),
//         accounts: vec![
//             AccountMeta::new(*funding_address, true),
//             AccountMeta::new(associated_account_address, false),
//             AccountMeta::new_readonly(*wallet_address, false),
//             AccountMeta::new_readonly(*spl_token_mint_address, false),
//             AccountMeta::new_readonly(solana_program::system_program::id(), false),
//             AccountMeta::new_readonly(spl_token::id(), false),
//             AccountMeta::new_readonly(sysvar::rent::id(), false),
//         ],
//         data: vec![],
//     }
// }
func CreateAssociatedTokenAccountIX(fundingAddress, targetWallet, splTokenMint common.PublicKey) (*types.Instruction, common.PublicKey, error) {
	associatedAccountAddress, err := GetAssociatedTokenAddress(
		targetWallet, splTokenMint,
	)
	if err != nil {
		return nil, associatedAccountAddress, err
	}

	return &types.Instruction{
		ProgramID: common.PublicKeyFromString(AssociatedTokenProgram),
		Accounts: []types.AccountMeta{
			{PubKey: fundingAddress, IsSigner: true, IsWritable: true},
			{PubKey: associatedAccountAddress, IsSigner: false, IsWritable: true},
			{PubKey: targetWallet, IsSigner: false, IsWritable: false},
			{PubKey: splTokenMint, IsSigner: false, IsWritable: false},
			{PubKey: common.SystemProgramID, IsSigner: false, IsWritable: false},
			{PubKey: common.TokenProgramID, IsSigner: false, IsWritable: false},
			{PubKey: common.SysVarRentPubkey, IsSigner: false, IsWritable: false},
		},
	}, associatedAccountAddress, nil
}

func CreateAssociatedTokenAccountIXNonFailing(fundingAddress, splTokenMint common.PublicKey) (types.Account, *types.Instruction) {
	targetWallet := types.NewAccount()

	ix, _, err := CreateAssociatedTokenAccountIX(fundingAddress, targetWallet.PublicKey, splTokenMint)
	if err != nil {
		return CreateAssociatedTokenAccountIXNonFailing(fundingAddress, splTokenMint)
	}

	return targetWallet, ix
}
