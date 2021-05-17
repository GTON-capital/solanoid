package models

import (
	"github.com/portto/solana-go-sdk/types"
)

type CommandResponse struct {
	SerializedMessage string
	TxSignature       string
	Account           types.Account
}
