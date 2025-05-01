package config

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/starknet-staking-v2/validator/constants"
)

type chainId int

const (
	mainnet chainId = iota
	sepolia
	unknown
)

// An array containing the defaults values for Starknet networks
var defaults = [2]ContractAddresses{
	// Mainnet default values
	{
		Staking: "",
		Attest:  "",
	},
	// Sepolia default values
	{
		Staking: constants.SEPOLIA_STAKING_CONTRACT_ADDRESS,
		Attest:  constants.SEPOLIA_ATTEST_CONTRACT_ADDRESS,
	},
}

func chainIdFromStr(s string) chainId {
	if s == "mainnet" {
		return mainnet
	} else if s == "sepolia" {
		return sepolia
	}
	return unknown
}

type ContractAddresses struct {
	Staking string
	Attest  string
}

func (ca *ContractAddresses) SetDefaults(chainIdStr string) *ContractAddresses {
	// chainIdStr, err := provider.ChainID(context.Background())
	// if err != nil {
	// 	return err
	// }

	chainId := chainIdFromStr(chainIdStr)
	if chainId == unknown {
		return ca
	}

	defaultConfig := defaults[int(chainId)]
	if isZero(ca.Staking) {
		ca.Staking = defaultConfig.Staking
	}
	if isZero(ca.Attest) {
		ca.Attest = defaultConfig.Attest
	}

	return ca
}

func (ca *ContractAddresses) Check() error {
	if isZero(ca.Staking) {
		return errors.New("staking contract address is not set")
	}
	if isZero(ca.Attest) {
		return errors.New("attest contract address is not set")
	}
	return nil
}

type recalculate int

const (
	once recalculate = iota
	always
	never
)

type AttestOptions struct {
	Fee         uint64
	Recalculate string
}

func (o *AttestOptions) Check() error {
	if o.Recalculate != "once" && o.Recalculate != "always" {
		return fmt.Errorf(
			"expected recaluclate option \"once\" or \"always\""+
				"but got \"%s\" instead.",
			o.Recalculate,
		)
	}
	return nil
}

type StarknetConfig struct {
	ContractAddresses ContractAddresses
	AttestOptions     AttestOptions
}

func (c *StarknetConfig) SetDefaults(chainId string) *StarknetConfig {
	c.ContractAddresses.SetDefaults(chainId)
	return c
}

func (c *StarknetConfig) Check() error {
	if err := c.AttestOptions.Check(); err != nil {
		return err
	}
	return c.ContractAddresses.Check()
}
