package chainalysis

import (
	"context"
	"errors"
	"time"

	"github.com/0xsequence/ethkit/ethrpc"
)

type onChainalysis struct {
	provider *ethrpc.Provider
}

func NewOnChainalysis() (Chainalysis, error) {
	provider, err := ethrpc.NewProvider("https://nodes.sequence.app/mainnet")
	if err != nil {
		return nil, err
	}
	return &onChainalysis{
		provider: provider,
	}, nil
}

func (o *onChainalysis) IsSanctioned(address string) (bool, error) {
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
