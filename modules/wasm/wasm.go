package wasm

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type WasmHandler struct {
	txConfig tx.Config
}

func NewWasmHandler(txConfig tx.Config) *WasmHandler {
	return &WasmHandler{
		txConfig: txConfig,
	}
}

func (h *WasmHandler) StoreCode(sender sdk.AccAddress, wasmFile []byte) (*sdk.TxResponse, error) {
	msg := wasmtypes.MsgStoreCode{
		Sender:       sender.String(),
		WASMByteCode: wasmFile,
	}

	return h.broadcastTx(sender, &msg)
}

func (h *WasmHandler) InstantiateContract(sender sdk.AccAddress, codeID uint64, initMsg []byte, label string) (*sdk.TxResponse, error) {
	msg := wasmtypes.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: codeID,
		Label:  label,
		Msg:    initMsg,
		Funds:  sdk.NewCoins(),
	}

	return h.broadcastTx(sender, &msg)
}

func (h *WasmHandler) ExecuteContract(sender sdk.AccAddress, contractAddr string, execMsg []byte, funds sdk.Coins) (*sdk.TxResponse, error) {
	msg := wasmtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contractAddr,
		Msg:      execMsg,
		Funds:    funds,
	}

	return h.broadcastTx(sender, &msg)
}

func (h *WasmHandler) broadcastTx(sender sdk.AccAddress, msg sdk.Msg) (*sdk.TxResponse, error) {
	// Implementation of broadcasting the transaction
	// This is a placeholder and needs to be implemented based on your specific SDK setup
	return nil, fmt.Errorf("broadcastTx not implemented")
}

// Helper function to create a StoreFile message
func CreateStoreFileMsg(data []byte) ([]byte, error) {
	msg := struct {
		StoreFile struct {
			Data []byte `json:"data"`
		} `json:"store_file"`
	}{
		StoreFile: struct {
			Data []byte `json:"data"`
		}{
			Data: data,
		},
	}

	return json.Marshal(msg)
}
