package main

import (
	"context"
	"log"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
)

func NewProvider(providerUrl string) *rpc.Provider {
	provider, err := rpc.NewProvider(providerUrl)
	if err != nil {
		log.Fatalf("Error connecting to RPC provider at %s: %s", providerUrl, err)
	}
	return provider
}

func NewAccount(provider *rpc.Provider, accountData *AccountData) *account.Account {
	// accountData should contain the information required:
	//  * address
	//  * public key
	//  * private key

	// place holder need to change
	var publicKey string
	var privateKey big.Int
	var accountAddr felt.Felt
	// ----------------------------

	ks := account.NewMemKeystore()
	ks.Put(publicKey, &privateKey)

	account, err := account.NewAccount(provider, &accountAddr, publicKey, ks, 1)
	if err != nil {
		log.Fatalf("Cannot create new account: %s", err)
	}
	return account
}

// Given a validator address returns the staked amount
func staked(address *Address) Balance {
	return Balance{}
}

// Given an address returns it's balance
func balance(address *Address) Balance {
	return Balance{}
}

// Given an account, returns it's nonce
func nonce(account *account.Account) (*felt.Felt, error) {
	return account.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, account.AccountAddress)
}
