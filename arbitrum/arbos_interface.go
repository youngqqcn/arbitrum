package arbitrum

import (
	"context"

	"github.com/youngqqcn/arbitrum/arbitrum_types"
	"github.com/youngqqcn/arbitrum/core"
	"github.com/youngqqcn/arbitrum/core/types"
)

type ArbInterface interface {
	PublishTransaction(ctx context.Context, tx *types.Transaction, options *arbitrum_types.ConditionalOptions) error
	BlockChain() *core.BlockChain
	ArbNode() interface{}
}
