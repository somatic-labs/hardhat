package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

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
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
)

var client = &http.Client{
	Timeout: 10 * time.Second, // Adjusted timeout to 10 seconds
	Transport: &http.Transport{
		MaxIdleConns:        100,              // Increased maximum idle connections
		MaxIdleConnsPerHost: 10,               // Increased maximum idle connections per host
		IdleConnTimeout:     90 * time.Second, // Increased idle connection timeout
		TLSHandshakeTimeout: 10 * time.Second, // Increased TLS handshake timeout
	},
}

var cdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

func init() {
	types.RegisterInterfaces(cdc.InterfaceRegistry())
}

func sendIBCTransferViaRPC(config Config, rpcEndpoint string, chainID string, sequence, accnum uint64, privKey cryptotypes.PrivKey, pubKey cryptotypes.PubKey, address string) (response *coretypes.ResultBroadcastTx, txbody string, err error) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	encodingConfig.Marshaler = cdc

	// Register IBC and other necessary types
	transferModule := transfer.AppModuleBasic{}
	ibcModule := ibc.AppModuleBasic{}

	ibcModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	transferModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// Create a new TxBuilder.
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	//	receiver, _ := generateRandomString()
	token := sdk.NewCoin(config.Denom, sdk.NewInt(1))

	// JSON structure for the memo
	memo := `{
		"forward": {
		  "receiver": "pfm", // purposely using invalid bech32 here*
		  "port": "transfer",
		  "channel": "channel-229",
		  "timeout": "12h",
		  "retries": 10,
		  }
		}
	  }`

	ibcaddr, err := generateRandomString(config)
	if err != nil {
		return nil, "", err
	}

	msg := types.NewMsgTransfer(
		"transfer",
		config.Channel,
		token,
		address,
		ibcaddr,
		clienttypes.NewHeight(uint64(config.RevisionNumber), uint64(config.TimeoutHeight)), // Adjusted timeout height
		uint64(0),
		memo,
	)

	// set messages
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, "", err
	}

	// Estimate gas limit based on transaction size
	txSize := msg.Size()
	gasLimit := uint64((txSize * config.Bytes) + config.BaseGas)
	txBuilder.SetGasLimit(gasLimit)

	// Calculate fee based on gas limit and a fixed gas price

	gasPrice := sdk.NewDecCoinFromDec(config.Denom, sdk.NewDecWithPrec(config.Gas.Low, config.Gas.Precision)) // 0.00051 token per gas unit
	//	gasPrice := getGasPrice(config.Gas.Low, config.Denom)
	feeAmount := gasPrice.Amount.MulInt64(int64(gasLimit)).RoundInt()
	feecoin := sdk.NewCoin(config.Denom, feeAmount)
	txBuilder.SetFeeAmount(sdk.NewCoins(feecoin))

	txBuilder.SetMemo(config.Memo)
	txBuilder.SetTimeoutHeight(0)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: sequence,
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

	signed, err := tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
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
		return nil, "", fmt.Errorf("failed to broadcast transaction: %w", err)
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
		fmt.Println(err)
		fmt.Println("error at broadcast")
		return nil, err
	}

	fmt.Println("other: ", res.Data)
	fmt.Println("log: ", res.Log)
	fmt.Println("code: ", res.Code)
	fmt.Println("code: ", res.Codespace)
	fmt.Println("txid: ", res.Hash)

	return res, nil
}

func generateRandomString(config Config) (string, error) {
	// Generate a random size between config.RandMin and config.RandMax
	sizeB, err := rand.Int(rand.Reader, big.NewInt(int64(config.RandMax-config.RandMin+1)))
	if err != nil {
		return "", err
	}
	sizeB = sizeB.Add(sizeB, big.NewInt(int64(config.RandMin)))

	// Calculate the number of bytes to generate (2 characters per byte in hex encoding)
	nBytes := int(sizeB.Int64()) / 2

	randomBytes := make([]byte, nBytes)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
