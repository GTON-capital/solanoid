package commands

import (
	"testing"
)



func TestEndpointInfer(t *testing.T) {
	rpc, err := InferSystemDefinedRPC()

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(rpc)
	
	if rpc == "" {
		t.Error("rpc is empty")
		t.FailNow()
	}
	
}

func TestTokenCreate(t *testing.T) {
	tokenRes, err := CreateToken("../private-keys/gravity3.json")

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(tokenRes.Owner.ToBase58())
	t.Log(tokenRes.Token.ToBase58())
	t.Log(tokenRes.Signature)
	
	// if balance == 0 {
	// 	t.Log("balance is empty")
	// }
}

func TestPersistedAccount(t *testing.T) {
	err := CreatePersistedAccount("../private-keys/test1234f44.json", true)

	t.Log(err)
}


func TestFullTokenBehaviour(t *testing.T) {
	// CreateTokenAccount
	testMockedPrivateKeyPath := "../private-keys/_test_owner-full-token-behaviour.json"
	burnerPrivateKeysPath := "../private-keys/_test_burner-full-token-behaviour.json"
	_ = CreatePersistedAccount(testMockedPrivateKeyPath, true)
	_ = CreatePersistedAccount(burnerPrivateKeysPath, true)

	tokenRes, err := CreateToken(testMockedPrivateKeyPath)

	if err != nil || tokenRes == nil {
		t.Log(err)
		t.FailNow()
	}

	tokenAddress := tokenRes.Token.ToBase58()

	t.Log(tokenRes.Owner.ToBase58())
	t.Log(tokenRes.Token.ToBase58())
	t.Log(tokenRes.Signature)
	
	associatedTokenDataAccount, err := CreateTokenAccount(testMockedPrivateKeyPath, tokenAddress)
	if err != nil || associatedTokenDataAccount == "" {
		t.Log(err)
		t.FailNow()
	}
	burnerTokenDataAccount, err := CreateTokenAccount(burnerPrivateKeysPath, tokenAddress)
	if err != nil || burnerTokenDataAccount == "" {
		t.Log(err)
		t.FailNow()
	}
	
	// mint some

	var currentBalance float64
	mintableAmount :=  10.2352

	updateCurrentBalance := func() {
		currentBalance, err = ReadSPLTokenBalance(burnerPrivateKeysPath, tokenAddress)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}

	err = MintToken(testMockedPrivateKeyPath, tokenAddress, mintableAmount, burnerTokenDataAccount)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	updateCurrentBalance()

	if currentBalance != mintableAmount {
		t.Log("balance mismatch")
		t.Logf("current: %v \n", currentBalance)
		t.Logf("mintable amount: %v \n", mintableAmount)
		t.FailNow()
	}

	err = BurnToken(burnerPrivateKeysPath, burnerTokenDataAccount, mintableAmount)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	updateCurrentBalance()

	if currentBalance != 0 {
		t.Log("balance mismatch")
		t.Logf("current: %v \n", currentBalance)
		t.Logf("desired balance: %v \n", 0)
		t.FailNow()
	}


}

func TestTokenDelegation(t *testing.T) {
	// CreateTokenAccount
	testMockedPrivateKeyPath := "../private-keys/_test_owner-full-token-behaviour.json"
	delegatePrivateKeysPath := "../private-keys/_test_burner-full-token-behaviour.json"
	distinctPrivateKeysPath := "../private-keys/_test_dist-full-token-behaviour.json"

	_ = CreatePersistedAccount(testMockedPrivateKeyPath, true)
	_ = CreatePersistedAccount(delegatePrivateKeysPath, true)
	_ = CreatePersistedAccount(distinctPrivateKeysPath, true)

	tokenRes, err := CreateToken(testMockedPrivateKeyPath)

	if err != nil || tokenRes == nil {
		t.Log(err)
		t.FailNow()
	}

	tokenAddress := tokenRes.Token.ToBase58()

	t.Log(tokenRes.Owner.ToBase58())
	t.Log(tokenRes.Token.ToBase58())
	t.Log(tokenRes.Signature)

	delegateAddress, err := ReadAccountAddress(delegatePrivateKeysPath)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	associatedTokenDataAccount, err := CreateTokenAccount(testMockedPrivateKeyPath, tokenAddress)
	if err != nil || associatedTokenDataAccount == "" {
		t.Log(err)
		t.FailNow()
	}
	tokenDelegateTokenDataAccount, err := CreateTokenAccount(delegatePrivateKeysPath, tokenAddress)
	if err != nil || tokenDelegateTokenDataAccount == "" {
		t.Log(err)
		t.FailNow()
	}
	distinctTokenDataAccount, err := CreateTokenAccount(distinctPrivateKeysPath, tokenAddress)
	if err != nil || distinctTokenDataAccount == "" {
		t.Log(err)
		t.FailNow()
	}
	
	// mint some

	var ownerBalance, delegateBalance float64
	mintableAmount :=  10.2352

	updateCurrentBalance := func() {
		ownerBalance, err = ReadSPLTokenBalance(testMockedPrivateKeyPath, tokenAddress)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		delegateBalance, err = ReadSPLTokenBalance(delegatePrivateKeysPath, tokenAddress)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}

	err = MintToken(testMockedPrivateKeyPath, tokenAddress, mintableAmount, associatedTokenDataAccount)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	updateCurrentBalance()

	if ownerBalance != mintableAmount {
		t.Log("balance mismatch")
		t.Logf("current: %v \n", ownerBalance)
		t.Logf("mintable amount: %v \n", mintableAmount)
		t.FailNow()
	}

	err = DelegateSPLTokenAmount(testMockedPrivateKeyPath, associatedTokenDataAccount, delegateAddress, mintableAmount)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("delegation occured successfully!")
	updateCurrentBalance()

	t.Logf("current: %v \n", ownerBalance)
	t.Logf("delegateBalance: %v \n", delegateBalance)
	t.Logf("mintable amount: %v \n", mintableAmount)

	// err = TransferSPLTokens(delegatePrivateKeysPath, tokenAddress, distinctTokenDataAccount, associatedTokenDataAccount, 1)
	// if err != nil {
	// 	t.Log(err)
	// 	t.FailNow()
	// }
	err = BurnToken(delegatePrivateKeysPath, associatedTokenDataAccount, mintableAmount)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	updateCurrentBalance()

	if ownerBalance != 0 || delegateBalance != 0 {
		t.Log("balance mismatch")
		t.Logf("ownerBalance: %v \n", ownerBalance)
		t.Logf("delegateBalance: %v \n", delegateBalance)
		t.Logf("desired balance: %v \n", 0)
		t.FailNow()
	}

	t.Log("Test passed successfully.")
}


func TestBalanceRead(t *testing.T) {
	addr := "Es5jjKAHHyDaVVzuFEqyV56eg1R1xCQKTDHVPp3enzTN"
	balance, err := ReadAccountBalance(addr)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(balance)

	if balance == 0 {
		t.Log("balance is empty")
	}
}