package commands

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mr-tron/base58/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

func TestMinterContract(t *testing.T) {
	var err error
	deployerPrivateKeysPath := "../private-keys/_test_only-port-deployer.json"
	tokenOwnerPath := "../private-keys/_test_only-token-owner.json"
	ibportProgramPath := "../private-keys/_test_only_ibport-owner.json"

	err = CreatePersistedAccount(deployerPrivateKeysPath, true)
	ValidateError(t, err)
	err = CreatePersistedAccount(tokenOwnerPath, true)
	ValidateError(t, err)
	err = CreatePersistedAccount(ibportProgramPath, true)
	ValidateError(t, err)

	deployerAddress, err := ReadAccountAddress(deployerPrivateKeysPath)
	ValidateError(t, err)

	tokenOwnerAddress, err := ReadAccountAddress(tokenOwnerPath)
	ValidateError(t, err)

	ibportAddress, err := ReadAccountAddress(ibportProgramPath)
	ValidateError(t, err)

	waitTransactionConfirmations := func() {
		time.Sleep(time.Second * 30)
	}

	SystemFaucet(t, tokenOwnerAddress, 10)
	ValidateError(t, err)

	tokenDeployResult, err := CreateToken(tokenOwnerPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()

	associatedDeployerTokenAccount, err := CreateTokenAccount(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	fmt.Println("Generateing PDA")
	var ibPortPDA common.PublicKey
	ibPortPDA, err = common.CreateProgramAddress([][]byte{[]byte("superminter")}, common.PublicKeyFromString(ibportAddress))
	if err != nil {
		fmt.Printf("PDA error: %v", err)
		t.FailNow()
	}

	fmt.Printf("minter PDA address: %s\n", ibPortPDA.ToBase58())

	t.Logf("tokenProgramAddress: %v", tokenProgramAddress)
	t.Logf("deployerAddress: %v", deployerAddress)
	t.Logf("tokenOwnerAddress: %v", tokenOwnerAddress)
	t.Logf("ibportAddress: %v", ibportAddress)
	t.Logf("associated token acc: %v", associatedDeployerTokenAccount)

	deployerPrivateKey, err := ReadPKFromPath(t, deployerPrivateKeysPath)
	ValidateError(t, err)

	deployerPKBytes, _ := base58.Decode(deployerPrivateKey)
	deployerAcc := types.AccountFromPrivateKeyBytes(deployerPKBytes)

	SystemFaucet(t, deployerAddress, 10)
	ValidateError(t, err)

	// love this *ucking timeouts
	time.Sleep(time.Second * 15)

	_, err = DeploySolanaProgram(t, "minter", ibportProgramPath, deployerPrivateKeysPath, "../binaries/minter.so")
	ValidateError(t, err)

	endpoint, _ := InferSystemDefinedRPC()

	// authorize ib port to mint token to provided account

	// AuthorizeToken
	err = AuthorizeToken(t, tokenOwnerPath, tokenProgramAddress, "mint", ibPortPDA.ToBase58())
	ValidateError(t, err)

	time.Sleep(35 * time.Second)
	mintAmount := float64(55.5)

	deployerBeforeMintBalance, err := ReadSPLTokenBalance(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	//----------------------------------------------------------------------
	log.Printf("Generating instruction for program %s", ibportAddress)
	ix := types.Instruction{
		ProgramID: common.PublicKeyFromString(ibportAddress),
		Accounts: []types.AccountMeta{
			//{PubKey: deployerAcc.PublicKey, IsWritable: false, IsSigner: true},
			{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
			{PubKey: common.PublicKeyFromString(tokenProgramAddress), IsWritable: true, IsSigner: false},
			//{PubKey: common.PublicKeyFromString(ibportAddress), IsWritable: false, IsSigner: false},
			{PubKey: common.PublicKeyFromString(associatedDeployerTokenAccount), IsWritable: true, IsSigner: false},
			{PubKey: ibPortPDA, IsWritable: false, IsSigner: false},
		},
		Data: []byte{0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	c := client.NewClient(endpoint)
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}

	message := types.NewMessage(
		deployerAcc.PublicKey,
		[]types.Instruction{
			ix,
		},
		res.Blockhash,
	)

	serializedMessage, err := message.Serialize()
	if err != nil {
		log.Fatalf("serialize message error, err: %v\n", err)
	}

	tx, err := types.CreateTransaction(message, map[common.PublicKey]types.Signature{
		deployerAcc.PublicKey: ed25519.Sign(deployerAcc.PrivateKey, serializedMessage),
	})
	if err != nil {
		log.Fatalf("generate tx error, err: %v\n", err)
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		log.Fatalf("serialize tx error, err: %v\n", err)
	}

	txSig, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("send tx error, err: %v\n", err)
	}

	log.Printf("upload program txHash : %s", txSig)

	//----------------------------------------------------------------------

	waitTransactionConfirmations()

	deployerAfterMintBalance, err := ReadSPLTokenBalance(deployerPrivateKeysPath, tokenProgramAddress)
	ValidateError(t, err)

	if deployerAfterMintBalance-deployerBeforeMintBalance != mintAmount {
		t.Log("error: balance mismatch")
		t.Logf("deployerBeforeMintBalance: %v", deployerBeforeMintBalance)
		t.Logf("deployerAfterMintBalance: %v", deployerAfterMintBalance)
	}

	t.Logf("IBPort Test Mint: %v \n", txSig)

	time.Sleep(time.Second * 20)
}
