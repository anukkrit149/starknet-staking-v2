package main

import (
	"context"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
)

const defaultAttestDelay = 10

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
	validator *Address,
	stakedAmount *Balance,
) {
	var currentEpoch uint64
	stakedAmountPerEpoch := NewStakedAmount()

	for {
		select {
		case event, ok := <-d.AttestRequired:
			if !ok {
				break
			}
			stakedAmountPerEpoch.Get(currentEpoch)
			resp, err := invokeAttest(account, &event)
			if err != nil {
				// throw a detailed error of what happened
				// and
				go repeatAttest(d.AttestRequired, event, defaultAttestDelay)
			}
			go trackAttest(provider, d.AttestRequired, event, resp)

		case event, ok := <-d.StakeUpdated:
			if !ok {
				break
			}
			stakedAmountPerEpoch.Update(&event)
		}
	}
}

func invokeAttest(
	account *account.Account, attest *AttestRequired,
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
			// PayMasterData: , // I don't know if this is applicable, investigate. Maybe not in v1
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

// Repeat an Attest event after the `delay` in seconds
func repeatAttest(attestChan chan AttestRequired, event AttestRequired, delay uint64) {
	time.Sleep(time.Second * time.Duration(delay))
	attestChan <- event
}

// do something with response
// Have it checked that is included in pending block
// Have it checked that it was included in the latest block (success, close the goroutine)
// ---
// If not included, then what was the reason? Try to include it again
// ---
// If the transaction was actually reverted log it and repeat the attestation
func trackAttest(
	provider *rpc.Provider,
	attestChan chan AttestRequired,
	event AttestRequired,
	txResp *rpc.TransactionResponse,
) {
	startTime := time.Now()
	txStatus, err := trackTransactionStatus(provider, txResp.TransactionHash)
	if err != nil {
		// log exactly what's the error

		elapsedTime := time.Now().Sub(startTime).Seconds()
		var attestDelay uint64
		if elapsedTime < defaultAttestDelay {
			attestDelay = defaultAttestDelay - uint64(elapsedTime)
		}
		repeatAttest(attestChan, event, attestDelay)
		return
	}

	if txStatus.FinalityStatus == rpc.TxnStatus_Rejected {
		// log exactly the rejection and why was it

		repeatAttest(attestChan, event, defaultAttestDelay)
		return
	}

	// if we got here, then the transaction status was accepted
}

func trackTransactionStatus(provider *rpc.Provider, txHash *felt.Felt) (*rpc.TxnStatusResp, error) {
	elapsedSeconds := 0
	const maxSeconds = defaultAttestDelay
	for elapsedSeconds < defaultAttestDelay {
		txStatus, err := provider.GetTransactionStatus(context.Background(), txHash)
		if err != nil {
			return nil, err
		}
		if txStatus.FinalityStatus != rpc.TxnStatus_Received {
			return txStatus, nil
		}
		time.Sleep(time.Second)
		elapsedSeconds++
	}

	// If we are here, it means the transaction didn't change it's status for a long time.
	// Should we track again?
	// Should we have a finite number of tries?
	return trackTransactionStatus(provider, txHash)
}
