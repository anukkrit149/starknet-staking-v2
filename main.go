package main

import (
	"fmt"

	"github.com/NethermindEth/juno/core/crypto"
	"github.com/NethermindEth/juno/core/felt"
	rpcv8 "github.com/NethermindEth/juno/rpc/v8"
	"github.com/NethermindEth/starknet.go/account"
)

type AccountData struct {
	address string
	privKey string
	pubKey  string
}

// struct should be (un)marshallable
type Config struct {
	providerUrl string
	accountData AccountData
}

func main() {
	var config Config // read from somewhere

	provider := NewProvider(config.providerUrl)

	account := NewAccount(provider, &config.accountData)

	validator := Address(felt.Felt{})
	staked := staked(&validator)

	dispatcher := NewEventDispatcher()
	go dispatcher.Dispatch(provider, account, &validator, &staked)
	// I have to make sure this function closes at the end

	// ------

	// Initially, fetch necessary info
	//
	// Note 1 (attest info): No need to listen to real-time events, and fetching once per epoch should work,
	// as any important updates (ie, related to stake & epoch_length) are effective only from the next epoch!
	//
	// Note 2 (attest window): Depending on the expected behaviour of attestation window, we might have to listen to `AttestationWindowChanged` event
	attestationInfo, attestationWindow, blockNumberToAttestTo, err := fetchEpochInfo(account)
	if err != nil {
		// TODO: implement a retry mechanism ?
	}

	// Subscribe to the block headers
	blockHeaderChan := make(chan rpcv8.BlockHeader) // could maybe make it buffered to allow for margin?
	go subscribeToBlockHeaders(config.providerUrl, blockHeaderChan)

	for blockHeader := range blockHeaderChan {
		fmt.Println("Block header:", blockHeader)

		// Refetch epoch info on new epoch (validity guaranteed for 1 epoch even if updates are made)
		if *blockHeader.Number == attestationInfo.CurrentEpochStartingBlock+attestationInfo.EpochLen {
			previousEpochInfo := attestationInfo

			attestationInfo, attestationWindow, blockNumberToAttestTo, err = fetchEpochInfo(account)
			if err != nil {
				// TODO: implement a retry mechanism ?
			}

			// Sanity check
			if attestationInfo.EpochId != previousEpochInfo.EpochId+1 ||
				attestationInfo.CurrentEpochStartingBlock != previousEpochInfo.CurrentEpochStartingBlock+previousEpochInfo.EpochLen {
				fmt.Println("Wrong epoch change: from %s to %s", previousEpochInfo, attestationInfo)
				// TODO: what should we do ?
			}
		}

		// Send an attestation if necessary
		if *blockHeader.Number == blockNumberToAttestTo {
			// TODO: should actually send the `AttestRequired` event from: now + MIN_ATTESTATION_WINDOW !
			// and dispatcher can retry until: now + attestationWindow
			// --> might need to revise project architecture ?
			// --> maybe use a Map<now + MIN_ATTESTATION_WINDOW, blockHash> to check at each new block if need to send an event
			dispatcher.AttestRequired <- AttestRequired{
				blockHash: blockHeader.Hash,
				window:    attestationWindow,
			}
		}
	}

	// --> I think we don't need to listen to stake events, we can get it when fetching AttestationInfo
	//
	// I've also need to check if the staked amount of the validator changes
	// The solution here is to subscribe to a possible event emitting
	// If it happens, send a StakeUpdated event with the necesary information

	// I've also would like to check the balance of the address from time to time to verify
	// that they have enough money for the next 10 attestation (value modifiable by user)
	// Once it goes below it, the console
	// should start giving warnings
	// This the least prio but we should implement nonetheless

	// Should also track re-org and check if the re-org means we have to attest again or not
}

func fetchEpochInfo(account *account.Account) (AttestationInfo, uint8, uint64, error) {
	attestationInfo, attestInfoErr := fetchAttestationInfo(account)
	if attestInfoErr != nil {
		return AttestationInfo{}, 0, 0, attestInfoErr
	}

	attestationWindow, windowErr := fetchAttestationWindow(account)
	if windowErr != nil {
		return AttestationInfo{}, 0, 0, windowErr
	}

	blockNumberToAttestTo := computeBlockNumberToAttestTo(account, attestationInfo, attestationWindow)

	return attestationInfo, attestationWindow, blockNumberToAttestTo, nil
}

func computeBlockNumberToAttestTo(account *account.Account, attestationInfo AttestationInfo, attestationWindow uint8) uint64 {
	startingBlock := attestationInfo.CurrentEpochStartingBlock + attestationInfo.EpochLen

	// TODO: might be hash(stake, hash(epoch_id, address))
	// or should we use PoseidonArray instead ?
	hash := crypto.Poseidon(
		crypto.Poseidon(
			new(felt.Felt).SetBigInt(attestationInfo.Stake.Big()),
			new(felt.Felt).SetUint64(attestationInfo.EpochId),
		),
		account.AccountAddress,
	)
	// TODO: hash (felt) might not fit into a uint64 --> use big.Int in that case ?
	blockOffset := hash.Uint64() % (attestationInfo.EpochLen - uint64(attestationWindow))

	return startingBlock + blockOffset
}
