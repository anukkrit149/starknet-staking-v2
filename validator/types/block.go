package types

import "github.com/NethermindEth/juno/core/felt"

type BlockHash felt.Felt

func (b *BlockHash) Felt() *felt.Felt {
	return (*felt.Felt)(b)
}

func (b *BlockHash) String() string {
	return (*felt.Felt)(b).String()
}

type BlockNumber uint64

// Delete this method
func (b BlockNumber) Uint64() uint64 {
	return uint64(b)
}
