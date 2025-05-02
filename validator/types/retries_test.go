package types_test

import (
	"strconv"
	"testing"

	"github.com/NethermindEth/starknet-staking-v2/validator/types"
	"github.com/stretchr/testify/require"
)

func TestRetriesFromString(t *testing.T) {
	invalidStr := []string{
		"infinte",
		"somtehing",
		"",
	}
	for _, s := range invalidStr {
		r, err := types.RetriesFromString(s)
		require.Zero(t, r)
		require.ErrorContains(t, err, "cannot create retries from string")
		require.ErrorContains(t, err, s)
	}

	r, err := types.RetriesFromString("0")
	require.Zero(t, r)
	require.ErrorContains(t, err, "value should be greater or equal than one")

	correctStr := []string{
		"1",
		"123",
		"infinite",
		"1999",
	}
	for _, s := range correctStr {
		r, err := types.RetriesFromString(s)
		require.NoError(t, err)
		require.Equal(t, s, r.String())
	}
}

func TestRetryIsZero(t *testing.T) {
	r := types.NewRetries()
	require.Equal(t, r.String(), "infinite")
	r.Sub()
	require.Equal(t, r.String(), "infinite")

	const val = 3
	r.Set(val)
	for i := val; i > 0; i-- {
		require.Equal(t, r.String(), strconv.FormatInt(int64(i), 10))
		r.Sub()
	}
	require.True(t, r.IsZero())
}
