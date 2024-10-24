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

// SendTransactionViaRPC sends a transaction using the provided TransactionParams and sequence number.
func SendTransactionViaRPC(txParams types.TransactionParams, sequence uint64) (response *coretypes.ResultBroadcastTx, txbody string, err error) {
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

	switch txParams.MsgType {
	case "ibc_transfer":
		msg, memo, err = meteoriteibc.CreateIBCTransferMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "bank_send":
		msg, memo, err = meteoritebank.CreateBankSendMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "store_code":
		msg, memo, err = wasm.CreateStoreCodeMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	case "instantiate_contract":
		msg, memo, err = wasm.CreateInstantiateContractMsg(txParams.Config, txParams.AcctAddress, txParams.MsgParams)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("unsupported message type: %s", txParams.MsgType)
	}

	// Set messages
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, "", err
	}

	// Estimate gas limit based on transaction size
	txSize := len(msg.String())
	gasLimit := uint64((int64(txSize) * txParams.Config.Bytes) + txParams.Config.BaseGas)
	txBuilder.SetGasLimit(gasLimit)

	// Calculate fee based on gas limit and a fixed gas price
	gasPrice := sdk.NewDecCoinFromDec(txParams.Config.Denom, sdkmath.LegacyNewDecWithPrec(txParams.Config.Gas.Low, txParams.Config.Gas.Precision))
	feeAmount := gasPrice.Amount.MulInt64(int64(gasLimit)).RoundInt()
	feecoin := sdk.NewCoin(txParams.Config.Denom, feeAmount)
	txBuilder.SetFeeAmount(sdk.NewCoins(feecoin))

	// Set the memo (either random for bank_send or as per IBC transfer)
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)

	// First round: gather all the signer infos using the "set empty signature" hack
	sigV2 := signing.SignatureV2{
		PubKey:   txParams.PubKey,
		Sequence: sequence,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode(encodingConfig.TxConfig.SignModeHandler().DefaultMode()),
		},
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		fmt.Println("Error setting signatures")
		return nil, "", err
	}

	signerData := authsigning.SignerData{
		ChainID:       txParams.ChainID,
		AccountNumber: txParams.AccNum,
		Sequence:      sequence,
	}

	ctx := context.Background()

	signed, err := tx.SignWithPrivKey(
		ctx,
		signing.SignMode(encodingConfig.TxConfig.SignModeHandler().DefaultMode()),
		signerData,
		txBuilder,
		txParams.PrivKey,
		encodingConfig.TxConfig,
		sequence,
	)
	if err != nil {
		fmt.Println("Couldn't sign")
		return nil, "", err
	}

	err = txBuilder.SetSignatures(signed)
	if err != nil {
		return nil, "", err
	}

	// Generate the encoded transaction bytes
	txBytes, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}

	resp, err := Transaction(txBytes, txParams.NodeURL)
	if err != nil {
		return resp, string(txBytes), fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return resp, string(txBytes), nil
}

// Transaction broadcasts the transaction bytes to the given RPC endpoint.
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
