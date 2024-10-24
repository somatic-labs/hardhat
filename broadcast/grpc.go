package broadcast

import (
	"context"
	"fmt"

	"github.com/cosmos/ibc-go/modules/apps/callbacks/testing/simapp/params"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	client "github.com/somatic-labs/meteorite/client"
	meteoritebank "github.com/somatic-labs/meteorite/modules/bank"
	meteoriteibc "github.com/somatic-labs/meteorite/modules/ibc"
	wasm "github.com/somatic-labs/meteorite/modules/wasm"
	types "github.com/somatic-labs/meteorite/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"

	wasmd "github.com/CosmWasm/wasmd/x/wasm"
)

func SendTransactionViaGRPC(
	ctx context.Context,
	txParams types.TransactionParams,
	sequence uint64,
	grpcClient *client.GRPCClient,
) (*sdk.TxResponse, string, error) {
	encodingConfig := params.MakeTestEncodingConfig()
	encodingConfig.Codec = cdc

	// Register necessary interfaces
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
	var memo string

	// Construct the message based on the message type
	switch txParams.MsgType {
	case "ibc_transfer":
		var err error
		msg, memo, err = meteoriteibc.CreateIBCTransferMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "bank_send":
		var err error
		msg, memo, err = meteoritebank.CreateBankSendMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "store_code":
		var err error
		msg, memo, err = wasm.CreateStoreCodeMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "instantiate_contract":
		var err error
		msg, memo, err = wasm.CreateInstantiateContractMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("unsupported message type: %s", txParams.MsgType)
	}

	// Set the message and other transaction parameters
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, "", err
	}

	// Estimate gas limit
	txSize := len(msg.String())
	gasLimit := uint64((int64(txSize) * txParams.Config.Bytes) + txParams.Config.BaseGas)
	txBuilder.SetGasLimit(gasLimit)

	// Calculate fee
	gasPrice := sdk.NewDecCoinFromDec(txParams.Config.Denom, sdkmath.LegacyNewDecWithPrec(txParams.Config.Gas.Low, txParams.Config.Gas.Precision))
	feeAmount := gasPrice.Amount.MulInt64(int64(gasLimit)).RoundInt()
	feeCoin := sdk.NewCoin(txParams.Config.Denom, feeAmount)
	txBuilder.SetFeeAmount(sdk.NewCoins(feeCoin))

	// Set memo and timeout height
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)

	// Set up signature
	sigV2 := signing.SignatureV2{
		PubKey:   txParams.PubKey,
		Sequence: sequence,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_DIRECT,
		},
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, "", err
	}

	signerData := authsigning.SignerData{
		ChainID:       txParams.ChainID,
		AccountNumber: txParams.AccNum,
		Sequence:      sequence,
	}

	// Sign the transaction
	if _, err := tx.SignWithPrivKey(
		ctx,
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		txParams.PrivKey,
		encodingConfig.TxConfig,
		sequence,
	); err != nil {
		return nil, "", err
	}

	// Encode the transaction
	txBytes, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, "", err
	}

	// Broadcast the transaction via gRPC
	grpcRes, err := grpcClient.SendTx(ctx, txBytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to broadcast transaction via gRPC: %w", err)
	}

	// Check for errors in the response
	if grpcRes.Code != 0 {
		return grpcRes, string(txBytes), fmt.Errorf("broadcast error code %d: %s", grpcRes.Code, grpcRes.RawLog)
	}

	return grpcRes, string(txBytes), nil
}

func BuildAndSignTx(
	txParams types.TransactionParams,
	sequence uint64,
) ([]byte, string, error) {
	encodingConfig := params.MakeTestEncodingConfig()
	encodingConfig.Codec = cdc

	// Create a new TxBuilder.
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	var msg sdk.Msg
	var memo string
	var err error

	// Construct the message based on the message type
	switch txParams.MsgType {
	case "ibc_transfer":
		msg, memo, err = meteoriteibc.CreateIBCTransferMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
	case "bank_send":
		msg, memo, err = meteoritebank.CreateBankSendMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
	case "store_code":
		msg, memo, err = wasm.CreateStoreCodeMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
	case "instantiate_contract":
		msg, memo, err = wasm.CreateInstantiateContractMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
	default:
		return nil, "", fmt.Errorf("unsupported message type: %s", txParams.MsgType)
	}

	if err != nil {
		return nil, "", err
	}

	// Set the message and other transaction parameters
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, "", err
	}

	// Estimate gas limit
	txSize := len(msg.String())
	gasLimit := uint64((int64(txSize) * txParams.Config.Bytes) + txParams.Config.BaseGas)
	txBuilder.SetGasLimit(gasLimit)

	// Calculate fee
	gasPrice := sdk.NewDecCoinFromDec(txParams.Config.Denom, sdkmath.LegacyNewDecWithPrec(txParams.Config.Gas.Low, txParams.Config.Gas.Precision))
	feeAmount := gasPrice.Amount.MulInt64(int64(gasLimit)).RoundInt()
	feeCoin := sdk.NewCoin(txParams.Config.Denom, feeAmount)
	txBuilder.SetFeeAmount(sdk.NewCoins(feeCoin))

	// Set memo and timeout height
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)

	// Set up signature
	sigV2 := signing.SignatureV2{
		PubKey:   txParams.PubKey,
		Sequence: sequence,
		Data: &signing.SingleSignatureData{
			SignMode: encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
		},
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, "", err
	}

	signerData := authsigning.SignerData{
		ChainID:       txParams.ChainID,
		AccountNumber: txParams.AccNum,
		Sequence:      sequence,
	}

	// Sign the transaction
	if err := tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder,
		txParams.PrivKey,
		encodingConfig.TxConfig,
		sequence,
	); err != nil {
		return nil, "", err
	}

	// Encode the transaction
	txBytes, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, "", err
	}

	return txBytes, memo, nil
}
