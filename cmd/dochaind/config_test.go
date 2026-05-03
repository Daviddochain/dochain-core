package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverrideConfigCacheSize(t *testing.T) {
	_, cfg := initAppConfig()
	doCfg, ok := cfg.(DoAppConfig)
	require.Equal(t, ok, true)
	require.Equal(t, doCfg.IAVLCacheSize, uint64(DefaultIAVLCacheSize))
	require.Equal(t, doCfg.IAVLDisableFastNode, IavlDisablefastNodeDefault)
}







