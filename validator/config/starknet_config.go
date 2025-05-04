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
	if s == "SN_MAINNET" {
		return mainnet
	} else if s == "SN_SEPOLIA" {
		return sepolia
	}
	return unknown
}

type ContractAddresses struct {
	Staking string
	Attest  string
}

func (ca *ContractAddresses) SetDefaults(chainIdStr string) *ContractAddresses {
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

func (ca *ContractAddresses) String() string {
	return fmt.Sprintf(`{
        staking contract address: %s,
        attestation contract address: %s,
    }`,
		ca.Staking,
		ca.Attest,
	)
}

type StarknetConfig struct {
	ContractAddresses ContractAddresses
	AttestOptions     string
}

func (c *StarknetConfig) SetDefaults(chainId string) *StarknetConfig {
	c.ContractAddresses.SetDefaults(chainId)
	return c
}

func (c *StarknetConfig) Check() error {
	return c.ContractAddresses.Check()
}
