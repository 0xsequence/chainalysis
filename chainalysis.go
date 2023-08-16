package chainalysis

const OracleAddress = "0x40C57923924B5c5c5455c48D93317139ADDaC8fb"

type ChainAlysis interface {
	IsSanctioned(address string) (bool, error)
}
