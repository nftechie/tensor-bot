package main

import (
	"context"
	"fmt"
	"os"
	"tensor/pkg/bot"
	"tensor/pkg/tensor"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	// Init Solana
	solanaClient := rpc.New(rpc.MainNetBeta_RPC)
	wsClient, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}

	keypath := os.Getenv("TENSOR_BOT_KEYPATH")
	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(keypath)
	if err != nil {
		panic(err)
	}

	owner := os.Getenv("TENSOR_BOT_PUBKEY")
	// owner := "2tP2YSMKCFBVEFk6bZzsjDwn3LmgYJvFuLxyXWV4eHPi" // trading wallet

	// Init Tensor
	tensorService := tensor.NewTensorService()

	bots := []*bot.Bot{
		bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "fe16c678-aaa9-4d4c-b082-bbc33110f88e", owner),
		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "dcf741e2-9c58-45a4-b9cd-3b540b0ddb15", owner), // cheap collection

		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "bccd337d-0d8d-4a1a-9703-2ff922253a86", owner),

		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "00d69fef-4c5f-4b71-bd34-24222696c80e", owner),

		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "e60dd60b-ac57-4bbe-baa1-20a13f94f294", owner),

		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "4c582543-a59d-46f5-a39c-1b0cca97979e", owner), // baby_tensos
		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "magicticket", owner),                          // magicticket
		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "66b8ffbe-b209-4f14-a050-5a85ea9c7c3f", owner), // fooks
		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "ce9789f9-1dd7-47d9-86eb-4a5aeb736fdc", owner), // bonkznft
		// bot.NewBot(tensorService, solanaClient, wsClient, privateKey, "7f5f0c7d-1639-4b7f-a914-85e2bad9111e", owner), // luckylouie
	}

	// Run bots
	go func() {
		for {
			for _, bot := range bots {
				fmt.Printf("Running bot %s\n", bot.TargetCollection)
				err = bot.Run()
				if err != nil {
					fmt.Errorf("bot execution error: %v", err)
				}

				// Short sleep between bots
				time.Sleep(2 * time.Second)
			}

			// time.Sleep(15 * time.Second)
		}
	}()

	// Keep the main program running
	select {}
}
