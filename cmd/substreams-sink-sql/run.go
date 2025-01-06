package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	sink "github.com/streamingfast/substreams-sink"
	"github.com/streamingfast/substreams-sink-sql/sinker"
	"github.com/streamingfast/substreams/manifest"
)

type ignoreUndoBufferSize struct{}

func (i ignoreUndoBufferSize) IsIgnored(in string) bool {
	return in == "undo-buffer-size"
}

var sinkRunCmd = Command(sinkRunE,
	"run <dsn> <manifest> [<start>:<stop>]",
	"Runs SQL sink process",
	RangeArgs(2, 3),
	Flags(func(flags *pflag.FlagSet) {
		sink.AddFlagsToSet(flags, ignoreUndoBufferSize{})
		AddCommonSinkerFlags(flags)

		flags.Int("undo-buffer-size", 0, "If non-zero, handling of reorgs in the database is disabled. Instead, a buffer is introduced to only process a blocks once it has been confirmed by that many blocks, introducing a latency but slightly reducing the load on the database when close to head.")
		flags.Int("batch-block-flush-interval", 1_000, "When in catch up mode, flush every N blocks or after batch-row-flush-interval, whichever comes first. Set to 0 to disable and only use batch-row-flush-interval. Ineffective if the sink is now in the live portion of the chain where only 'live-block-flush-interval' applies.")
		flags.Int("batch-row-flush-interval", 100_000, "When in catch up mode, flush every N rows or after batch-block-flush-interval, whichever comes first. Set to 0 to disable and only use batch-block-flush-interval. Ineffective if the sink is now in the live portion of the chain where only 'live-block-flush-interval' applies.")
		flags.Int("live-block-flush-interval", 1, "When processing in live mode, flush every N blocks.")
		flags.Int("flush-interval", 0, "(deprecated) please use --batch-block-flush-interval instead")
		flags.StringP("endpoint", "e", "", "Specify the substreams endpoint, ex: `mainnet.eth.streamingfast.io:443`")
	}),
	Example("substreams-sink-sql run 'postgres://localhost:5432/posgres?sslmode=disable' uniswap-v3@v0.2.10"),
	OnCommandErrorLogAndExit(zlog),
)

func sinkRunE(cmd *cobra.Command, args []string) error {
	app := NewApplication(cmd.Context())

	sink.RegisterMetrics()
	sinker.RegisterMetrics()

	dsn := args[0]
	manifestPath := args[1]
	blockRange := ""
	if len(args) > 2 {
		blockRange = args[2]
	}

	endpoint := sflags.MustGetString(cmd, "endpoint")
	if endpoint == "" {
		network := sflags.MustGetString(cmd, "network")
		if network == "" {
			reader, err := manifest.NewReader(manifestPath)
			if err != nil {
				return fmt.Errorf("setup manifest reader: %w", err)
			}
			pkgBundle, err := reader.Read()
			if err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}
			network = pkgBundle.Package.Network
		}
		var err error
		endpoint, err = manifest.ExtractNetworkEndpoint(network, sflags.MustGetString(cmd, "endpoint"), zlog)
		if err != nil {
			return err
		}
	}

	handleReorgs := sflags.MustGetInt(cmd, "undo-buffer-size") == 0

	sink, err := sink.NewFromViper(
		cmd,
		supportedOutputTypes,
		endpoint,
		manifestPath,
		sink.InferOutputModuleFromPackage,
		blockRange,
		zlog,
		tracer,
	)
	if err != nil {
		return fmt.Errorf("new base sinker: %w", err)
	}

	batchBlockFlushInterval := sflags.MustGetInt(cmd, "batch-block-flush-interval")
	if sflags.MustGetInt(cmd, "flush-interval") != 0 {
		batchBlockFlushInterval = sflags.MustGetInt(cmd, "flush-interval")
	}

	dbLoader, err := newDBLoader(cmd, dsn, batchBlockFlushInterval, sflags.MustGetInt(cmd, "batch-row-flush-interval"), sflags.MustGetInt(cmd, "live-block-flush-interval"), handleReorgs)
	if err != nil {
		return fmt.Errorf("new db loader: %w", err)
	}

	postgresSinker, err := sinker.New(sink, dbLoader, zlog, tracer)
	if err != nil {
		return fmt.Errorf("unable to setup postgres sinker: %w", err)
	}

	app.SuperviseAndStart(postgresSinker)

	return app.WaitForTermination(zlog, 0*time.Second, 30*time.Second)
}
