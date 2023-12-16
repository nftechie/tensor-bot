package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"tensor/pkg/tensor"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var (
	solanaClient *rpc.Client
	wsClient     *ws.Client
	privateKey   solana.PrivateKey
	pool         string
)

func dataToBytes(data []int) []byte {
	bytes := make([]byte, len(data))
	for i, v := range data {
		bytes[i] = byte(v)
	}

	return bytes
}

func signAndSignTx(bytes []byte) (*solana.Signature, error) {
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(bytes))
	if err != nil {
		return nil, err
	}

	tx.Signatures = []solana.Signature{}

	out, err := solanaClient.GetLatestBlockhash(
		context.Background(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return nil, err
	}

	tx.Message.RecentBlockhash = out.Value.Blockhash

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if privateKey.PublicKey().Equals(key) {
				return &privateKey
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
		solanaClient,
		wsClient,
		tx,
	)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func init() {
	flag.StringVar(&pool, "pool", "", "pool address")
}

func main() {
	flag.Parse()
	if pool == "" {
		log.Fatalln("pool not defined")
	}

	var err error

	tensorService := tensor.NewTensorService()

	solanaClient = rpc.New(rpc.MainNetBeta_RPC)
	wsClient, err = ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	privateKey, err = solana.PrivateKeyFromSolanaKeygenFile("/Users/alexwormuth/.config/solana/id.json")
	if err != nil {
		panic(err)
	}

	closePoolResp, err := tensorService.ClosePool(pool)
	if err != nil {
		panic(err)
	}

	bytes := dataToBytes(closePoolResp.Data.TswapClosePoolTx.Txs[0].Tx.Data)

	sig, err := signAndSignTx(bytes)
	if err != nil {
		panic(err)
	}

	fmt.Println("Close pool: tx: ", sig.String())
}
