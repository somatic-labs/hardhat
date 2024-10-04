package wasm

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/somatic-labs/hardhat/types"
)

func CreateStoreCodeMsg(config types.Config, sender string, wasmFile []byte) (sdk.Msg, error) {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	msg := wasmtypes.MsgStoreCode{
		Sender:       senderAddr.String(),
		WASMByteCode: wasmFile,
	}
	return &msg, nil
}

func CreateInstantiateContractMsg(config types.Config, sender string, codeID uint64, initMsg []byte, label string, funds sdk.Coins) (sdk.Msg, error) {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	msg := wasmtypes.MsgInstantiateContract{
		Sender: senderAddr.String(),
		Admin:  senderAddr.String(), // Using sender as admin, adjust if needed
		CodeID: codeID,
		Label:  label,
		Msg:    initMsg,
		Funds:  funds,
	}
	return &msg, nil
}

func CreateExecuteContractMsg(config types.Config, sender string, contractAddr string, execMsg []byte, funds sdk.Coins) (sdk.Msg, error) {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	contractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	msg := wasmtypes.MsgExecuteContract{
		Sender:   senderAddr.String(),
		Contract: contractAddress.String(),
		Msg:      execMsg,
		Funds:    funds,
	}
	return &msg, nil
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
