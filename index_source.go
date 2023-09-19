package chainalysis

type IndexSource interface {
	FetchSanctionedAddressEvents() ([]sanctionedAddressEvent, error)

	// SetIndex sets the index of the source to the given events.
	// this is no-op for the web source
	SetIndex([]sanctionedAddressEvent) error
}
