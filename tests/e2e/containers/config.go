package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	InitRepository string
	InitTag        string

	DoRepository string
	DoTag        string

	RelayerRepository string
	RelayerTag        string
}

//nolint:deadcode
const (
	// Current Git branch do repo/version. It is meant to be built locally.
	// This image should be pre-built with `make docker-build-debug` either in CI or locally.
	CurrentBranchDoRepository = "do"
	CurrentBranchDoTag        = "debug"
	// Hermes repo/version for relayer
	relayerRepository = "informalsystems/hermes"
	relayerTag        = "1.13.1"
)

// Returns ImageConfig needed for running e2e test.
func NewImageConfig() ImageConfig {
	config := ImageConfig{
		RelayerRepository: relayerRepository,
		RelayerTag:        relayerTag,
	}

	config.DoRepository = CurrentBranchDoRepository
	config.DoTag = CurrentBranchDoTag
	return config
}







