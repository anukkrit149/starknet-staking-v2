package main

import (
	"context"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
)

// Keep track of the stake amount through the different epochs
// Because there is a latency the most current stake doesn't need to match
// to the current epoch attestation
type stakedAmount struct{}

func NewStakedAmount() stakedAmount {
	return stakedAmount{}
}

func (s *stakedAmount) Update(newStake *StakeUpdated) {

}

func (s *stakedAmount) Get(epoch uint64) {

}

// requires filling with the right values
type AttestRequired struct{}

// requires filling with the right values
type StakeUpdated struct{}

type EventDispatcher struct {
	StakeUpdated   chan StakeUpdated
	AttestRequired chan AttestRequired
}

func NewEventDispatcher() EventDispatcher {
	return EventDispatcher{
		AttestRequired: make(chan AttestRequired),
		StakeUpdated:   make(chan StakeUpdated),
	}
}

func (d *EventDispatcher) Dispatch(
	provider *rpc.Provider,
	account *account.Account,
	stakedAmount felt.Felt,
	validatorAddress felt.Felt,
) {
	var currentEpoch uint64
	stakedAmountPerEpoch := NewStakedAmount()

	select {
	case event, ok := <-d.AttestRequired:
		if !ok {
			break
		}
		stakedAmountPerEpoch.Get(currentEpoch)
		resp, err := invokeAttest(provider, account, &event)
		if err != nil {
			// Here you want to wait a few seconds and repeat the attestation
		}
		// do something with response. On a separate go function
		// Have it checked that is included in pending block
		// Have it checked that it was included in the latest block (success, close the goroutine)
		// ---
		// If not included, what do we do, TBD
		// ---
		// If resp reverted give an error

	case event, ok := <-d.StakeUpdated:
		if !ok {
			break
		}
		stakedAmountPerEpoch.Update(&event)
	}
}

func invokeAttest(
	provider *rpc.Provider, account *account.Account, attest *AttestRequired,
) (*rpc.TransactionResponse, error) {
	// Todo this might be worth doing async
	nonce, err := nonce(account)
	if err != nil {
		return nil, err
	}

	fnCall := rpc.FunctionCall{
		// ContractAddress: , // Some predetermined address
		// EntryPointSelector: , // Some predetermined selector
		// Calldata: , // Some calldata which is not predetermined
	}

	invokeCalldata, err := account.FmtCalldata([]rpc.FunctionCall{fnCall})
	if err != nil {
		return nil, err
	}

	invoke := rpc.BroadcastInvokev3Txn{
		InvokeTxnV3: rpc.InvokeTxnV3{
			Type:          rpc.TransactionType_Invoke,
			SenderAddress: account.AccountAddress,
			Calldata:      invokeCalldata,
			Version:       rpc.TransactionV3,
			// Signature: , // Set during signing below
			Nonce: nonce,
			// ResourceBounds: ,
			// Tip: rpc.U64, // Investigate if this is applicable, perhaps it can be a way of
			//               // prioritizing this transaction if there is congestion
			// PayMasterData: , // I don't know if this is applicable, investigate
			// AccountDeploymentData: , // It shouldn't be required
			// NonceDataMode: , // Investigate what goes here
			// FeeMode: , // Investigate
		},
	}

	err = account.SignInvokeTransaction(context.Background(), &invoke.InvokeTxnV3)
	if err != nil {
		return nil, err
	}

	return account.SendTransaction(context.Background(), invoke)
}
