package types

import (
	"go/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type TransactionParams struct {
	Config      types.Config
	NodeURL     string
	ChainID     string
	Sequence    uint64
	AccNum      uint64
	PrivKey     cryptotypes.PrivKey
	PubKey      cryptotypes.PubKey
	AcctAddress string
	MsgType     string
	MsgParams   map[string]interface{}
}
