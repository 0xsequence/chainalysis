package main

import (
	"log"
	"os"

	"github.com/0xsequence/chainalysis"
)

func main() {
	file, err := os.OpenFile("sanctioned_addresses.json", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	chainalysis.FetchAndUpdateSanctionedAddresses(file, 16734673, 17118381)

}
