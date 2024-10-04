package bank

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/somatic-labs/hardhat/lib"
	types "github.com/somatic-labs/hardhat/types"
)

func CreateBankSendMsg(config types.Config, fromAddress string, msgParams map[string]interface{}) (sdk.Msg, string, error) {
	fromAccAddress, err := sdk.AccAddressFromBech32(fromAddress)
	if err != nil {
		return nil, "", fmt.Errorf("invalid from address: %w", err)
	}

	toAccAddressInterface, ok := msgParams["to_address"]
	if !ok || toAccAddressInterface == nil {
		return nil, "", fmt.Errorf("missing 'to_address' in msgParams")
	}
	toAddressStr := toAccAddressInterface.(string)
	toAccAddress, err := sdk.AccAddressFromBech32(toAddressStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid to address: %w", err)
	}

	amount := sdk.NewCoins(sdk.NewCoin(config.Denom, sdkmath.NewInt(msgParams["amount"].(int64))))

	msg := banktypes.NewMsgSend(fromAccAddress, toAccAddress, amount)

	memo, err := lib.GenerateRandomStringOfLength(256)
	if err != nil {
		return nil, "", fmt.Errorf("error generating random memo: %w", err)
	}

	return msg, memo, nil
}
