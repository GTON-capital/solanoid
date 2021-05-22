package commands

import "testing"



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