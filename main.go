package main

import "github.com/NethermindEth/juno/core/felt"

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
	var config Config // read from somwere

	provider := NewProvider(config.providerUrl)

	account := NewAccount(provider, &config.accountData)

    // Validator address should be part of config?
    // Validator staked funds should be also obtained here

	dispatcher := NewEventDispatcher()
	go dispatcher.Dispatch(provider, account, , )
	// I have to make sure this function closes at the end

	// ------

	// Here I need to subscribe to the blok headers and track them
	// Once I get a new header, I have to check if I should do an attestation for it
	// If yes, send an AttestRequired event with the necesary information

	// I've also need to check if the staked amount of the validator changes
	// The solution here is to subscribe to a possible event emitting
	// If it happens, send a StakeUpdated event with the necesary information

	// I've also would like to check the balance of the address from time to time to verify
	// that they have enough money for the next 10 attestation (value modifiable by user)
	// Once it goes below it, the console
	// should start giving warnings
	// This the least prio but we should implement nonetheless
}
