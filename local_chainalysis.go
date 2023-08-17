package chainalysis

import (
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/go-sequence/lib/prototyp"
)

const sanctionedAddressesSource = "https://raw.githubusercontent.com/0xsequence/chainalysis/initial-version/index/sanctioned_addresses.json"

type sanctionedAddressEvent struct {
	BlockNum  uint64          `json:"blockNum"`
	BlockHash string          `json:"blockHash"`
	Addrs     []prototyp.Hash `json:"addrs"`
}

type localChainalysis struct {
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

	return &localChainalysis{
		SanctionedAddresses: sanctionedAddresses,
	}, nil
}

func (l *localChainalysis) IsSanctioned(address string) (bool, error) {
	formattedAddress := prototyp.HashFromString(address).String()
	_, ok := l.SanctionedAddresses[formattedAddress]
	return ok, nil
}

func FetchAndUpdateSanctionedAddresses(file *os.File, startingBlock uint64, endingBlock uint64) error {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	preFetchedEvents := []sanctionedAddressEvent{}

	if len(fileData) > 0 {
		err = json.Unmarshal(fileData, &preFetchedEvents)

		if err != nil {
			return err
		}
	}

	if len(preFetchedEvents) > 0 {
		lastFetchedEvent := preFetchedEvents[len(preFetchedEvents)-1]
		if lastFetchedEvent.BlockNum >= startingBlock {
			startingBlock = lastFetchedEvent.BlockNum + 1
		} else if lastFetchedEvent.BlockNum >= endingBlock {
			return nil
		}
	}

	newEvents, err := fetchSanctionedAddressEvents(startingBlock, endingBlock)
	if err != nil {
		return err
	}

	newEvents = append(preFetchedEvents, newEvents...)

	data, err := json.Marshal(newEvents)
	if err != nil {
		return err
	}
	_, err = file.WriteAt(data, 0)
	if err != nil {
		return err
	}
	return nil
}

func fetchSanctionedAddressEvents(startingBlock uint64, endingBlock uint64) ([]sanctionedAddressEvent, error) {
	provider, err := ethrpc.NewProvider("https://nodes.sequence.app/mainnet")
	if err != nil {
		return nil, err
	}

	result := []sanctionedAddressEvent{}

	contract := common.HexToAddress(OracleAddress)
	logs, _, err := fetchEthereumLogs(context.Background(), provider, 10000, 8000, startingBlock, endingBlock, &contract, "0x2596d7dd6966c5673f9c06ddb0564c4f0e6d8d206ea075b83ad9ddd71a4fb927")

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

func fetchSanctionedAddressEventsFromSource(source string) ([]sanctionedAddressEvent, error) {
	res, err := http.DefaultClient.Get(source)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	events := []sanctionedAddressEvent{}
	err = json.Unmarshal(buf, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}
