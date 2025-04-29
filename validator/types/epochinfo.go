package types

import (
	"encoding/json"

	"lukechampine.com/uint128"
)

type EpochInfo struct {
	StakerAddress             Address         `json:"staker_address"`
	Stake                     uint128.Uint128 `json:"stake"`
	EpochLen                  uint64          `json:"epoch_len"`
	EpochId                   uint64          `json:"epoch_id"`
	CurrentEpochStartingBlock BlockNumber     `json:"current_epoch_starting_block"`
}

func (e *EpochInfo) String() string {
	jsonData, err := json.Marshal(e)
	if err != nil {
		panic("cannot marshall epoch info")
	}

	return string(jsonData)
}
