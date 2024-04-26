chainalysis
===========

Track sanctioned addresses from Chainalysis Oracle.

```go
OracleAddress       = "0x40C57923924B5c5c5455c48D93317139ADDaC8fb"
OracleStartingBlock = 14356508
OracleChainID       = 1
```

## Usage

Please see _example.

Note: there is an embedded index in the library, and if you call .Run()
it will listen for updates every 10 minutes.

Calling `make update-index` will update the packages index/sanctioned_addresses.json file

## LICENSE

MIT or Apache 2.0
