package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/streamingfast/cli/sflags"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/substreams-sink-sql/db"
)

var createUserCmd = Command(createUserE,
	"create-user <dsn> <username> <database>",
	"Create a user in the database",
	ExactArgs(3),
	Flags(func(flags *pflag.FlagSet) {
		flags.Int("retries", 3, "Number of retries to attempt when a connection error occurs")
		flags.Bool("read-only", false, "Create a read-only user")
		flags.String("password-env", "", "Name of the environment variable containing the password")
	}),
)

func createUserE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	dsn := args[0]
	username := args[1]
	database := args[2]

	readOnly := sflags.MustGetBool(cmd, "read-only")
	passwordEnv := sflags.MustGetString(cmd, "password-env")

	if passwordEnv == "" {
		return fmt.Errorf("password-env is required")
	}

	password := os.Getenv(passwordEnv)
	if password == "" {
		return fmt.Errorf("non-empty password is required")
	}

	if err := retry(ctx, func(ctx context.Context) error {
		dbLoader, err := db.NewLoader(dsn, 0, 0, 0, db.OnModuleHashMismatchError, nil, zlog, tracer)
		if err != nil {
			return fmt.Errorf("new psql loader: %w", err)
		}

		err = dbLoader.CreateUser(ctx, username, password, database, readOnly)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		return nil
	}, sflags.MustGetInt(cmd, "retries")); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func retry(ctx context.Context, f func(ctx context.Context) error, reties int) error {
	var err error

	for i := 0; i < reties; i++ {
		err = f(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(5*time.Duration(i)*time.Second + 1*time.Second)
	}

	return err
}
