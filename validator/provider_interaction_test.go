package validator_test

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/juno/utils"
	"github.com/NethermindEth/starknet-staking-v2/validator"
	main "github.com/NethermindEth/starknet-staking-v2/validator"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	logger := utils.NewNopZapLogger()

	t.Run("Error creating provider", func(t *testing.T) {
		providerUrl := "wrong url"

		provider, err := main.NewProvider(providerUrl, logger)

		require.Nil(t, provider)
		expectedErrorMsg := fmt.Sprintf(`cannot create RPC provider at %s`, providerUrl)
		require.ErrorContains(t, err, expectedErrorMsg)
	})

	t.Run("Error connecting to provider", func(t *testing.T) {
		providerUrl := "http://localhost:1234"

		provider, err := main.NewProvider(providerUrl, logger)

		require.Nil(t, provider)

		expectedErrorMsg := fmt.Sprintf(`cannot connect to RPC provider at %s`, providerUrl)
		require.ErrorContains(t, err, expectedErrorMsg)
	})

	envVars, err := validator.LoadEnv(t)
	loadedEnvVars := err == nil
	if loadedEnvVars {
		t.Run("Successful provider creation", func(t *testing.T) {
			if err != nil {
				t.Skip(err)
			}

			provider, err := main.NewProvider(envVars.HttpProviderUrl, logger)

			// Cannot deeply compare 2 providers (comparing channels does not works)
			require.NotNil(t, provider)
			require.Nil(t, err)
		})
	} else {
		t.Logf("Ignoring tests that require env variables: %s", err)
	}
}

func TestBlockHeaderSubscription(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	logger := utils.NewNopZapLogger()

	t.Run("Error creating provider", func(t *testing.T) {
		wsProviderUrl := "wrong url"
		wsProvider, headerFeed, clientSubscription, err := main.SubscribeToBlockHeaders(wsProviderUrl, logger)

		require.Nil(t, wsProvider)
		require.Nil(t, headerFeed)
		require.Nil(t, clientSubscription)
		expectedErrorMsg := fmt.Sprintf(`dialing WS provider at %s`, wsProviderUrl)
		require.ErrorContains(t, err, expectedErrorMsg)
	})

	// Cannot test error when subscribing to new block headers

	envVars, err := validator.LoadEnv(t)
	if loadedEnvVars := err == nil; loadedEnvVars {
		t.Run("Successfully subscribing to new block headers", func(t *testing.T) {
			wsProvider, headerChannel, clientSubscription, err := main.SubscribeToBlockHeaders(envVars.WsProviderUrl, logger)

			require.NotNil(t, wsProvider)
			require.NotNil(t, headerChannel)
			require.NotNil(t, clientSubscription)
			require.Nil(t, err)

			wsProvider.Close()
			close(headerChannel)
		})
	} else {
		t.Logf("Ignoring tests that require env variables: %s", err)
	}
}
