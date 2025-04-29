package types

type AttestRequired struct {
	BlockHash BlockHash
}

type AttestInfo struct {
	TargetBlock     BlockNumber
	TargetBlockHash BlockHash
	WindowStart     BlockNumber
	WindowEnd       BlockNumber
}

type AttestFee struct {
	defined bool
	value   uint64
}
