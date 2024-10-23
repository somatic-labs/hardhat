package broadcast

import (
	"context"
	"fmt"
	"log"

	cometrpc "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/ibc-go/modules/apps/callbacks/testing/simapp/params"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	meteoritebank "github.com/somatic-labs/meteorite/modules/bank"
	meteoriteibc "github.com/somatic-labs/meteorite/modules/ibc"
	wasm "github.com/somatic-labs/meteorite/modules/wasm"
	types "github.com/somatic-labs/meteorite/types"

	sdkmath "cosmossdk.io/math"

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
	"github.com/cosmos/cosmos-sdk/x/gov"

	wasmd "github.com/CosmWasm/wasmd/x/wasm"
)

var cdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

func init() {
	transfertypes.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
}

func SendTransactionViaRPC(config types.Config, rpcEndpoint, chainID string, sequence, accnum uint64,
	privKey cryptotypes.PrivKey, pubKey cryptotypes.PubKey, fromAddress, msgType string,
	msgParams types.MsgParams,
) (response *coretypes.ResultBroadcastTx, txbody string, err error) {
	encodingConfig := params.MakeTestEncodingConfig()
	encodingConfig.Codec = cdc

	// Register IBC and other necessary types
	transferModule := transfer.AppModuleBasic{}
	ibcModule := ibc.AppModuleBasic{}
	bankModule := bank.AppModuleBasic{}
	wasmModule := wasmd.AppModuleBasic{}
	govModule := gov.AppModuleBasic{}

	ibcModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	transferModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	bankModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	wasmModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	govModule.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// Create a new TxBuilder.
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	var msg sdk.Msg
	var memo string // Declare a variable to hold the memo

	switch msgType {
	case "ibc_transfer":
		msg, memo, err = meteoriteibc.CreateIBCTransferMsg(config, fromAddress, msgParams)
		if err != nil {
			return nil, "", err
		}
	case "bank_send":
		msg, memo, err = meteoritebank.CreateBankSendMsg(config, fromAddress, msgParams)
		if err != nil {
			return nil, "", err
		}
	case "store_code":
		msg, memo, err = wasm.CreateStoreCodeMsg(config, fromAddress, msgParams)
		if err != nil {
			return nil, "", err
		}
	case "instantiate_contract":
		msg, memo, err = wasm.CreateInstantiateContractMsg(config, fromAddress, msgParams)
		if err != nil {
			return nil, "", err
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

	resp, err := Transaction(txJSONBytes, rpcEndpoint)
	if err != nil {
		return resp, string(txJSONBytes), fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return resp, string(txJSONBytes), nil
}

func Transaction(txBytes []byte, rpcEndpoint string) (*coretypes.ResultBroadcastTx, error) {
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
