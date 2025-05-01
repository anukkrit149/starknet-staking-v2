package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NethermindEth/juno/utils"
	"github.com/NethermindEth/starknet-staking-v2/validator"
	configP "github.com/NethermindEth/starknet-staking-v2/validator/config"
	"github.com/spf13/cobra"
)

func NewCommand() cobra.Command {
	var configPath string
	var logLevelF string

	var config configP.Config
	var snConfig configP.StarknetConfig
	var logger utils.ZapLogger

	preRunE := func(cmd *cobra.Command, args []string) error {
		// Config takes the values from flags directly,
		// then fills the missing ones from the env vars
		configFromEnv := configP.FromEnv()
		config.Fill(&configFromEnv)

		// It fills the missing one from the ones defined
		// in a config file
		if configPath != "" {
			configFromFile, err := configP.FromFile(configPath)
			if err != nil {
				return err
			}
			config.Fill(&configFromFile)
		}

		if err := config.Check(); err != nil {
			return err
		}

		var logLevel utils.LogLevel
		if err := logLevel.Set(logLevelF); err != nil {
			return err
		}
		loadedLogger, err := utils.NewZapLogger(logLevel, true)
		if err != nil {
			return err
		}
		logger = *loadedLogger

		return nil
	}

	run := func(cmd *cobra.Command, args []string) {
		if err := validator.Attest(&config, &snConfig, logger); err != nil {
			logger.Error(err)
		}
	}

	cmd := cobra.Command{
		Use:     "validator",
		Short:   "Program for Starknet validators to attest to epochs with respect to Staking v2",
		PreRunE: preRunE,
		Run:     run,
		Args:    cobra.NoArgs,
	}

	// Config file path flag
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to JSON config file")

	// Config provider flags
	cmd.Flags().StringVar(&config.Provider.Http, "provider-http", "", "Provider http address")
	cmd.Flags().StringVar(&config.Provider.Ws, "provider-ws", "", "Provider ws address")
	// Config signer flags
	cmd.Flags().StringVar(
		&config.Signer.ExternalUrl,
		"signer-url",
		"",
		"Signer url address, required if using an external signer",
	)
	cmd.Flags().StringVar(
		&config.Signer.PrivKey, "signer-priv-key", "", "Signer private key, required for signing",
	)
	cmd.Flags().StringVar(
		&config.Signer.OperationalAddress,
		"signer-op-address",
		"",
		"Signer operational address, required for attesting",
	)
	// Config starknet flags
	cmd.Flags().StringVar(
		&snConfig.ContractAddresses.Attest,
		"attest-contract-address",
		"",
		"Staking contract address. Defaults values are provided for Sepolia and Mainnet",
	)
	cmd.Flags().StringVar(
		&snConfig.ContractAddresses.Staking,
		"staking-contract-address",
		"",
		"Staking contract address. Defaults values are provided for Sepolia and Mainnet",
	)
	cmd.Flags().Uint64Var(
		&snConfig.AttestOptions.Fee,
		"attest-fee",
		0,
		"Predefined fee to pay for each attest transaction."+
			" If not provided, a call to estimate fee is done according to"+
			"`estimate-attest-fee` flag value.",
	)
	cmd.Flags().StringVar(
		&snConfig.AttestOptions.Recalculate,
		"estimate-atttest-fee",
		"once",
		"When to perform an estimate fee call to know the cost of performing an attestation"+
			" if no value is provided in the `attest-fee` flag. Options:\n"+
			" - \"once\": attest fee is estimated once and succesive calls use that value.\n"+
			" - \"always\": an estimate fee call is done before submitting each attestation.",
	)

	// Other flags
	cmd.Flags().StringVar(
		&logLevelF, "log-level", utils.INFO.String(), "Options: trace, debug, info, warn, error.",
	)

	return cmd
}

func main() {
	command := NewCommand()
	if err := command.ExecuteContext(context.Background()); err != nil {
		fmt.Println("Unexpected error:\n", err)
		os.Exit(1)
	}
}
