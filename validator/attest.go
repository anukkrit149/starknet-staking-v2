package validator

import (
	"context"
	"strconv"
	"time"

	"github.com/NethermindEth/juno/utils"
	"github.com/NethermindEth/starknet-staking-v2/validator/config"
	signerP "github.com/NethermindEth/starknet-staking-v2/validator/signer"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/conc"
)

// Main execution loop of the program. Listens to the blockchain and sends
// attest invoke when it's the right time
func Attest(
	config *config.Config, snConfig *config.StarknetConfig, logger utils.ZapLogger,
) error {
	provider, err := NewProvider(config.Provider.Http, &logger)
	if err != nil {
		return err
	}

	var signer signerP.Signer
	if config.Signer.External() {
		externalSigner, err := signerP.NewExternalSigner(
			provider, &config.Signer, &snConfig.ContractAddresses,
		)
		if err != nil {
			return err
		}
		signer = &externalSigner

	} else {
		internalSigner, err := signerP.NewInternalSigner(
			provider, &logger, &config.Signer, &snConfig.ContractAddresses,
		)
		if err != nil {
			return err
		}
		signer = &internalSigner
	}

	dispatcher := NewEventDispatcher[signerP.Signer]()
	wg := conc.NewWaitGroup()
	wg.Go(func() { dispatcher.Dispatch(signer, &logger) })
	defer wg.Wait()
	defer close(dispatcher.AttestRequired)

	return RunBlockHeaderWatcher(config, &logger, signer, &dispatcher, wg)
}

func RunBlockHeaderWatcher[Account signerP.Signer](
	config *config.Config,
	logger *utils.ZapLogger,
	signer Account,
	dispatcher *EventDispatcher[Account],
	wg *conc.WaitGroup,
) error {
	cleanUp := func(wsProvider *rpc.WsProvider, headersFeed chan *rpc.BlockHeader) {
		wsProvider.Close()
		close(headersFeed)
	}

	for {
		wsProvider, headersFeed, clientSubscription, err := SubscribeToBlockHeaders(config.Provider.Ws, logger)
		if err != nil {
			return err
		}

		stopProcessingHeaders := make(chan error)

		wg.Go(func() {
			err := ProcessBlockHeaders(headersFeed, signer, logger, dispatcher)
			if err != nil {
				stopProcessingHeaders <- err
			}
		})

		select {
		case err := <-clientSubscription.Err():
			logger.Errorw("Error in block header subscription", "error", err)
			logger.Debugw("Ending headers subscription, closing websocket connection, and retrying...")
			cleanUp(wsProvider, headersFeed)
		case err := <-stopProcessingHeaders:
			cleanUp(wsProvider, headersFeed)
			return err
		}
	}
}

func ProcessBlockHeaders[Account signerP.Signer](
	headersFeed chan *rpc.BlockHeader,
	account Account,
	logger *utils.ZapLogger,
	dispatcher *EventDispatcher[Account],
) error {
	noEpochSwitch := func(*EpochInfo, *EpochInfo) bool { return true }
	epochInfo, attestInfo, err := FetchEpochAndAttestInfoWithRetry(
		account, logger, nil, noEpochSwitch, "at app startup",
	)
	if err != nil {
		return err
	}

	SetTargetBlockHashIfExists(account, logger, &attestInfo)

	for blockHeader := range headersFeed {
		logger.Infow("Block header received", "block header", blockHeader)

		if blockHeader.BlockNumber == epochInfo.CurrentEpochStartingBlock.Uint64()+epochInfo.EpochLen {
			logger.Infow("New epoch start", "epoch id", epochInfo.EpochId+1)
			prevEpochInfo := epochInfo
			epochInfo, attestInfo, err = FetchEpochAndAttestInfoWithRetry(
				account,
				logger,
				&prevEpochInfo,
				CorrectEpochSwitch,
				strconv.FormatUint(prevEpochInfo.EpochId+1, 10),
			)
			if err != nil {
				return err
			}
		}

		if BlockNumber(blockHeader.BlockNumber) == attestInfo.TargetBlock {
			attestInfo.TargetBlockHash = BlockHash(*blockHeader.BlockHash)
			logger.Infow(
				"Target block reached",
				"block number", blockHeader.BlockNumber,
				"block hash", blockHeader.BlockHash,
			)
			logger.Infow("Window to attest to",
				"start", attestInfo.WindowStart,
				"end", attestInfo.WindowEnd,
			)
		}

		if BlockNumber(blockHeader.BlockNumber) >= attestInfo.WindowStart-1 &&
			BlockNumber(blockHeader.BlockNumber) < attestInfo.WindowEnd {
			dispatcher.AttestRequired <- AttestRequired{
				BlockHash: attestInfo.TargetBlockHash,
			}
		}

		if BlockNumber(blockHeader.BlockNumber) == attestInfo.WindowEnd {
			dispatcher.EndOfWindow <- struct{}{}
		}
	}

	return nil
}

func SetTargetBlockHashIfExists[Account signerP.Signer](
	account Account,
	logger *utils.ZapLogger,
	attestInfo *AttestInfo,
) {
	targetBlockNumber := attestInfo.TargetBlock.Uint64()
	res, err := account.BlockWithTxHashes(
		context.Background(), rpc.BlockID{Number: &targetBlockNumber},
	)

	// If no error, then target block already exists
	if err == nil {
		if block, ok := res.(*rpc.BlockTxHashes); ok {
			attestInfo.TargetBlockHash = BlockHash(*block.BlockHash)
			logger.Infow(
				"Target block already exists, registered block hash to attest to.",
				"block hash", attestInfo.TargetBlockHash.String(),
			)
		}
	}
}

func FetchEpochAndAttestInfoWithRetry[Account signerP.Signer](
	account Account,
	logger *utils.ZapLogger,
	prevEpoch *EpochInfo,
	isEpochSwitchCorrect func(prevEpoch *EpochInfo, newEpoch *EpochInfo) bool,
	newEpochId string,
) (EpochInfo, AttestInfo, error) {
	newEpoch, newAttestInfo, err := signerP.FetchEpochAndAttestInfo(account, logger)

	for err != nil || !isEpochSwitchCorrect(prevEpoch, &newEpoch) {
		if err != nil {
			logger.Debugw("Failed to fetch epoch info", "epoch id", newEpochId, "error", err.Error())
		} else {
			logger.Debugw("Wrong epoch switch", "from epoch", prevEpoch, "to epoch", &newEpoch)
		}
		logger.Debugw("Retrying to fetch epoch info...")
		Sleep(time.Second)
		newEpoch, newAttestInfo, err = signerP.FetchEpochAndAttestInfo(account, logger)
	}

	// if err != nil {
	// 	return EpochInfo{},
	// 		AttestInfo{},
	// 		errors.Errorf(
	// 			"Failed to fetch epoch info for epoch id %s: %s", newEpochId, err.Error(),
	// 		)
	if !isEpochSwitchCorrect(prevEpoch, &newEpoch) {
		return EpochInfo{},
			AttestInfo{},
			errors.Errorf("Wrong epoch switch: from epoch %s to epoch %s", prevEpoch, &newEpoch)
	}

	return newEpoch, newAttestInfo, nil
}

func CorrectEpochSwitch(prevEpoch *EpochInfo, newEpoch *EpochInfo) bool {
	return newEpoch.EpochId == prevEpoch.EpochId+1 &&
		newEpoch.CurrentEpochStartingBlock.Uint64() == prevEpoch.CurrentEpochStartingBlock.Uint64()+prevEpoch.EpochLen
}
