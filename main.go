package main

import (
	"context"
	"log"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
)

// The general worklfow would be to:
// 1. Listen for a block where you have to make the attestation
//    - How much do you have to ask for a block number, do you have to ask every one second?
// 2. Make the attestation. Verify the transaction got included
//    - How do you verify a transaction got included, verifying the latest block
// Repeat until the next epoch
// How do we know there was an epoch change?

func main() {
	var providerURL string
	provider, err := rpc.NewProvider(providerURL)
	if err != nil {
		log.Fatalf("Error connecting to RPC provider at %s: %s", providerURL, err)
	}

	var publicKey string
	var privateKey big.Int
	ks := account.NewMemKeystore()
	ks.Put(publicKey, &privateKey)

	var accountAddr felt.Felt
	account, err := account.NewAccount(provider, &accountAddr, publicKey, ks, 1)
	if err != nil {
		log.Fatalf("Cannot create new account: %s", err)
	}

	// At this point, communication with the provider should be set

	nonce, err := account.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, account.AccountAddress)
	if err != nil {
		log.Fatalf("Cannot get account nonce: %s", err)
	}

	// Create the function call first

	invoke := rpc.BroadcastInvokev3Txn{
		InvokeTxnV3: rpc.InvokeTxnV3{
			Type:          rpc.TransactionType_Invoke,
			SenderAddress: account.AccountAddress,
			// Calldata: ,
			Version: rpc.TransactionV3,
			// Signature: ,
			Nonce: nonce,
			// ResourceBounds: ,
			// Tip: rpc.U64,
			// leaving the rest uninitialized
		},
	}

	// we shouldn't need to estimate the fee because we would use be doing the same transaction

}
