package commands

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/types"
)

func Test_hello(t *testing.T) {
	c := client.NewClient(client.TestnetRPCEndpoint)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		go resent(c)
	}
	time.Sleep(time.Second * 5)
	t.FailNow()
}

func resent(c *client.Client) {
	res, _ := c.GetRecentBlockhash()
	log.Print(res.Blockhash)
}

func Test_info(t *testing.T) {
	c := client.NewClient(client.TestnetRPCEndpoint)
	r, err := c.GetAccountInfo("9M8v1uHdMT96K8QN4rGz8BQ96hjASKhRN3cJKRZe6vtW", client.GetAccountInfoConfig{
		Encoding: "base64",
		DataSlice: client.GetAccountInfoConfigDataSlice{
			Length: 8,
			Offset: 66,
		},
	})
	if err != nil {
		t.FailNow()
	}
	sval, ok := r.Data.([]interface{})[0].(string)
	if !ok {
		t.FailNow()
	}

	val, err := base64.StdEncoding.DecodeString(sval)
	if err != nil {
		t.FailNow()
	}
	round := binary.LittleEndian.Uint64(val)
	log.Printf("round: %d", round)
	log.Printf("round hex: %s", hex.EncodeToString(val))

	t.FailNow()
}

func Test_BFT(t *testing.T) {
	c := client.NewClient(client.TestnetRPCEndpoint)
	r, err := c.GetAccountInfo("9M8v1uHdMT96K8QN4rGz8BQ96hjASKhRN3cJKRZe6vtW", client.GetAccountInfoConfig{
		Encoding: "base64",
		DataSlice: client.GetAccountInfoConfigDataSlice{
			Length: 1,
			Offset: 33,
		},
	})
	if err != nil {
		t.FailNow()
	}
	sval, ok := r.Data.([]interface{})[0].(string)
	if !ok {
		t.FailNow()
	}

	val, err := base64.StdEncoding.DecodeString(sval)
	if err != nil {
		t.FailNow()
	}

	log.Printf("round: %d", val[0])
	log.Printf("round hex: %s", hex.EncodeToString(val))

	t.FailNow()
}
func TestPK(t *testing.T) {
	pkb := []byte{191, 139, 28, 63, 155, 126, 202, 169, 17, 76, 109, 156, 253, 201, 126, 148, 169, 144, 159, 134, 35, 193, 105, 145, 81, 60, 249, 174, 143, 24, 200, 193, 169, 141, 86, 111, 41, 125, 46, 121, 106, 232, 52, 19, 48, 30, 84, 182, 30, 35, 85, 87, 49, 115, 176, 10, 151, 57, 116, 29, 176, 195, 75, 157}
	account := types.AccountFromPrivateKeyBytes(pkb)
	fmt.Println(account.PublicKey.ToBase58())
	fmt.Println(base58.Encode(account.PrivateKey))
	t.FailNow()
}
