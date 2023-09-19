package chainalysis

import (
	"encoding/json"
	"io"
	"net/http"
)

const DefaultWebSourceURL = "https://raw.githubusercontent.com/0xsequence/chainalysis/initial-version/index/sanctioned_addresses.json"

type webSource struct {
	source string
}

// NewWebSource creates a new web source, this is the default source for the chainalysis package
// it uses the index we have stored in index/sanctioned_addresses.json
func NewWebSource(opSourceURL ...string) IndexSource {
	sourceURL := DefaultWebSourceURL
	if len(opSourceURL) > 0 {
		sourceURL = opSourceURL[0]
	}

	return &webSource{
		source: sourceURL,
	}
}

func (w *webSource) FetchSanctionedAddressEvents() ([]sanctionedAddressEvent, error) {
	res, err := http.DefaultClient.Get(w.source)
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

// SetIndex is no-op for the web source
func (w *webSource) SetIndex([]sanctionedAddressEvent) error {
	return nil
}
