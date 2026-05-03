package keeper

import (
    errorsmod "cosmossdk.io/errors"
    wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
    wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
    wasmvmtypes "github.com/CosmWasm/wasmvm/v3/types"
    treasurykeeper "github.com/Daviddochain/dochain-core/v4/x/treasury/keeper"
    "github.com/cosmos/cosmos-sdk/baseapp"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
    bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// msgEncoder is an extension point to customize encodings
type msgEncoder interface {
    // Encode converts wasmvm message to n cosmos message types
    Encode(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Msg, error)
}

// MessageRouter ADR 031 request type routing type
type MessageRouter interface {
    Handler(msg sdk.Msg) baseapp.MsgServiceHandler
}

// SDKMessageHandler can handle messages that can be encoded into sdk.Message types and routed.
type SDKMessageHandler struct {
    router   MessageRouter
    encoders msgEncoder
}

func NewMessageHandler(
    router MessageRouter,
    ics4Wrapper wasmtypes.ICS4Wrapper,
    channelKeeper wasmtypes.ChannelKeeper,
    bankKeeper bankKeeper.Keeper,
    taxexemptionKeeper interface{},
    treasuryKeeper treasurykeeper.Keeper,
    accountKeeper authkeeper.AccountKeeper,
    taxKeeper interface{},
    unpacker codectypes.AnyUnpacker,
    portSource wasmtypes.ICS20TransferPortSource,
    customEncoders ...*wasmkeeper.MessageEncoders,
) wasmkeeper.Messenger {
    _ = ics4Wrapper
    _ = channelKeeper
    _ = taxexemptionKeeper
    _ = treasuryKeeper
    _ = accountKeeper
    _ = taxKeeper

    encoders := wasmkeeper.DefaultEncoders(unpacker, portSource)
    for _, e := range customEncoders {
        encoders = encoders.Merge(e)
    }

    return wasmkeeper.NewMessageHandlerChain(
        NewSDKMessageHandler(router, encoders),
        wasmkeeper.NewBurnCoinMessageHandler(bankKeeper),
    )
}

func NewSDKMessageHandler(router MessageRouter, encoders msgEncoder) SDKMessageHandler {
    return SDKMessageHandler{
        router:   router,
        encoders: encoders,
    }
}

func (h SDKMessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgs [][]*codectypes.Any, err error) {
    sdkMsgs, err := h.encoders.Encode(ctx, contractAddr, contractIBCPortID, msg)
    if err != nil {
        return nil, nil, nil, err
    }

    for _, sdkMsg := range sdkMsgs {
        res, err := h.handleSdkMessage(ctx, contractAddr, sdkMsg)
        if err != nil {
            return nil, nil, nil, err
        }

        data = append(data, res.Data)

        sdkEvents := make([]sdk.Event, len(res.Events))
        for i := range res.Events {
            sdkEvents[i] = sdk.Event(res.Events[i])
        }
        events = append(events, sdkEvents...)
    }

    return events, data, nil, nil
}

func (h SDKMessageHandler) handleSdkMessage(ctx sdk.Context, contractAddr sdk.Address, msg sdk.Msg) (*sdk.Result, error) {
    if msgValidate, ok := msg.(sdk.HasValidateBasic); ok {
        if err := msgValidate.ValidateBasic(); err != nil {
            return nil, err
        }
    }

    if msgSigners, ok := msg.(sdk.LegacyMsg); ok {
        for _, acct := range msgSigners.GetSigners() {
            if !acct.Equals(contractAddr) {
                return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "contract doesn't have permission")
            }
        }
    }

    if handler := h.router.Handler(msg); handler != nil {
        msgResult, err := handler(ctx, msg)
        return msgResult, err
    }

    return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
}
