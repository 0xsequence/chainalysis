package chainalysis

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed index/sanctioned_addresses.json
var indexSanctionedAddresses []byte

var embeddedSource IndexSource

func init() {
	contents := strings.TrimSpace(string(indexSanctionedAddresses))
	if len(contents) == 0 {
		return
	}
	var err error
	embeddedSource, err = NewEmbedSource([]byte(contents))
	if err != nil {
		panic(err)
	}
}

type embedSource struct {
	events []SanctionedAddressEvent
}

func NewEmbedSource(data []byte) (IndexSource, error) {
	events := []SanctionedAddressEvent{}
	err := json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}

	return &embedSource{
		events: events,
	}, nil
}

func (f *embedSource) FetchSanctionedAddressEvents() ([]SanctionedAddressEvent, error) {
	return f.events, nil
}

func (f *embedSource) SetIndex(events []SanctionedAddressEvent) error {
	f.events = events
	return nil
}
