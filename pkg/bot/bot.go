package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"tensor/pkg/tensor"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gorilla/websocket"
)

const MIN_BID_SPREAD = 0.0
const MIN_SALES_THRESHOLD = 0

type BotState struct {
	Mint        string
	GrossAmount string
}

type Bot struct {
	tensorService    *tensor.TensorService
	solanaClient     *rpc.Client
	wsClient         *ws.Client
	privateKey       solana.PrivateKey
	TargetCollection string
	owner            string
	botState         BotState
}

type TensorTxPayload struct {
	Payload struct {
		Data struct {
			NewTransactionTV2 struct {
				Tx struct {
					GrossAmount   string  `json:"grossAmount"`
					MintOnchainId string  `json:"mintOnchainId"`
					TxId          string  `json:"txId"`
					TxType        string  `json:"txType"`
					BuyerId       *string `json:"buyerId"`
					SellerId      string  `json:"sellerId"`
					Source        string  `json:"source"`
				} `json:"tx"`
			} `json:"newTransactionTV2"`
		} `json:"data"`
	} `json:"payload"`
}

func (b *Bot) reconnectWebsocket() (*websocket.Conn, error) {
	d := websocket.DefaultDialer
	d.Subprotocols = []string{"graphql-transport-ws"}

	c, _ /* resp */, err := d.Dial("wss://api.tensor.so/graphql", nil)
	if err != nil {
		return nil, err
	}

	// Send the initial connection payload
	initPayload := map[string]interface{}{
		"type": "connection_init",
		"payload": map[string]string{
			"x-tensor-api-key": "e3a7eaf1-abc4-4fb7-addd-5755a6bb197d",
		},
	}

	err = c.WriteJSON(initPayload)
	if err != nil {
		return nil, err
	}

	// Wait for the ack
	_, _, err = c.ReadMessage()
	if err != nil {
		return nil, err
	}

	// Assuming the ack is received, send the subscribe payload
	subscriptionPayload := map[string]interface{}{
		"id":   "e386777a-329e-470d-84d4-b51457db7cff",
		"type": "subscribe",
		"payload": map[string]interface{}{
			"operationName": "NewTransactionTV2",
			"query":         "subscription NewTransactionTV2($slug: String!) {\n  newTransactionTV2(slug: $slug) {\n    tx {\n      grossAmount\n      mintOnchainId\n      txAt\n      txId\n      txType\n      buyerId\n      sellerId\n      source\n    }\n  }\n}",
			"variables": map[string]string{
				"slug": b.TargetCollection,
			},
		},
	}

	err = c.WriteJSON(subscriptionPayload)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (b *Bot) initWs() {
	go func() {
		// Connection logic for the websocket
		c, err := b.reconnectWebsocket()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c, err = b.reconnectWebsocket()
				if err != nil {
					fmt.Println(err)
					return
				}
				defer c.Close()
				continue
			}

			var payload TensorTxPayload

			err = json.Unmarshal([]byte(message), &payload)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Filter for sales from bot wallet
			if payload.Payload.Data.NewTransactionTV2.Tx.TxType == "SWAP_SELL_NFT" && *payload.Payload.Data.NewTransactionTV2.Tx.BuyerId == b.owner {
				spew.Dump(payload)

				// Bot was sold an NFT
				mint := payload.Payload.Data.NewTransactionTV2.Tx.MintOnchainId
				price := payload.Payload.Data.NewTransactionTV2.Tx.GrossAmount

				b.botState.GrossAmount = price
				b.botState.Mint = mint

				fmt.Printf("bot purchased mint %s for %s lamports\n", mint, price)
			} else if payload.Payload.Data.NewTransactionTV2.Tx.TxType == "SALE_BUY_NOW" && payload.Payload.Data.NewTransactionTV2.Tx.SellerId == b.owner {
				spew.Dump(payload)

				// Bot sold an NFT
				mint := payload.Payload.Data.NewTransactionTV2.Tx.MintOnchainId
				price := payload.Payload.Data.NewTransactionTV2.Tx.GrossAmount

				b.botState.GrossAmount = ""
				b.botState.Mint = ""

				fmt.Printf("bot sold mint %s for %s lamports\n", mint, price)
			}
		}
	}()
}

