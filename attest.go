package main

import (
	"fmt"
	"github.com/sourcegraph/conc"
)

type Aaaaa struct {
	EpochInfo EpochInfo
}

// Main execution loop of the program. Listens to the blockchain and sends
// attest invoke when it's the right time
func Attest(config *Config) {
	provider := NewProvider(config.providerUrl)
	validatorAccount := NewValidatorAccount(provider, &config.accountData)
	dispatcher := NewEventDispatcher[*ValidatorAccount]()

	wg := conc.NewWaitGroup()
	defer wg.Wait()
	wg.Go(func() {
		dispatcher.Dispatch(&validatorAccount, make(map[BlockHash]AttestStatus))
	})

	// Subscribe to the block headers
	wsProvider, headersFeed := BlockHeaderSubscription(config.providerUrl)
	defer wsProvider.Close()
	defer close(headersFeed)

	attestInfo, attestationWindow, blockNumberToAttestTo, err := fetchEpochInfo(&validatorAccount)
	if err != nil {
		// If we fail at this point it means there is probably something wrong with the
		// configuration we might log the error and do a re-try just to make sure
	}

	pendingAttests := make(map[BlockNumber]AttestRequiredWithValidity)

	activeAttests := make(map[BlockNumber][]AttestRequired)

	for blockHeader := range headersFeed {
		fmt.Println("Block header:", blockHeader)

		// Re-fetch epoch info on new epoch (validity guaranteed for 1 epoch even if updates are made)
		if blockHeader.BlockNumber == attestInfo.CurrentEpochStartingBlock.Uint64()+attestInfo.EpochLen {
			previousEpochInfo := attestInfo

			attestInfo, attestationWindow, blockNumberToAttestTo, err = fetchEpochInfo(&validatorAccount)
			if err != nil {
				// TODO: implement a retry mechanism ?
			}

			// Sanity check
			if attestInfo.EpochId != previousEpochInfo.EpochId+1 ||
				attestInfo.CurrentEpochStartingBlock.Uint64() != previousEpochInfo.CurrentEpochStartingBlock.Uint64()+previousEpochInfo.EpochLen {
				// TODO: give more details concerning the epoch info
				fmt.Printf("Wrong epoch change: from %d to %d", previousEpochInfo.EpochId, attestInfo.EpochId)
				// TODO: what should we do ?
			}
		}

		SchedulePendingAttestations(
			blockHeader, blockNumberToAttestTo, pendingAttests, attestationWindow,
		)
		MovePendingAttestationsToActive(
			pendingAttests, activeAttests, BlockNumber(blockHeader.BlockNumber),
		)
		SendAllActiveAttestations(
			activeAttests, &dispatcher, BlockNumber(blockHeader.BlockNumber),
		)
	}

	// --> I think we don't need to listen to stake events, we can get it when fetching AttestationInfo
	//
	// I also need to check if the staked amount of the validator changes
	// The solution here is to subscribe to a possible event emitting
	// If it happens, send a StakeUpdated event with the necessary information

	// I'd also like to check the balance of the address from time to time to verify
	// that they have enough money for the next 10 attestations (value modifiable by user)
	// Once it goes below it, the console should start giving warnings
	// This the least prio but we should implement nonetheless

	// Should also track re-org and check if the re-org means we have to attest again or not
}
