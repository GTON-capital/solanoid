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