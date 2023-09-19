package chainalysis

import (
	"encoding/json"
	"io"
	"os"
)

type fileSource struct {
	source *os.File
}

// NewFileSource creates a new file source, which implements the IndexSource interface
// we assume this file has a similar structure to our pre-indexed json file in index/sanctioned_addresses.json
// this can be used so that we don't have to re-index the logs from the point we left off
func NewFileSource(sourceFile *os.File) IndexSource {
	return &fileSource{
		source: sourceFile,
	}
}

func (f *fileSource) FetchSanctionedAddressEvents() ([]sanctionedAddressEvent, error) {
	fileData, err := io.ReadAll(f.source)
	if err != nil {
		return nil, err
	}

	preFetchedEvents := []sanctionedAddressEvent{}

	if len(fileData) > 0 {
		err = json.Unmarshal(fileData, &preFetchedEvents)

		if err != nil {
			return nil, err
		}
	}

	return preFetchedEvents, nil
}

// SetIndex is no-op for the file source
func (f *fileSource) SetIndex(events []sanctionedAddressEvent) error {
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}

	_, err = f.source.WriteAt(data, 0)
	return err
}
