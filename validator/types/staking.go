package types

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NethermindEth/starknet-staking-v2/validator/config"
	"lukechampine.com/uint128"
)

type AttestRequired struct {
	BlockHash BlockHash
}

type AttestInfo struct {
	TargetBlock     BlockNumber
	TargetBlockHash BlockHash
	WindowStart     BlockNumber
	WindowEnd       BlockNumber
}

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

type recalculate int

const (
	once recalculate = iota
	always
	never
)

type AttestFee struct {
	recalculate recalculate
	value       uint64
}

func AttestFeeFromString(attestOption string) (AttestFee, error) {
	if attestOption == "always" {
		return AttestFee{
			recalculate: always,
			value:       0,
		}, nil
	}
	if attestOption == "once" {
		return AttestFee{
			recalculate: once,
			value:       0,
		}, nil
	}
	val, err := strconv.ParseUint(attestOption, 10, 64)
	if err != nil {
		return AttestFee{},
			fmt.Errorf(
				"cannot parse attest options :"+
					" `%s` is not a valid number nor a valid option",
				attestOption,
			)
	}
	return AttestFee{
		recalculate: never,
		value:       val,
	}, nil
}

func (a *AttestFee) Get() uint64 {
	switch a.recalculate {
	case always:
		// do the recalculation
		return 1000
	case once:
		// do the recalculation
		a.recalculate = never
		return a.value
	case never:
		return a.value
	default:
		panic("Unknown recalculate option for AttestFee")
	}
}

type ValidationContracts struct {
	Staking Address
	Attest  Address
}

func ValidationContractsFromAddresses(ca *config.ContractAddresses) ValidationContracts {
	return ValidationContracts{
		Attest:  AddressFromString(ca.Attest),
		Staking: AddressFromString(ca.Staking),
	}
}

func (c *ValidationContracts) String() string {
	return fmt.Sprintf(`{
        Staking contract address: %s,
        Attestation contract address: %s,
    }`,
		c.Staking.String(),
		c.Attest.String(),
	)
}