func NewBot(tensorService *tensor.TensorService, solanaClient *rpc.Client, wsClient *ws.Client, privateKey solana.PrivateKey, targetCollection string, owner string) *Bot {
	bot := &Bot{
		tensorService:    tensorService,
		solanaClient:     solanaClient,
		wsClient:         wsClient,
		privateKey:       privateKey,
		TargetCollection: targetCollection,
		owner:            owner,
		botState:         BotState{
			// Mint:        "8wQqyMUM81vrAcXzGUiWbEPmDnhU8YK2wUrbMCBbiUMB",
			// GrossAmount: "10500001",
		}, // TODO: can this be pulled from chain?
	}

	bot.initWs()

	return bot
}

func (b *Bot) signAndSignTx(bytes []byte) (*solana.Signature, error) {
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(bytes))
	if err != nil {
		return nil, err
	}

	tx.Signatures = []solana.Signature{}

	out, err := b.solanaClient.GetLatestBlockhash(
		context.Background(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return nil, err
	}

	tx.Message.RecentBlockhash = out.Value.Blockhash

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if b.privateKey.PublicKey().Equals(key) {
				return &b.privateKey
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Send transaction, and wait for confirmation:
	sig, err := confirm.SendAndConfirmTransaction(
		context.TODO(),
		b.solanaClient,
		b.wsClient,
		tx,
	)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func dataToBytes(data []int) []byte {
	bytes := make([]byte, len(data))
	for i, v := range data {
		bytes[i] = byte(v)
	}

	return bytes
}

func (b *Bot) closePool(pool string) (*solana.Signature, error) {
	closePoolResp, err := b.tensorService.ClosePool(pool)
	if err != nil {
		return nil, err
	}

	if len(closePoolResp.Data.TswapClosePoolTx.Txs) == 0 {
		return nil, errors.New("no tx returned")
	}

	bytes := dataToBytes(closePoolResp.Data.TswapClosePoolTx.Txs[0].Tx.Data)

	sig, err := b.signAndSignTx(bytes)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (b *Bot) poolDepositSol(pool string, lamports string) (*solana.Signature, error) {
	poolDepositSol, err := b.tensorService.PoolDepositSol(pool, lamports)
	if err != nil {
		return nil, err
	}

	if len(poolDepositSol.Data.TswapDepositWithdrawSolTx.Txs) == 0 {
		return nil, errors.New("no tx returned")
	}

	bytes := dataToBytes(poolDepositSol.Data.TswapDepositWithdrawSolTx.Txs[0].Tx.Data)

	sig, err := b.signAndSignTx(bytes)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (b *Bot) createPool(lamports string) (*UserPool, error) {
	createPoolResponse, err := b.tensorService.CreatePool(b.TargetCollection, b.owner, lamports)
	if err != nil {
		return nil, err
	}

	pool := createPoolResponse.Data.TswapInitPoolTx.Pool

	if len(createPoolResponse.Data.TswapInitPoolTx.Txs) == 0 {
		return nil, errors.New("no tx returned")
	}

	// Create pool
	{
		bytes := dataToBytes(createPoolResponse.Data.TswapInitPoolTx.Txs[0].Tx.Data)

		sig, err := b.signAndSignTx(bytes)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Created pool %s: tx: %s\n", pool, sig.String())
	}

	// Deposit sol
	{
		txSig, err := b.poolDepositSol(pool, lamports)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Deposited sol into pool %s: tx: %s\n", pool, txSig.String())
	}

	return &UserPool{PoolAddress: pool, BidPrice: lamports}, nil
}

type UserPool struct {
	PoolAddress string
	BidPrice    string
}

func (b *Bot) getUserPool() (*UserPool, error) {
	var userPool UserPool

	// Get user pools
	userPools, err := b.tensorService.UserPools(b.owner)
	if err != nil {
		return nil, err
	}

	// Check if there's an open pool for the collection we're trading
	for _, pool := range userPools.Data.UserTswapOrders {
		if pool.Slug == b.TargetCollection {
			userPool.PoolAddress = pool.Pool.Address
			userPool.BidPrice = pool.Pool.StartingPrice

			if pool.Pool.TakerSellCount > 0 {
				// Bot purchased an NFT, close the pool
				txSig, err := b.closePool(pool.Pool.Address)
				if err != nil {
					return nil, err
				}

				fmt.Printf("Closed pool %s: tx: %s\n", pool.Pool.Address, txSig.String())
			}

			// Sometimes the initial pool deposit may fail, so we should check here to make sure the pool has enough SOL balance.
			if pool.Pool.SolBalance == "0" {
				// Deposit sol

				txSig, err := b.poolDepositSol(userPool.PoolAddress, userPool.BidPrice)
				if err != nil {
					return nil, err
				}

				fmt.Println("Deposited sol into pool: tx: ", txSig.String())
			}

			return &userPool, nil
		}
	}

	return nil, errors.New("not found")
}

func addOneLamport(input string) (string, error) {
	lamports, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return "", err
	}

	lamports += 1337

	return strconv.FormatInt(lamports, 10), nil
}

func subtractOneLamport(input string) (string, error) {
	lamports, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return "", err
	}

	lamports -= 1337

	return strconv.FormatInt(lamports, 10), nil
}

// shouldPlaceBid determines whether a bid should be placed based on spread percentage.
func shouldPlaceBid(collectionStats *tensor.CollectionStatsResponse) (bool, error) {
	// Convert string prices to int64
	buyPrice, err := strconv.ParseInt(collectionStats.Data.InstrumentTV2.StatsV2.BuyNowPrice, 10, 64)
	if err != nil {
		return false, err
	}
	sellPrice, err := strconv.ParseInt(collectionStats.Data.InstrumentTV2.StatsV2.SellNowPrice, 10, 64)
	if err != nil {
		return false, err
	}

	// Calculate spread
	spread := buyPrice - sellPrice

	// Calculate spread as a percentage of sell price
	spreadPercent := (float64(spread) / float64(sellPrice)) * 100

	recentSales := collectionStats.Data.InstrumentTV2.StatsV2.Sales1h

	// Check if spread percentage meets our criteria
	return spreadPercent >= MIN_BID_SPREAD && recentSales > MIN_SALES_THRESHOLD, nil
}

func (b *Bot) executeBidLogic() error {
	activeOrders, err := b.tensorService.GetActiveOrders(b.TargetCollection)
	if err != nil {
		return err
	}

	nonNullOrders := make([]tensor.TswapOrder, 0)
	for _, order := range activeOrders.Data.TswapOrders {
		if order.SellNowPrice != nil {
			nonNullOrders = append(nonNullOrders, order)
		}
	}

	sort.Slice(nonNullOrders, func(i, j int) bool {
		priceI, _ := strconv.ParseInt(*nonNullOrders[i].SellNowPrice, 10, 64)
		priceJ, _ := strconv.ParseInt(*nonNullOrders[j].SellNowPrice, 10, 64)
		return priceI > priceJ
	})

	bids := make([]string, 0)

	for _, order := range nonNullOrders {
		bids = append(bids, *order.SellNowPrice)
	}

	// Debug order book
	// for i := 0; i < len(asks); i++ {
	// 	fmt.Println(asks[len(asks)-i-1])
	// }

	// fmt.Println(strings.Repeat("-", 10))

	// for i := 0; i < 10; i++ {
	// 	fmt.Println(bids[i])
	// }

	// fmt.Println("")

	// TODO: Determine bid price
	// For now, let's just do something really really simple and keep the price at a fixed depth

	targetIndex := 0
	bidPrice := *nonNullOrders[0].SellNowPrice

	// Only add one to top bid if it's not the bot's bid
	if nonNullOrders[targetIndex].OwnerAddress != b.owner {
		bidPrice, err = addOneLamport(*nonNullOrders[targetIndex].SellNowPrice)
		if err != nil {
			return err
		}
	}

	// Check to make sure collection has enough recent volume
	collectionStats, err := b.tensorService.GetCollectionStats(b.TargetCollection)
	if err != nil {
		return err
	}

	shouldPlaceBid, err := shouldPlaceBid(collectionStats)
	if err != nil {
		return err
	}

	var userPool *UserPool

	// Get user pool, or create one if it doesn't exist.
	userPool, err = b.getUserPool()
	if err != nil && err.Error() == "not found" {
		// Don't create a pool and return if market conditions aren't ready for bids.
		if !shouldPlaceBid {
			return nil
		}

		userPool, err = b.createPool(bidPrice)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// If user already has a pool, and market conditions turned sour, close the pool.
	if !shouldPlaceBid {
		txSig, err := b.closePool(userPool.PoolAddress)
		if err != nil {
			return err
		}

		fmt.Printf("Closed pool %s: tx: %s\n", userPool.PoolAddress, txSig.String())

		return nil
	}

	fmt.Printf("Active bid for pool %s: %s\n", userPool.PoolAddress, userPool.BidPrice)

	// Check current bid price in pool. If it's different from the most recent calculated bid price, update it.
	if userPool.BidPrice != bidPrice {
		fmt.Printf("Setting new bid price for pool %s: %s\n", userPool.PoolAddress, bidPrice)
		// For now, we'll just close the order and the bot will create a new order on the next loop.
		// Down the road, we can optimize this to edit the order directly.

		txSig, err := b.closePool(userPool.PoolAddress)
		if err != nil {
			return err
		}

		fmt.Printf("Closed pool %s: tx: %s\n", userPool.PoolAddress, txSig.String())
	}

	// For now, let's assume our orders never get filled, and I'm not holding an NFT.

	return nil
}

func (b *Bot) listNFT(listPrice string) error {
	listNftResp, err := b.tensorService.ListNFT(b.botState.Mint, b.owner, listPrice)
	if err != nil {
		return err
	}

	bytes := dataToBytes(listNftResp.Data.TswapListNftTx.Txs[0].Tx.Data)

	sig, err := b.signAndSignTx(bytes)
	if err != nil {
		return err
	}

	fmt.Printf("Listed NFT %s: tx: %s\n", b.botState.Mint, sig.String())

	return nil
}

func (b *Bot) editSingleListing(listPrice string) error {
	editSingleListingResp, err := b.tensorService.EditSingleListing(b.botState.Mint, b.owner, listPrice)
	if err != nil {
		return err
	}

	spew.Dump(editSingleListingResp)

	bytes := dataToBytes(editSingleListingResp.Data.TswapEditSingleListingTx.Txs[0].Tx.Data)

	sig, err := b.signAndSignTx(bytes)
	if err != nil {
		return err
	}

	fmt.Printf("Updated NFT listing %s: tx: %s\n", b.botState.Mint, sig.String())

	return nil
}

func (b *Bot) executeListLogic() error {
	// TODO: implement
	fmt.Println("attempting to sell NFT")
	fmt.Println(b.botState)

	activeListings, err := b.tensorService.GetActiveListings(b.TargetCollection)
	if err != nil {
		return err
	}

	listPrice := activeListings.Data.ActiveListingsV2.Txs[0].Tx.GrossAmount

	// Only subtract one to lowest ask if it's not the bot's listing
	if activeListings.Data.ActiveListingsV2.Txs[0].Tx.SellerId != b.owner {
		listPrice, err = subtractOneLamport(listPrice)
		if err != nil {
			return err
		}
	}

	// get active listings for user to check if it's already listed.
	userActiveListings, err := b.tensorService.UserActiveListings(b.owner, b.TargetCollection)
	if err != nil {
		return err
	}

	if len(userActiveListings.Data.UserActiveListingsV2.Txs) == 0 {
		// List NFT
		err = b.listNFT(listPrice)
		if err != nil {
			return err
		}
		// err = b.listNFT(b.botState.GrossAmount)
		// if err != nil {
		// 	return err
		// }
	} else {
		// activeListing := userActiveListings.Data.UserActiveListingsV2.Txs[0].Tx

		// if listPrice != activeListing.GrossAmount {
		// 	// Update listing.
		// 	err = b.editSingleListing(listPrice)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}

	return nil
}

func (b *Bot) executeBotLogic() error {
	// Check to see if bot is holding an NFT
	if b.botState.Mint != "" {
		return b.executeListLogic()
	} else {
		return b.executeBidLogic()
	}
}

func (b *Bot) Run() error {
	err := b.executeBotLogic()
	if err != nil {
		return err
	}

	return nil
}
