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

// lastCursorFilename is the name of the file where the last cursor is stored, no extension as it's added by the store
const lastCursorFilename = "last_cursor"

var generateCsvCmd = Command(generateCsvE,
	"generate-csv <dsn> <manifest> [start]:<stop>",
	"Generates CSVs for each table so it can be bulk inserted with `inject-csv` (for postgresql only)",
	Description(`
		This command command is the first of a multi-step process to bulk insert data into a PostgreSQL database.
		It creates a folder for each table and generates CSVs for block ranges. This files can be used with
		the 'inject-csv' command to bulk insert data into the database.

		It needs that the database already exists and that the schema is already created.

		The process is as follows:

		- Generate CSVs for each table with this command
		- Inject the CSVs into the database with the 'inject-csv' command (contains 'cursors' table, double check you injected it correctly!)
		- Start streaming with the 'run' command
	`),
	ExactArgs(3),
	Flags(func(flags *pflag.FlagSet) {
		sink.AddFlagsToSet(flags, sink.FlagIgnore("final-blocks-only"))
		AddCommonSinkerFlags(flags)

		flags.Uint64("bundle-size", 10000, "Size of output bundle, in blocks")
		flags.String("working-dir", "./workdir", "Path to local folder used as working directory")
		flags.String("output-dir", "./csv-output", "Path to local folder used as destination for CSV")
		flags.StringP("endpoint", "e", "", "Specify the substreams endpoint, ex: `mainnet.eth.streamingfast.io:443`")
		flags.Uint64("buffer-max-size", 4*1024*1024, FlagDescription(`
			Amount of memory bytes to allocate to the buffered writer. If your data set is small enough that every is hold in memory, we are going to avoid
			the local I/O operation(s) and upload accumulated content in memory directly to final storage location.

			Ideally, you should set this as to about 80%% of RAM the process has access to. This will maximize amount of element in memory,
			and reduce 'syscall' and I/O operations to write to the temporary file as we are buffering a lot of data.

			This setting has probably the greatest impact on writing throughput.

			Default value for the buffer is 4 MiB.
		`))
	}),
	OnCommandErrorLogAndExit(zlog),
)

func generateCsvE(cmd *cobra.Command, args []string) error {
	app := NewApplication(cmd.Context())

	sink.RegisterMetrics()
	sinker.RegisterMetrics()

	dsn := args[0]
	manifestPath := args[1]
	blockRange := args[2]

	outputDir := sflags.MustGetString(cmd, "output-dir")
	bundleSize := sflags.MustGetUint64(cmd, "bundle-size")
	bufferMaxSize := sflags.MustGetUint64(cmd, "buffer-max-size")
	workingDir := sflags.MustGetString(cmd, "working-dir")

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

	sink, err := sink.NewFromViper(
		cmd,
		supportedOutputTypes,
		endpoint,
		manifestPath,
		sink.InferOutputModuleFromPackage,
		blockRange,
		zlog,
		tracer,
		sink.WithFinalBlocksOnly(),
	)
	if err != nil {
		return fmt.Errorf("new base sinker: %w", err)
	}

	dbLoader, err := newDBLoader(cmd, dsn, 0, 0, 0, false) // flush interval not used in CSV mode
	if err != nil {
		return fmt.Errorf("new db loader: %w", err)
	}

	generateCSVSinker, err := sinker.NewGenerateCSVSinker(
		sink,
		outputDir,
		workingDir,
		bundleSize,
		bufferMaxSize,
		dbLoader,
		lastCursorFilename,
		zlog,
		tracer,
	)
	if err != nil {
		return fmt.Errorf("unable to setup generate csv sinker: %w", err)
	}

	app.Supervise(generateCSVSinker.Shutter)

	go func() {
		generateCSVSinker.Run(app.Context())
	}()

	return app.WaitForTermination(zlog, 0*time.Second, 30*time.Second)
}
