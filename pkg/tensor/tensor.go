package tensor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type TensorService struct {
	httpClient   *http.Client
	apiKey       string
	tensorApiUrl string
}

func NewTensorService() *TensorService {
	return &TensorService{
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		apiKey:       "e3a7eaf1-abc4-4fb7-addd-5755a6bb197d",
		tensorApiUrl: "https://api.tensor.so/graphql",
	}
}

func (s *TensorService) GetCollectionStats(collectionSlug string) (*CollectionStatsResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "CollectionStats",
		"query": "query CollectionStats($slug: String!) { instrumentTV2(slug: $slug) { statsV2 { currency buyNowPrice buyNowPriceNetFees sellNowPrice sellNowPriceNetFees numListed numMints floor7d sales1h sales24h sales7d salesAll volume1h volume24h volume7d volumeAll } } }",
		"variables": {
			"slug": "` + collectionSlug + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse CollectionStatsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) GetActiveListings(collectionSlug string) (*ActiveListingsResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "ActiveListingsV2",
		"query": "query ActiveListingsV2(  $slug: String!  $sortBy: ActiveListingsSortBy!  $filters: ActiveListingsFilters  $limit: Int  $cursor: ActiveListingsCursorInputV2) {  activeListingsV2(    slug: $slug    sortBy: $sortBy    filters: $filters    limit: $limit    cursor: $cursor  ) {    page {      endCursor {        str      }      hasMore    }    txs {      mint {        onchainId      }      tx {        sellerId        grossAmount        grossAmountUnit      }    }  }}",
		"variables": {
			"slug": "` + collectionSlug + `",
			"sortBy": "PriceAsc",
			"filters": {
			  "sources": ["TENSORSWAP", "TCOMP"]
			},
			"limit": 10,
			"cursor": null 
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ActiveListingsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) GetActiveOrders(collectionSlug string) (*ActiveOrdersResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TensorSwapActiveOrders",
		"query": "query TensorSwapActiveOrders($slug: String!) {  tswapOrders(slug: $slug) {    address    createdUnix    curveType    delta    nftsHeld    ownerAddress    poolType    solBalance    startingPrice    buyNowPrice    sellNowPrice    statsAccumulatedMmProfit    statsTakerBuyCount    statsTakerSellCount    takerBuyCount    takerSellCount    updatedAt  }}",
		"variables": {
			"slug": "` + collectionSlug + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ActiveOrdersResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) CreatePool(collectionSlug string, owner string, lamports string) (*CreatePoolResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TswapInitPoolTx",
		"query": "query TswapInitPoolTx($config: PoolConfig!, $owner: String!, $slug: String!) {  tswapInitPoolTx(config: $config, owner: $owner, slug: $slug) {    pool    txs {      lastValidBlockHeight      tx    }  }}",
		"variables": {  
			"config": {
			  "poolType": "TOKEN",
			  "curveType": "LINEAR",
			  "delta": "0",
			  "startingPrice": "` + lamports + `",
			  "mmCompoundFees": false
			},
			"owner": "` + owner + `",
			"slug": "` + collectionSlug + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse CreatePoolResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) ClosePool(pool string) (*ClosePoolResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TswapClosePoolTx",
		"query": "query TswapClosePoolTx($pool: String!) {  tswapClosePoolTx(pool: $pool) {    txs {      lastValidBlockHeight      tx    }  }}",
		"variables": {  
			"pool": "` + pool + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ClosePoolResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) PoolDepositSol(pool string, lamports string) (*PoolDepositSolResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TswapDepositWithdrawSolTx",
		"query": "query TswapDepositWithdrawSolTx($action: DepositWithdrawAction!, $lamports: Decimal!, $pool: String!) {  tswapDepositWithdrawSolTx(action: $action, lamports: $lamports, pool: $pool) {    txs {      lastValidBlockHeight      tx    }  }}",
		"variables": {  
			"action": "DEPOSIT",
			"lamports": "` + lamports + `",
			"pool": "` + pool + `"
		  }
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse PoolDepositSolResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) UserPools(owner string) (*UserPoolsResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "UserTensorSwapOrders",
		"query": "query UserTensorSwapOrders($owner: String!) {  userTswapOrders(owner: $owner) {    slug pool {      address      createdUnix      curveType      delta      mmFeeBps      nftsForSale {        onchainId      }      nftsHeld      ownerAddress      poolType      solBalance      startingPrice      buyNowPrice      sellNowPrice      statsAccumulatedMmProfit      statsTakerBuyCount      statsTakerSellCount      takerBuyCount      takerSellCount      updatedAt    }  }}",
		"variables": {
			"owner": "` + owner + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse UserPoolsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) ListNFT(mint string, owner string, price string) (*ListNFTResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TswapListNftTx",
		"query": "query TswapListNftTx($mint: String!, $owner: String!, $price: Decimal!) {  tswapListNftTx(mint: $mint, owner: $owner, price: $price) {    txs {      lastValidBlockHeight      tx      txV0    }  }}",
		"variables": {
			"mint": "` + mint + `",
			"owner": "` + owner + `",
			"price": "` + price + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ListNFTResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) EditSingleListing(mint string, owner string, price string) (*EditSingleListingResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "TswapEditSingleListingTx",
		"query": "query TswapEditSingleListingTx($mint: String!, $owner: String!, $price: Decimal!) {  tswapEditSingleListingTx(mint: $mint, owner: $owner, price: $price) {    txs {      lastValidBlockHeight      tx      txV0    }  }}",
		"variables": {
			"mint": "` + mint + `",
			"owner": "` + owner + `",
			"price": "` + price + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse EditSingleListingResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *TensorService) UserActiveListings(owner string, slug string) (*UserActiveListingsResponse, error) {
	requestBody := []byte(`{ 
		"operationName": "UserActiveListingsV2",
		"query": "query UserActiveListingsV2(  $wallets: [String!]!  $sortBy: ActiveListingsSortBy!  $cursor: ActiveListingsCursorInputV2  $limit: Int  $slug: String) {  userActiveListingsV2(    wallets: $wallets    cursor: $cursor    limit: $limit    sortBy: $sortBy    slug: $slug  ) {    page {      endCursor {        str      }      hasMore    }    txs {      tx {        txId        txAt        source        mintOnchainId        grossAmount      }    }  }}",
		"variables": {
			"wallets": ["` + owner + `"],
			"sortBy": "PriceAsc",
			"cursor": null,
			"limit": 1,
			"slug": "` + slug + `"
		}
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", s.tensorApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tensor-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse UserActiveListingsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}
