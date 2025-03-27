package main

const (
	STAKING_CONTRACT_ADDRESS          = "0x123"
	ATTEST_CONTRACT_ADDRESS           = "0x123"
	STRK_CONTRACT_ADDRESS             = "0x04718f5a0fc34cc1af16a1cdee98ffb20c31f5cd61d6ab07201858f4287c938d"
	ETH_CONTRACT_ADDRESS              = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	MIN_ATTESTATION_WINDOW    uint64  = 10
	FEE_ESTIMATION_MULTIPLIER float64 = 1.5
)

var (
	StakingContract   = AddressFromString(STAKING_CONTRACT_ADDRESS)
	AttestContract    = AddressFromString(ATTEST_CONTRACT_ADDRESS)
	StrkTokenContract = AddressFromString(STRK_CONTRACT_ADDRESS)
)
