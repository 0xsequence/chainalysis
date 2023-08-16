package chainalysis

import (
	"context"
	"errors"
	"time"

	"github.com/0xsequence/ethkit/ethrpc"
)

type onChainAlysis struct {
	provider *ethrpc.Provider
}

func NewOnChainAlysis() (ChainAlysis, error) {
	provider, err := ethrpc.NewProvider("https://nodes.sequence.app/mainnet")
	if err != nil {
		return nil, err
	}
	return &onChainAlysis{
		provider: provider,
	}, nil
}

func (o *onChainAlysis) IsSanctioned(address string) (bool, error) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	result, err := o.provider.ContractQuery(ctx, OracleAddress, "isSanctioned(address addr)", "bool", []string{address})
	if err != nil {
		return false, err
	}

	return decodeOutputArray(result)
}

func decodeOutputArray(output []string) (bool, error) {
	if len(output) != 1 {
		return false, errors.New("invalid output length")
	}
	return output[0] == "true", nil
}
