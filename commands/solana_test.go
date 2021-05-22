package commands

import "testing"



func TestEndpointInfer(t *testing.T) {
	rpc, err := InferSystemDefinedRPC()

	t.Log(err)
	t.Log(rpc)

	if rpc == "" {
		t.Error("rpc is empty")
		t.FailNow()
	}
}