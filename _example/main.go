package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xsequence/chainalysis"
)

func main() {
	localChainAlysis, err := chainalysis.NewLocalChainalysis(nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		localChainAlysis.Run(context.Background())
	}()

	// wait for the chainalysis to start
	var counter int
	for !localChainAlysis.IsRunning() {
		time.Sleep(1 * time.Second)
		counter++
		if counter > 5 {
			log.Fatal("chainalysis is not running")
		}
	}

	fmt.Println(localChainAlysis.IsSanctioned("0x01e2919679362dFBC9ee1644Ba9C6da6D6245BB1"))

	localChainAlysis.Stop()
}
