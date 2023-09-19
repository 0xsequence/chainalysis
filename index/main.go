package main

import (
	"context"
	"log"
	"os"

	"github.com/0xsequence/chainalysis"
	"github.com/0xsequence/ethkit/ethrpc"
)

// this script is to update the sanctioned_addresses.json file
func main() {
	file, err := os.OpenFile("sanctioned_addresses.json", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	source := chainalysis.NewFileSource(file)
	provider, err := ethrpc.NewProvider("https://nodes.sequence.app/mainnet")
	if err != nil {
		log.Fatal(err)
	}

	latestBlock, err := provider.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	chainalysis.FetchAndUpdateSanctionedAddresses(context.Background(), provider, source, 16734673, latestBlock)
}
