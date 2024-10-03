package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	cometrpc "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/modules/apps/callbacks/testing/simapp/params"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)

var client = &http.Client{
	Timeout: 500 * time.Millisecond, // Adjusted timeout to 10 seconds
	Transport: &http.Transport{
		MaxIdleConns:        100,              // Increased maximum idle connections
		MaxIdleConnsPerHost: 10,               // Increased maximum idle connections per host
		IdleConnTimeout:     90 * time.Second, // Increased idle connection timeout
		TLSHandshakeTimeout: 10 * time.Second, // Increased TLS handshake timeout
	},
}

// Memo represents the structure of the memo field in the transaction
type Memo struct {
	Forward Forward `json:"forward"`
}

// Forward contains details about the forwarding information
type Forward struct {
	Receiver string   `json:"receiver"`
	Port     string   `json:"port"`
	Channel  string   `json:"channel"`
	Timeout  string   `json:"timeout"`
	Retries  int      `json:"retries"`
	Next     *Forward `json:"next,omitempty"`
}

// ToJSON converts the Memo struct to a JSON string
func (m *Memo) ToJSON() (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NewMemo creates a new Memo struct with default values
func NewMemo(config Config) *Memo {
	return &Memo{
		Forward: Forward{
			Receiver: strings.Repeat(config.IbCMemo, config.IbCMemoRepeat), // Note: This is an invalid bech32 address
			Port:     "transfer",
			Channel:  "channel-569",
			Timeout:  "12h",
			Retries:  10,
		},
	}
}

var cdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

func init() {
	types.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
}

func sendTransactionViaRPC(config Config, rpcEndpoint string, chainID string, sequence, accnum uint64,
	privKey cryptotypes.PrivKey, pubKey cryptotypes.PubKey, fromAddress string, msgType string,
	msgParams map[string]interface{}) (response *coretypes.ResultBroadcastTx, txbody string, err error) {

	encodingConfig := params.MakeTestEncodingConfig()
	encodingConfig.Codec = cdc

	// Register IBC and other necessary types
	transferModule := transfer.AppModuleBasic{}
	ibcModule := ibc.AppModuleBasic{}
	bankModule := bank.AppModuleBasic{}

	ibcModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	transferModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	bankModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// Create a new TxBuilder.
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	var msg sdk.Msg
	var memo string // Declare a variable to hold the memo

	switch msgType {
	case "ibc_transfer":
		token := sdk.NewCoin(config.Denom, sdkmath.NewInt(msgParams["amount"].(int64)))
		memoStruct := NewMemo(config)
		jsonMemo, err := memoStruct.ToJSON()
		if err != nil {
			return nil, "", fmt.Errorf("error converting memo to JSON: %w", err)
		}
		ibcaddr, err := generateRandomString(config)
		if err != nil {
			return nil, "", err
		}
		ibcaddr = sdk.MustBech32ifyAddressBytes(config.Prefix, []byte(ibcaddr))
		msg = types.NewMsgTransfer(
			"transfer",
			config.Channel,
			token,
			fromAddress,
			ibcaddr,
			clienttypes.NewHeight(uint64(config.RevisionNumber), uint64(config.TimeoutHeight)),
			uint64(0),
			jsonMemo,
		)
		memo = jsonMemo // Set the memo for IBC transfer

	case "bank_send":
		// Decode 'fromAddress' from Bech32 to sdk.AccAddress
		fromAccAddress, err := sdk.AccAddressFromBech32(fromAddress)
		if err != nil {
			return nil, "", fmt.Errorf("invalid from address: %w", err)
		}

		// Decode 'toAddress' from Bech32 to sdk.AccAddress
		toAccAddressInterface, ok := msgParams["to_address"]
		if !ok || toAccAddressInterface == nil {
			return nil, "", fmt.Errorf("missing 'to_address' in msgParams")
		}
		toAddressStr := toAccAddressInterface.(string)
		toAccAddress, err := sdk.AccAddressFromBech32(toAddressStr)
		if err != nil {
			return nil, "", fmt.Errorf("invalid to address: %w", err)
		}

		// Construct the amount
		amount := sdk.NewCoins(sdk.NewCoin(config.Denom, sdkmath.NewInt(msgParams["amount"].(int64))))

		// Create the MsgSend message
		msg = banktypes.NewMsgSend(fromAccAddress, toAccAddress, amount)

		// Generate a 256-byte random string for the memo
		memo, err = generateRandomStringOfLength(256)
		if err != nil {
			return nil, "", fmt.Errorf("error generating random memo: %w", err)
		}

	default:
		return nil, "", fmt.Errorf("unsupported message type: %s", msgType)
	}

	// Set messages
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, "", err
	}

	// Estimate gas limit based on transaction size
	txSize := len(msg.String())
	gasLimit := uint64((int64(txSize) * config.Bytes) + config.BaseGas)
	txBuilder.SetGasLimit(gasLimit)

	// Calculate fee based on gas limit and a fixed gas price
	gasPrice := sdk.NewDecCoinFromDec(config.Denom, sdkmath.LegacyNewDecWithPrec(config.Gas.Low, config.Gas.Precision))
	feeAmount := gasPrice.Amount.MulInt64(int64(gasLimit)).RoundInt()
	feecoin := sdk.NewCoin(config.Denom, feeAmount)
	txBuilder.SetFeeAmount(sdk.NewCoins(feecoin))

	// Set the memo (either random for bank_send or as per IBC transfer)
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)

	// First round: we gather all the signer infos. We use the "set empty signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey:   pubKey,
		Sequence: sequence,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode(encodingConfig.TxConfig.SignModeHandler().DefaultMode()),
		},
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		fmt.Println("error setting signatures")
		return nil, "", err
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
	}

	ctx := context.Background()

	signed, err := tx.SignWithPrivKey(ctx,
		signing.SignMode(encodingConfig.TxConfig.SignModeHandler().DefaultMode()), signerData,
		txBuilder, privKey, encodingConfig.TxConfig, sequence)
	if err != nil {
		fmt.Println("couldn't sign")
		return nil, "", err
	}

	err = txBuilder.SetSignatures(signed)
	if err != nil {
		return nil, "", err
	}

	// Generate a JSON string.
	txJSONBytes, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}

	resp, err := BroadcastTransaction(txJSONBytes, rpcEndpoint)
	if err != nil {
		return resp, string(txJSONBytes), fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return resp, string(txJSONBytes), nil
}

