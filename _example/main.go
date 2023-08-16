package main

import (
	"fmt"

	"github.com/0xsequence/chainalysis"
)

func main() {
	localChainAlysis, err := chainalysis.NewLocalChainalysis()
	if err != nil {
		panic(err)
	}
	onChainAlysis, err := chainalysis.NewOnChainalysis()
	if err != nil {
		panic(err)
	}

	fmt.Println(localChainAlysis.IsSanctioned("0x01e2919679362dFBC9ee1644Ba9C6da6D6245BB1"))
	fmt.Println(onChainAlysis.IsSanctioned("0x01e2919679362dFBC9ee1644Ba9C6da6D6245BB1"))
}
