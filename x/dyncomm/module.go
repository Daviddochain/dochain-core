package dyncomm

import (
    "context"
    "encoding/json"
    "fmt"
    "math/rand"

    "cosmossdk.io/math"
    "github.com/Daviddochain/dochain-core/v4/x/dyncomm/client/cli"
    "github.com/Daviddochain/dochain-core/v4/x/dyncomm/keeper"
    "github.com/Daviddochain/dochain-core/v4/x/dyncomm/types"
    "github.com/Daviddochain/dochain-core/v4/x/market/simulation"

    abci "github.com/cometbft/cometbft/abci/types"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/spf13/cobra"
)

var (
    _ module.AppModule      = AppModule{}
    _ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct {
    cdc codec.Codec
}

type AppModule struct {
    AppModuleBasic
    keeper        keeper.Keeper
    stakingKeeper types.StakingKeeper
}

func NewAppModule(
    cdc codec.Codec,
    keeper keeper.Keeper,
    stakingKeeper types.StakingKeeper,
) AppModule {
    return AppModule{
        AppModuleBasic: AppModuleBasic{cdc},
        keeper:         keeper,
        stakingKeeper:  stakingKeeper,
    }
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
    types.RegisterLegacyAminoCodec(cdc)
}

func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
    types.RegisterInterfaces(registry)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
    return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
    // temporarily disabled
}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (AppModuleBasic) GetQueryCmd() *cobra.Command { return cli.GetQueryCmd() }

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (AppModule) IsAppModule() {}

func (AppModule) IsOnePerModuleType() {}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    gs := ExportGenesis(ctx, am.keeper)
    return cdc.MustMarshalJSON(gs)
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
    var genesisState types.GenesisState
    cdc.MustUnmarshalJSON(data, &genesisState)
    InitGenesis(ctx, am.keeper, &genesisState)
    return nil
}

func (AppModule) QuerierRoute() string { return types.QuerierRoute }

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) RegisterServices(cfg module.Configurator) {
    // temporarily disabled gRPC query registration while dyncomm query proto/service is inconsistent
}

func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
    dyncommGenesis := types.DefaultGenesisState()
    params := types.DefaultParams()
    params.Cap = math.LegacyZeroDec()
    dyncommGenesis.Params = params
    bz, err := json.MarshalIndent(&dyncommGenesis, "", " ")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Selected default dyncomm parameters:\n%s\n", bz)
    simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(dyncommGenesis)
}

func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
    return []simtypes.WeightedProposalContent{}
}

func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.LegacyParamChange {
    return []simtypes.LegacyParamChange{}
}

func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
    sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

func (am AppModule) WeightedOperations(module.SimulationState) []simtypes.WeightedOperation {
    return nil
}

func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    EndBlocker(sdkCtx, am.keeper)
    return []abci.ValidatorUpdate{}, nil
}
