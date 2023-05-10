package eth

import (
	"context"

	"github.com/youngqqcn/arbitrum/core"
	"github.com/youngqqcn/arbitrum/core/state"
	"github.com/youngqqcn/arbitrum/core/types"
	"github.com/youngqqcn/arbitrum/core/vm"
	"github.com/youngqqcn/arbitrum/eth/tracers"
	"github.com/youngqqcn/arbitrum/ethdb"
)

func NewArbEthereum(
	blockchain *core.BlockChain,
	chainDb ethdb.Database,
) *Ethereum {
	return &Ethereum{
		blockchain: blockchain,
		chainDb:    chainDb,
	}
}

func (eth *Ethereum) StateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (core.Message, vm.BlockContext, *state.StateDB, tracers.StateReleaseFunc, error) {
	return eth.stateAtTransaction(ctx, block, txIndex, reexec)
}
