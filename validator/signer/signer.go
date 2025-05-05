package signer

import (
	"context"
	"math/big"

	"github.com/NethermindEth/juno/core/crypto"
	"github.com/NethermindEth/juno/core/felt"
	junoUtils "github.com/NethermindEth/juno/utils"
	"github.com/NethermindEth/starknet-staking-v2/validator/constants"
	"github.com/NethermindEth/starknet-staking-v2/validator/types"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"lukechampine.com/uint128"
)

//go:generate go tool mockgen -destination=../../mocks/mock_signer.go -package=mocks github.com/NethermindEth/starknet-staking-v2/validator/signer Signer
type Signer interface {
	// Methods from Starknet.go Account implementation
	GetTransactionStatus(
		ctx context.Context, transactionHash *felt.Felt,
	) (*rpc.TxnStatusResult, error)
	BuildAndSendInvokeTxn(
		ctx context.Context, functionCalls []rpc.InvokeFunctionCall, multiplier float64,
	) (*rpc.AddInvokeTransactionResponse, error)
	Call(ctx context.Context, call rpc.FunctionCall, blockId rpc.BlockID) ([]*felt.Felt, error)
	BlockWithTxHashes(ctx context.Context, blockID rpc.BlockID) (interface{}, error)

	// Custom Methods
	Address() *Address
	ValidationContracts() *ValidationContracts
}

// I believe all these functions down here should be methods
// Postponing for now to not affect test code

func FetchEpochInfo[S Signer](signer S) (EpochInfo, error) {
	functionCall := rpc.FunctionCall{
		ContractAddress: signer.ValidationContracts().Staking.Felt(),
		EntryPointSelector: utils.GetSelectorFromNameFelt(
			"get_attestation_info_by_operational_address",
		),
		Calldata: []*felt.Felt{signer.Address().Felt()},
	}

	result, err := signer.Call(context.Background(), functionCall, rpc.BlockID{Tag: "latest"})
	if err != nil {
		return EpochInfo{},
			entrypointInternalError("get_attestation_info_by_operational_address", err)
	}

	if len(result) != 5 {
		return EpochInfo{}, entrypointResponseError("get_attestation_info_by_operational_address")
	}

	stake := result[1].Bits()
	return EpochInfo{
		StakerAddress:             Address(*result[0]),
		Stake:                     uint128.New(stake[0], stake[1]),
		EpochLen:                  result[2].Uint64(),
		EpochId:                   result[3].Uint64(),
		CurrentEpochStartingBlock: BlockNumber(result[4].Uint64()),
	}, nil
}

func FetchAttestWindow[S Signer](signer S) (uint64, error) {
	result, err := signer.Call(
		context.Background(),
		rpc.FunctionCall{
			ContractAddress:    signer.ValidationContracts().Attest.Felt(),
			EntryPointSelector: utils.GetSelectorFromNameFelt("attestation_window"),
			Calldata:           []*felt.Felt{},
		},
		rpc.BlockID{Tag: "latest"},
	)
	if err != nil {
		return 0, entrypointInternalError("attestation_window", err)
	}

	if len(result) != 1 {
		return 0, entrypointResponseError("attestation_window")
	}

	return result[0].Uint64(), nil
}

// For near future when tracking validator's balance
func FetchValidatorBalance[Account Signer](account Account) (Balance, error) {
	StrkTokenContract := types.AddressFromString(constants.STRK_CONTRACT_ADDRESS)
	result, err := account.Call(
		context.Background(),
		rpc.FunctionCall{
			ContractAddress:    StrkTokenContract.Felt(),
			EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
			Calldata:           []*felt.Felt{account.Address().Felt()},
		},
		rpc.BlockID{Tag: "latest"},
	)
	if err != nil {
		return Balance{}, entrypointInternalError("balanceOf", err)
	}

	if len(result) != 1 {
		return Balance{}, entrypointResponseError("balanceOf")
	}

	return Balance(*result[0]), nil
}

func FetchEpochAndAttestInfo[S Signer](
	signer S, logger *junoUtils.ZapLogger,
) (EpochInfo, AttestInfo, error) {
	epochInfo, err := FetchEpochInfo(signer)
	if err != nil {
		return EpochInfo{}, AttestInfo{}, err
	}
	logger.Debugw(
		"Fetched epoch info",
		"epoch ID", epochInfo.EpochId,
		"epoch starting block", epochInfo.CurrentEpochStartingBlock,
		"epoch ending block", epochInfo.CurrentEpochStartingBlock+BlockNumber(epochInfo.EpochLen),
	)

	attestWindow, windowErr := FetchAttestWindow(signer)
	if windowErr != nil {
		return EpochInfo{}, AttestInfo{}, windowErr
	}

	blockNum := ComputeBlockNumberToAttestTo(&epochInfo, attestWindow)

	attestInfo := AttestInfo{
		TargetBlock: blockNum,
		WindowStart: blockNum + BlockNumber(constants.MIN_ATTESTATION_WINDOW),
		WindowEnd:   blockNum + BlockNumber(attestWindow),
	}

	logger.Infof(
		"Target block to attest to at %d. Attestation window: %d <> %d",
		attestInfo.TargetBlock.Uint64(),
		attestInfo.WindowStart.Uint64(),
		attestInfo.WindowEnd.Uint64(),
	)
	logger.Debugw(
		"Epoch and Attestation info",
		"Epoch", epochInfo,
		"Attestation", attestInfo,
	)
	return epochInfo, attestInfo, nil
}

func InvokeAttest[S Signer](signer S, attest *AttestRequired) (
	*rpc.AddInvokeTransactionResponse, error,
) {
	calls := []rpc.InvokeFunctionCall{{
		ContractAddress: signer.ValidationContracts().Attest.Felt(),
		FunctionName:    "attest",
		CallData:        []*felt.Felt{attest.BlockHash.Felt()},
	}}

	return signer.BuildAndSendInvokeTxn(
		context.Background(), calls, constants.FEE_ESTIMATION_MULTIPLIER,
	)
}

func ComputeBlockNumberToAttestTo(epochInfo *EpochInfo, attestWindow uint64) BlockNumber {
	hash := crypto.PoseidonArray(
		new(felt.Felt).SetBigInt(epochInfo.Stake.Big()),
		new(felt.Felt).SetUint64(epochInfo.EpochId),
		epochInfo.StakerAddress.Felt(),
	)

	hashBigInt := new(big.Int)
	hashBigInt = hash.BigInt(hashBigInt)

	blockOffset := new(big.Int)
	blockOffset = blockOffset.Mod(hashBigInt, big.NewInt(int64(epochInfo.EpochLen-attestWindow)))

	return BlockNumber(epochInfo.CurrentEpochStartingBlock.Uint64() + blockOffset.Uint64())
}
