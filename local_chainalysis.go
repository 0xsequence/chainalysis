package chainalysis

import (
	"context"
	"sync"
	"time"

	"github.com/0xsequence/go-sequence/lib/prototyp"
	"github.com/rs/zerolog/log"
)

const sanctionedAddressesSource = "https://raw.githubusercontent.com/0xsequence/chainalysis/initial-version/index/sanctioned_addresses.json"

type sanctionedAddressEvent struct {
	BlockNum  uint64          `json:"blockNum"`
	BlockHash string          `json:"blockHash"`
	Addrs     []prototyp.Hash `json:"addrs"`
}

type localChainalysis struct {
	mu                  sync.RWMutex
	SanctionedAddresses map[string]struct{}
}

func NewLocalChainalysis(ctx context.Context) (Chainalysis, error) {
	sanctionedAddresses := map[string]struct{}{}

	sanctionedAddressEvents, err := fetchSanctionedAddressEventsFromSource(sanctionedAddressesSource)
	if err != nil {
		return nil, err
	}

	for _, event := range sanctionedAddressEvents {
		for _, addr := range event.Addrs {
			sanctionedAddresses[addr.String()] = struct{}{}
		}
	}

	lc := &localChainalysis{
		SanctionedAddresses: sanctionedAddresses,
	}

	go lc.fetcher(ctx)

	return lc, nil
}

func (l *localChainalysis) IsSanctioned(address string) (bool, error) {
	formattedAddress := prototyp.HashFromString(address).String()
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.SanctionedAddresses[formattedAddress]
	return ok, nil
}

func (l *localChainalysis) fetcher(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sanctionedAddressesFromSource, err := fetchSanctionedAddressEventsFromSource(sanctionedAddressesSource)
			if err != nil {
				log.Warn().Err(err).Msg("failed to fetch and update sanctioned addresses")
			}
			l.mu.Lock()
			for _, event := range sanctionedAddressesFromSource {
				for _, addr := range event.Addrs {
					l.SanctionedAddresses[addr.String()] = struct{}{}
				}
			}
			l.mu.Unlock()
		}
	}
}
