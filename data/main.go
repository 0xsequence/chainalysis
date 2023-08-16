package main

import (
	"os"

	"github.com/0xSequence/chainalysis"
)

func main() {
	file, err := os.OpenFile("sanctioned_addresses.json", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	chainalysis.FetchAndUpdateSanctionedAddresses(file, 16734673, 17118381)

}
