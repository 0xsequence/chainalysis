package chainalysis

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/goware/breaker"
)

func fetchEthereumLogs(ctx context.Context, provider ethrpc.Interface, maxBatchSize, lastBatchSize, from, to uint64, optContractFilter *common.Address, topicID string) ([]types.Log, uint64, error) {
	result := []types.Log{}

	batchSize := lastBatchSize
	additiveFactor := uint64(float64(batchSize) * 0.10)

	br := breaker.New(slog.Default(), time.Second*2, 2, 5)

	for i := from; i < to; {
		dst := min(i+batchSize, to)

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(i)),
			ToBlock:   big.NewInt(int64(dst - 1)),
			Topics:    [][]common.Hash{{common.HexToHash(topicID)}},
		}

		// optional contract filter, useful for debugging
		if optContractFilter != nil {
			query.Addresses = []common.Address{*optContractFilter}
		}

		var logs []types.Log

		err := br.Do(ctx, func() error {
			var err error
			logs, err = provider.FilterLogs(ctx, query)
			if err != nil {
				if tooMuchDataRequestedError(err) {
					slog.Warn("fetchEthereumLogs hit too-much-data error", "batchSize", batchSize)
					batchSize = uint64(float64(batchSize) / 1.5)
					return err
				}

				if !errors.Is(err, context.Canceled) {
					slog.Error("fetchEthereumLogs failed", "error", err)
				}

				slog.Warn("fetchEthereumLogs error", "error", err)
				return err
			}
			return nil
		})

		if err != nil {
			return nil, batchSize, err
		}

		// append logs to result
		result = append(result, logs...)

		// check if the execution is over after each query batch
		if err := ctx.Err(); err != nil {
			return nil, batchSize, err
		}

		i = dst

		// update the batchSize with additive increase
		if i < to && batchSize < maxBatchSize {
			batchSize = min(maxBatchSize, batchSize+additiveFactor)
		}
	}

	slog.Debug("fetchEthereumLogs completed", "fromBlock", from, "toBlock", to, "logCount", len(result))

	return result, batchSize, nil
}

func tooMuchDataRequestedError(err error) bool {
	return strings.Contains(err.Error(), "query returned more than")
}
