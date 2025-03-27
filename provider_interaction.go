package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NethermindEth/starknet.go/rpc"
)

func NewProvider(providerUrl string) *rpc.Provider {
	provider, err := rpc.NewProvider(providerUrl)
	if err != nil {
		log.Fatalf("Error connecting to RPC provider at %s: %s", providerUrl, err)
	}
	return provider
}

func subscribeToBlockHeader(providerUrl string, blockHeaderFeed chan<- *rpc.BlockHeader) {
	fmt.Println("Starting websocket connection...")

	// Take the providerUrl parts (host & port) and build the ws url
	wsProviderUrl := "ws://" + "localhost" + ":" + "6061" + "/v0_8"

	// Initialize connection to WS provider
	wsClient, err := rpc.NewWebsocketProvider(wsProviderUrl)
	if err != nil {
		log.Fatalf("Error dialing the WS provider: %s", err)
	}
	defer wsClient.Close()       // Close the WS client when the program finishes
	defer close(blockHeaderFeed) // Close the headers channel

	fmt.Println("Established connection with the client")

	sub, err := wsClient.SubscribeNewHeads(context.Background(), blockHeaderFeed, nil)
	if err != nil {
		log.Fatalf("Error subscribing to new block headers: %s", err)
	}

	fmt.Println("Successfully subscribed to the node. Subscription ID:", sub.ID())
}
