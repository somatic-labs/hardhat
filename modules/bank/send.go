package bank

import (
	"fmt"

	"github.com/somatic-labs/hardhat/lib"
	types "github.com/somatic-labs/hardhat/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func CreateBankSendMsg(config types.Config, fromAddress string, msgParams types.MsgParams) (sdk.Msg, string, error) {
	fromAccAddress, err := sdk.AccAddressFromBech32(fromAddress)
	if err != nil {
		return nil, "", fmt.Errorf("invalid from address: %w", err)
	}

	toAccAddress, err := sdk.AccAddressFromBech32(msgParams.ToAddress)
	if err != nil {
		fmt.Println("invalid to address, spamming random new accounts")
		toAccAddress, err = lib.GenerateRandomAccount()
		if err != nil {
			return nil, "", fmt.Errorf("error generating random account: %w", err)
		}
	}

	amount := sdk.NewCoins(sdk.NewCoin(config.Denom, sdkmath.NewInt(msgParams.Amount)))

	msg := banktypes.NewMsgSend(fromAccAddress, toAccAddress, amount)

	memo, err := lib.GenerateRandomStringOfLength(256)
	if err != nil {
		return nil, "", fmt.Errorf("error generating random memo: %w", err)
	}

	return msg, memo, nil
}
