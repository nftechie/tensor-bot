package tensor

type CollectionStatsResponse struct {
	Data struct {
		InstrumentTV2 struct {
			StatsV2 struct {
				Currency            *string `json:"currency"`
				BuyNowPrice         string  `json:"buyNowPrice"`
				BuyNowPriceNetFees  string  `json:"buyNowPriceNetFees"`
				SellNowPrice        string  `json:"sellNowPrice"`
				SellNowPriceNetFees string  `json:"sellNowPriceNetFees"`
				NumListed           int     `json:"numListed"`
				NumMints            int     `json:"numMints"`
				Sales1h             int     `json:"sales1h"`
				Sales24h            int     `json:"sales24h"`
				Sales7d             int     `json:"sales7d"`
				SalesAll            int     `json:"salesAll"`
				Volume1h            string  `json:"volume1h"`
				Volume24h           string  `json:"volume24h"`
				Volume7d            string  `json:"volume7d"`
				VolumeAll           string  `json:"volumeAll"`
			} `json:"statsV2"`
		} `json:"instrumentTV2"`
	} `json:"data"`
}

type ActiveListingsResponse struct {
	Data struct {
		ActiveListingsV2 struct {
			Page struct {
				EndCursor struct {
					Str string `json:"str"`
				} `json:"endCursor"`
				HasMore bool `json:"hasMore"`
			} `json:"page"`
			Txs []struct {
				Mint struct {
					OnchainId string `json:"onchainId"`
				} `json:"mint"`
				Tx struct {
					SellerId        string `json:"sellerId"`
					GrossAmount     string `json:"grossAmount"`
					GrossAmountUnit string `json:"grossAmountUnit"`
				} `json:"tx"`
			} `json:"txs"`
		} `json:"activeListingsV2"`
	} `json:"data"`
}

type TswapOrder struct {
	Address                  string  `json:"address"`
	CreatedUnix              int64   `json:"createdUnix"`
	CurveType                string  `json:"curveType"`
	Delta                    string  `json:"delta"`
	NftsHeld                 int     `json:"nftsHeld"`
	OwnerAddress             string  `json:"ownerAddress"`
	PoolType                 string  `json:"poolType"`
	SolBalance               string  `json:"solBalance"`
	StartingPrice            string  `json:"startingPrice"`
	BuyNowPrice              *string `json:"buyNowPrice"`
	SellNowPrice             *string `json:"sellNowPrice"`
	StatsAccumulatedMmProfit string  `json:"statsAccumulatedMmProfit"`
	StatsTakerBuyCount       int     `json:"statsTakerBuyCount"`
	StatsTakerSellCount      int     `json:"statsTakerSellCount"`
	TakerBuyCount            int     `json:"takerBuyCount"`
	TakerSellCount           int     `json:"takerSellCount"`
	UpdatedAt                int64   `json:"updatedAt"`
}

type ActiveOrdersResponse struct {
	Data struct {
		TswapOrders []TswapOrder `json:"tswapOrders"`
	} `json:"data"`
}

type Transaction struct {
	LastValidBlockHeight int64 `json:"lastValidBlockHeight"`
	Tx                   struct {
		Type string `json:"type"`
		Data []int  `json:"data"`
	}
}

type CreatePoolResponse struct {
	Data struct {
		TswapInitPoolTx struct {
			Pool string        `json:"pool"`
			Txs  []Transaction `json:"txs"`
		} `json:"tswapInitPoolTx"`
	} `json:"data"`
}

type ClosePoolResponse struct {
	Data struct {
		TswapClosePoolTx struct {
			Pool string        `json:"pool"`
			Txs  []Transaction `json:"txs"`
		} `json:"tswapClosePoolTx"`
	} `json:"data"`
}

type PoolDepositSolResponse struct {
	Data struct {
		TswapDepositWithdrawSolTx struct {
			Pool string        `json:"pool"`
			Txs  []Transaction `json:"txs"`
		} `json:"tswapDepositWithdrawSolTx"`
	} `json:"data"`
}

type UserPoolsResponse struct {
	Data struct {
		UserTswapOrders []struct {
			Slug string `json:"slug"`
			Pool struct {
				Address                  string   `json:"address"`
				CreatedUnix              int64    `json:"createdUnix"`
				CurveType                string   `json:"curveType"`
				Delta                    string   `json:"delta"`
				MmFeeBps                 *int     `json:"mmFeeBps"` // Pointer to handle null
				NftsForSale              []string `json:"nftsForSale"`
				NftsHeld                 int      `json:"nftsHeld"`
				OwnerAddress             string   `json:"ownerAddress"`
				PoolType                 string   `json:"poolType"`
				SolBalance               string   `json:"solBalance"`
				StartingPrice            string   `json:"startingPrice"`
				BuyNowPrice              *string  `json:"buyNowPrice"` // Pointer to handle null
				SellNowPrice             string   `json:"sellNowPrice"`
				StatsAccumulatedMmProfit string   `json:"statsAccumulatedMmProfit"`
				StatsTakerBuyCount       int      `json:"statsTakerBuyCount"`
				StatsTakerSellCount      int      `json:"statsTakerSellCount"`
				TakerBuyCount            int      `json:"takerBuyCount"`
				TakerSellCount           int      `json:"takerSellCount"`
				UpdatedAt                int64    `json:"updatedAt"`
			} `json:"pool"`
		} `json:"userTswapOrders"`
	} `json:"data"`
}

type ListNFTResponse struct {
	Data struct {
		TswapListNftTx struct {
			Txs []Transaction `json:"txs"`
		} `json:"tswapListNftTx"`
	} `json:"data"`
}

type EditSingleListingResponse struct {
	Data struct {
		TswapEditSingleListingTx struct {
			Txs []Transaction `json:"txs"`
		} `json:"tswapEditSingleListingTx"`
	} `json:"data"`
}

type UserActiveListingsResponse struct {
	Data struct {
		UserActiveListingsV2 struct {
			Page struct {
				EndCursor struct {
					Str string `json:"str"`
				} `json:"endCursor"`
				HasMore bool `json:"hasMore"`
			} `json:"page"`
			Txs []struct {
				Tx struct {
					TxId          string `json:"txId"`
					Source        string `json:"source"`
					MintOnchainId string `json:"mintOnchainId"`
					GrossAmount   string `json:"grossAmount"`
				} `json:"tx"`
			} `json:"txs"`
		} `json:"userActiveListingsV2"`
	} `json:"data"`
}
