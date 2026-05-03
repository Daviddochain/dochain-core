package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/cosmos/cosmos-sdk/client"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
)

const (
	DefaultGenesisBlockMaxBytes int64 = 22020096
	DefaultGenesisBlockMaxGas   int64 = 40000000
)

// InitCmd wraps the SDK init command, then rewrites config.toml, app.toml,
// and genesis.json so newly created nodes are born with the desired defaults.
func InitCmd(genesisBasicMgr module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := genutilcli.InitCmd(genesisBasicMgr, defaultNodeHome)

	originalRunE := cmd.RunE

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if originalRunE != nil {
			if err := originalRunE(cmd, args); err != nil {
				return err
			}
		}

		clientCtx := client.GetClientContextFromCmd(cmd)
		homeDir := clientCtx.HomeDir
		if homeDir == "" {
			homeDir = defaultNodeHome
		}

		configDir := filepath.Join(homeDir, "config")
		configTomlPath := filepath.Join(configDir, "config.toml")
		appTomlPath := filepath.Join(configDir, "app.toml")
		genesisPath := filepath.Join(configDir, "genesis.json")

		// ---------------------------------------------------------------------
		// Rewrite config.toml with custom Comet defaults
		// ---------------------------------------------------------------------
		tmCfg := initTendermintConfig()
		tmCfg.SetRoot(homeDir)
		tmconfig.WriteConfigFile(configTomlPath, tmCfg)

		// ---------------------------------------------------------------------
		// Rewrite app.toml with custom app defaults
		// ---------------------------------------------------------------------
		_, appCfgAny := initAppConfig()
		doAppCfg, ok := appCfgAny.(DoAppConfig)
		if !ok {
			return fmt.Errorf("unexpected app config type %T", appCfgAny)
		}
		srvconfig.WriteConfigFile(appTomlPath, doAppCfg)

		// ---------------------------------------------------------------------
		// Rewrite genesis.json directly as JSON
		// ---------------------------------------------------------------------
		genBz, err := os.ReadFile(genesisPath)
		if err != nil {
			return fmt.Errorf("failed to read genesis file %s: %w", genesisPath, err)
		}

		var gen map[string]interface{}
		if err := json.Unmarshal(genBz, &gen); err != nil {
			return fmt.Errorf("failed to unmarshal genesis file %s: %w", genesisPath, err)
		}

		// Top-level consensus_params.block
		consensusParams := ensureMap(gen, "consensus_params")
		topBlock := ensureMap(consensusParams, "block")
		topBlock["max_bytes"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxBytes)
		topBlock["max_gas"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxGas)

		legacyConsensus := ensureMap(gen, "consensus")
		legacyParams := ensureMap(legacyConsensus, "params")
		legacyBlock := ensureMap(legacyParams, "block")
		legacyBlock["max_bytes"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxBytes)
		legacyBlock["max_gas"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxGas)

		// app_state.consensus.block
		appState := ensureMap(gen, "app_state")
		consensusState := ensureMap(appState, "consensus")
		appBlock := ensureMap(consensusState, "block")
		appBlock["max_bytes"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxBytes)
		appBlock["max_gas"] = fmt.Sprintf("%d", DefaultGenesisBlockMaxGas)

		updatedBz, err := json.MarshalIndent(gen, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated genesis json: %w", err)
		}

		if err := os.WriteFile(genesisPath, updatedBz, 0o644); err != nil {
			return fmt.Errorf("failed to write updated genesis file %s: %w", genesisPath, err)
		}

		cmd.PrintErrf(
			"rewrote %s, %s and %s with fast defaults and consensus max_bytes=%d max_gas=%d\n",
			configTomlPath,
			appTomlPath,
			genesisPath,
			DefaultGenesisBlockMaxBytes,
			DefaultGenesisBlockMaxGas,
		)

		return nil
	}

	return cmd
}

func ensureMap(parent map[string]interface{}, key string) map[string]interface{} {
	if existing, ok := parent[key]; ok {
		if m, ok := existing.(map[string]interface{}); ok {
			return m
		}
	}

	m := map[string]interface{}{}
	parent[key] = m
	return m
}