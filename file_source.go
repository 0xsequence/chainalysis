package chainalysis

import (
	"encoding/json"
	"io"
	"os"
)

type fileSource struct {
	source *os.File
	events []SanctionedAddressEvent
}

// NewFileSource creates a new file source, which implements the IndexSource interface
// we assume this file has a similar structure to our pre-indexed json file in index/sanctioned_addresses.json
// this can be used so that we don't have to re-index the logs from the point we left off
func NewFileSource(sourceFile *os.File) IndexSource {
	return &fileSource{
		source: sourceFile,
		events: []SanctionedAddressEvent{},
	}
}

func (f *fileSource) FetchSanctionedAddressEvents() ([]SanctionedAddressEvent, error) {
	if f.events != nil {
		return f.events, nil
	}

	fileData, err := io.ReadAll(f.source)
	if err != nil {
		return nil, err
	}

	preFetchedEvents := []SanctionedAddressEvent{}

	if len(fileData) > 0 {
		err = json.Unmarshal(fileData, &preFetchedEvents)

		if err != nil {
			return nil, err
		}
	}

	f.events = preFetchedEvents

	return preFetchedEvents, nil
}

// SetIndex writes the events to the file source
func (f *fileSource) SetIndex(events []SanctionedAddressEvent) error {
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}

	_, err = f.source.WriteAt(data, 0)
	f.events = events
	return err
}
