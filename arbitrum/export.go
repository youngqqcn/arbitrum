package arbitrum

import (
	"context"

	"github.com/youngqqcn/arbitrum/common/hexutil"
	"github.com/youngqqcn/arbitrum/core"
	"github.com/youngqqcn/arbitrum/internal/ethapi"
	"github.com/youngqqcn/arbitrum/rpc"
)

type TransactionArgs = ethapi.TransactionArgs

func EstimateGas(ctx context.Context, b ethapi.Backend, args TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, gasCap uint64) (hexutil.Uint64, error) {
	return ethapi.DoEstimateGas(ctx, b, args, blockNrOrHash, gasCap)
}

func NewRevertReason(result *core.ExecutionResult) error {
	return ethapi.NewRevertError(result)
}
