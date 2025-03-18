package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/gorilla/websocket"
	"lukechampine.com/uint128"
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
	//  * staker operational address
	//  * public key
	//  * private key

	// place holder need to change
	var publicKey string
	var privateKey big.Int
	var accountAddr felt.Felt
	// ----------------------------

	ks := account.NewMemKeystore()
	ks.Put(publicKey, &privateKey)

	account, err := account.NewAccount(provider, &accountAddr, publicKey, ks, 2)
	if err != nil {
		log.Fatalf("Cannot create new account: %s", err)
	}
	return account
}

// TODO: might not need those 2 endpoints if we get info directly from staking contract

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

// Subscribe to block headers
func subscribeToBlockHeaders(providerUrl string, blockHeaderChan chan<- BlockHeader) {
	// Take the providerUrl parts (host & port) and build the ws url
	wsURL := "ws://" + "localhost" + ":" + "6061" + "/v0_8"

	// Connect to the WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn.Close()
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}

	// JSON RPC request to subscribe to new block headers
	subscribeMessage := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "starknet_subscribeNewHeads"
	}`

	// Send the subscription request
	if err := conn.WriteMessage(websocket.TextMessage, []byte(subscribeMessage)); err != nil {
		log.Fatal("Error sending subscription request:", err)
	}

	// TODO: implement a logger
	fmt.Println("Subscribed to new Starknet block headers...")

	// Listen for new block headers
	for {
		msgType, msgBytes, err := conn.ReadMessage()
		fmt.Println("Message type:", msgType)
		if err != nil {
			log.Fatal("Error reading message:", err)
		}
		// TODO: in logger
		// fmt.Println("New block header:", string(msgBytes))

		var response HeadersSubscriptionResponse
		if err = json.Unmarshal(msgBytes, &response); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
		} else {
			blockHeaderChan <- response.Params.Result
		}
	}
}

func fetchAttestationInfo(account *account.Account) AttestationInfo {
	stakingContractAddress, _ := new(felt.Felt).SetString(STAKING_CONTRACT_ADDRESS)

	functionCall := rpc.FunctionCall{
		ContractAddress:    stakingContractAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt("get_attestation_info_by_operational_address"),
		Calldata:           []*felt.Felt{account.AccountAddress},
	}

	result, err := account.Call(context.Background(), functionCall, rpc.BlockID{Tag: "latest"})
	if err != nil {
		log.Fatal("Error calling function:", err)
		//TODO: retry ?
	}

	if len(result) != 5 {
		log.Fatal("Invalid response from function call")
		//TODO: retry ?
	}

	// TODO: verify once endpoint is available
	stake := result[1].Bits()
	return AttestationInfo{
		StakerAddress:             Address(*result[0]),
		Stake:                     uint128.Uint128{Lo: stake[0], Hi: stake[1]},
		EpochLen:                  result[2].Uint64(),
		EpochId:                   result[3].Uint64(),
		CurrentEpochStartingBlock: result[4].Uint64(),
	}
}

func fetchAttestationWindow(account *account.Account) uint8 {
	attestationContractAddress, _ := new(felt.Felt).SetString(ATTESTATION_CONTRACT_ADDRESS)

	result, err := account.Call(
		context.Background(),
		rpc.FunctionCall{
			ContractAddress:    attestationContractAddress,
			EntryPointSelector: utils.GetSelectorFromNameFelt("attestation_window"),
			Calldata:           []*felt.Felt{},
		},
		rpc.BlockID{Tag: "latest"},
	)

	if err != nil {
		log.Fatal("Error calling function:", err)
		//TODO: retry ?
	}

	if len(result) != 1 {
		log.Fatal("Invalid response from function call")
		//TODO: retry ?
	}

	return uint8(result[0].Uint64())
}

func fetchValidatorBalance(account *account.Account) big.Int {
	sepoliaStrkTokenAddress, _ := new(felt.Felt).SetString(SEPOLIA_TOKENS.Strk)

	result, err := account.Call(
		context.Background(),
		rpc.FunctionCall{
			ContractAddress:    sepoliaStrkTokenAddress,
			EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
			Calldata:           []*felt.Felt{account.AccountAddress},
		},
		rpc.BlockID{Tag: "latest"},
	)

	if err != nil {
		log.Fatal("Error calling function:", err)
		//TODO: retry ?
	}

	if len(result) != 1 {
		log.Fatal("Invalid response from function call")
		//TODO: retry ?
	}

	return *result[0].BigInt(&big.Int{})
}
