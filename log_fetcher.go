package chainalysis

import (
	"context"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/go-sequence/lib/prototyp"
)

func FetchAndUpdateSanctionedAddresses(ctx context.Context, provider *ethrpc.Provider, source IndexSource, startingBlock uint64, endingBlock uint64) error {
	preFetchedEvents, err := source.FetchSanctionedAddressEvents()
	if err != nil {
		return err
	}

	if len(preFetchedEvents) > 0 {
		lastFetchedEvent := preFetchedEvents[len(preFetchedEvents)-1]
		if lastFetchedEvent.BlockNum >= startingBlock {
			startingBlock = lastFetchedEvent.BlockNum + 1
		} else if lastFetchedEvent.BlockNum >= endingBlock {
			return nil
		}
	}

	newEvents, err := fetchSanctionedAddressEvents(ctx, provider, startingBlock, endingBlock)
	if err != nil {
		return err
	}

	newEvents = append(preFetchedEvents, newEvents...)

	err = source.SetIndex(newEvents)
	return err
}

func fetchSanctionedAddressEvents(ctx context.Context, provider *ethrpc.Provider, startingBlock uint64, endingBlock uint64) ([]sanctionedAddressEvent, error) {
	result := []sanctionedAddressEvent{}

	contract := common.HexToAddress(OracleAddress)
	logs, _, err := fetchEthereumLogs(ctx, provider, 10000, 8000, startingBlock, endingBlock, &contract, "0x2596d7dd6966c5673f9c06ddb0564c4f0e6d8d206ea075b83ad9ddd71a4fb927")
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		logData, err := ethcoder.AbiDecoderWithReturnedValues([]string{"address[]"}, log.Data)
		if err != nil {
			return nil, err
		}
		addrs, ok := logData[0].([]common.Address)
		if !ok {
			return nil, err
		}
		event := sanctionedAddressEvent{
			BlockNum:  log.BlockNumber,
			BlockHash: log.BlockHash.Hex(),
			Addrs:     []prototyp.Hash{},
		}

		for _, address := range addrs {
			event.Addrs = append(event.Addrs, prototyp.HashFromString(address.Hex()))
		}
		result = append(result, event)
	}

	return result, nil
}
