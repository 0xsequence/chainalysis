package chainalysis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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
	running int32
	ctx     context.Context
	stop    context.CancelFunc

	mu                  sync.RWMutex
	SanctionedAddresses map[string]struct{}
}

func NewLocalChainalysis() (Chainalysis, error) {
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

	return lc, nil
}

func (l *localChainalysis) Run(ctx context.Context) error {
	if l.IsRunning() {
		return fmt.Errorf("chainalysis: already running")
	}

	atomic.StoreInt32(&l.running, 1)

	l.ctx, l.stop = context.WithCancel(ctx)

	return l.fetcher(ctx)
}

func (l *localChainalysis) Stop() error {
	if !l.IsRunning() {
		return fmt.Errorf("chainalysis: not running")
	}

	atomic.StoreInt32(&l.running, 0)

	l.stop()
	return nil
}

func (l *localChainalysis) IsRunning() bool {
	return atomic.LoadInt32(&l.running) == 1
}

func (l *localChainalysis) IsSanctioned(address string) (bool, error) {
	formattedAddress := prototyp.HashFromString(address).String()
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.SanctionedAddresses[formattedAddress]
	return ok, nil
}

func (l *localChainalysis) fetcher(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
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
