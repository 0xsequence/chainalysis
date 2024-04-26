package chainalysis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/go-sequence/lib/prototyp"
)

// OracleAddress is the address of the Chainalysis Oracle
// https://go.chainalysis.com/chainalysis-oracle-docs.html
const (
	OracleAddress       = "0x40C57923924B5c5c5455c48D93317139ADDaC8fb"
	OracleStartingBlock = 14356508 - 10 // -10 so we don't miss anything
	OracleChainID       = 1
)

type Chainalysis interface {
	IsSanctioned(address string) (bool, error)
	Run(ctx context.Context) error
	Stop() error
	IsRunning() bool
}

// IndexSource is an interface that allows the chainalysis package to fetch pre-indexed events
type IndexSource interface {
	FetchSanctionedAddressEvents() ([]SanctionedAddressEvent, error)

	// SetIndex sets the index of the source to the given events.
	// this is no-op for the web source
	SetIndex([]SanctionedAddressEvent) error
}

type Options struct {
	Provider *ethrpc.Provider
	Source   IndexSource
}

type SanctionedAddressEvent struct {
	BlockNum  uint64          `json:"blockNum"`
	BlockHash string          `json:"blockHash"`
	Addrs     []prototyp.Hash `json:"addrs"`
}

type chainalysis struct {
	running int32
	ctx     context.Context
	stop    context.CancelFunc

	provider *ethrpc.Provider
	source   IndexSource

	mu                      sync.RWMutex
	sanctionedAddresses     map[string]struct{}
	sanctionedAddressEvents []SanctionedAddressEvent
}

func NewChainalysis(options *Options) (Chainalysis, error) {
	var provider *ethrpc.Provider
	var source IndexSource
	var err error

	if options == nil {
		options = &Options{}
	}

	if options.Provider != nil {
		provider = options.Provider
	} else {
		provider, err = ethrpc.NewProvider("https://nodes.sequence.app/mainnet")
		if err != nil {
			return nil, err
		}
	}

	if options.Source != nil {
		source = options.Source
	} else {
		source = embeddedSource // default we use our embedded source
	}

	lc := &chainalysis{
		sanctionedAddresses:     make(map[string]struct{}),
		sanctionedAddressEvents: []SanctionedAddressEvent{},
		source:                  source,
		provider:                provider,
	}

	return lc, nil
}

func (l *chainalysis) Run(ctx context.Context) error {
	if l.IsRunning() {
		return fmt.Errorf("chainalysis: already running")
	}

	// inital sync
	var err error
	l.sanctionedAddressEvents, err = l.source.FetchSanctionedAddressEvents()
	if err != nil {
		return err
	}

	for _, event := range l.sanctionedAddressEvents {
		for _, addr := range event.Addrs {
			l.sanctionedAddresses[addr.String()] = struct{}{}
		}
	}

	atomic.StoreInt32(&l.running, 1)

	l.ctx, l.stop = context.WithCancel(ctx)

	return l.fetcher(ctx)
}

func (l *chainalysis) Stop() error {
	if !l.IsRunning() {
		return fmt.Errorf("chainalysis: not running")
	}

	atomic.StoreInt32(&l.running, 0)
	l.stop()

	// Update the IndexSource with the latest events fetched from the provider
	// this only works if its a file source, otherwise it's a no-op
	l.mu.Lock()
	l.source.SetIndex(l.sanctionedAddressEvents)
	l.mu.Unlock()

	return nil
}

func (l *chainalysis) IsRunning() bool {
	return atomic.LoadInt32(&l.running) == 1
}

func (l *chainalysis) IsSanctioned(address string) (bool, error) {
	formattedAddress := prototyp.HashFromString(address).String()
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.sanctionedAddresses[formattedAddress]
	return ok, nil
}

func (l *chainalysis) fetcher(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			l.mu.RLock()
			lastFetchedBlockNum := l.sanctionedAddressEvents[len(l.sanctionedAddressEvents)-1].BlockNum
			l.mu.RUnlock()

			latestBlock, err := l.provider.BlockNumber(ctx)
			if err != nil {
				continue
			}

			if latestBlock <= lastFetchedBlockNum {
				continue
			}

			sanctionedAddressesFromSource, err := fetchSanctionedAddressEvents(ctx, l.provider, lastFetchedBlockNum+1, latestBlock)
			if err != nil {
				continue
			}

			l.mu.Lock()
			l.sanctionedAddressEvents = append(l.sanctionedAddressEvents, sanctionedAddressesFromSource...)
			for _, event := range sanctionedAddressesFromSource {
				for _, addr := range event.Addrs {
					l.sanctionedAddresses[addr.String()] = struct{}{}
				}
			}
			l.mu.Unlock()
		}
	}
}
