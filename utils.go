package chainalysis

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
)

func fetchEthereumLogs(ctx context.Context, provider ethrpc.Interface, maxBatchSize, lastBatchSize, from, to uint64, optContractFilter *common.Address, topicID string) ([]types.Log, uint64, error) {
	result := []types.Log{}

	batchSize := lastBatchSize
	additiveFactor := uint64(float64(batchSize) * 0.10)

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

		logs, err := provider.FilterLogs(ctx, query)
		if err != nil {
			if tooMuchDataRequestedError(err) {
				log.Warn().Msgf("fetchEthereumLogs hit too-much-data error for batchSize %d", batchSize)
				batchSize = uint64(float64(batchSize) / 1.5)
				continue
			}
			if !errors.Is(err, context.Canceled) {
				log.Err(err).Msgf("fetchEthereumLogs failed")
			}
			log.Warn().Msgf("fetchEthereumLogs error '%v'", err)
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

	log.Debug().Msgf("fetchEthereumLogs from block %d to %d retrieved %d logs", from, to, len(result))

	return result, batchSize, nil
}

func tooMuchDataRequestedError(err error) bool {
	if strings.Contains(err.Error(), "query returned more than") { // 10000 results") {
		return true
	}
	return false
}