func BroadcastTransaction(txBytes []byte, rpcEndpoint string) (*coretypes.ResultBroadcastTx, error) {
	cmtCli, err := cometrpc.New(rpcEndpoint, "/websocket")
	if err != nil {
		log.Fatal(err)
	}

	t := tmtypes.Tx(txBytes)

	ctx := context.Background()
	res, err := cmtCli.BroadcastTxSync(ctx, t)
	if err != nil {
		fmt.Println("Error at broadcast:", err)
		return nil, err
	}

	if res.Code != 0 {
		// Return an error containing the code and log message
		return res, fmt.Errorf("broadcast error code %d: %s", res.Code, res.Log)
	}

	return res, nil
}

func generateRandomString(config Config) (string, error) {
	// Generate a random size between config.RandMin and config.RandMax
	sizeB, err := rand.Int(rand.Reader, big.NewInt(config.RandMax-config.RandMin+1))
	if err != nil {
		return "", err
	}
	sizeB = sizeB.Add(sizeB, big.NewInt(config.RandMin))

	// Calculate the number of bytes to generate (2 characters per byte in hex encoding)
	nBytes := int(sizeB.Int64()) / 2

	randomBytes := make([]byte, nBytes)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

func generateRandomStringOfLength(n int) (string, error) {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}
