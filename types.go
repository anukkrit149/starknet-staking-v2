package main

import (
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
)

type Address felt.Felt

type Balance struct{}

// HeadersSubscriptionResponse is the response from the subscription to new block headers
type HeadersSubscriptionResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result         BlockHeader `json:"result"`
		SubscriptionID uint64      `json:"subscription_id"`
	} `json:"params"`
}

type BlockHeader struct {
	Hash             felt.Felt `json:"block_hash"`
	ParentHash       felt.Felt `json:"parent_hash"`
	Number           uint64    `json:"block_number"`
	NewRoot          felt.Felt `json:"new_root"`
	Timestamp        uint64    `json:"timestamp"`
	SequencerAddress Address   `json:"sequencer_address"`
	L1GasPrice       GasPrice  `json:"l1_gas_price"`
	L1DataGasPrice   GasPrice  `json:"l1_data_gas_price"`
	L1DAMode         string    `json:"l1_da_mode"` // make it an enum here ?
	StarknetVersion  string    `json:"starknet_version"`
	L2GasPrice       GasPrice  `json:"l2_gas_price"`
}

type GasPrice struct {
	PriceInFri string `json:"price_in_fri"`
	PriceInWei string `json:"price_in_wei"`
}

// AttestationInfo is the response from the get_attestation_info_by_operational_address
type AttestationInfo struct {
	StakerAddress             Address `json:"staker_address"`
	Stake                     big.Int `json:"stake"` // uin128, maybe use felt? or custom uint128 struct ?
	EpochLen                  uint64  `json:"epoch_len"`
	EpochId                   uint64  `json:"epoch_id"`
	CurrentEpochStartingBlock uint64  `json:"current_epoch_starting_block"`
}
