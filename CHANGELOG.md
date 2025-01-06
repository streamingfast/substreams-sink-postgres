# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v4.4.0

* Added support for the Clickhouse `Date` type.

* Fixed handling of the Clickhouse `Array` type.

* Removed the check for duplicate primary keys on the Clickhouse dialect. This allows inserting multiple rows with the same primary key.

## v4.3.0

* Added a check for non-existent columns in Clickhouse
* Added support for `Nullable` types in Clickhouse.
* Bump substreams to `v0.11.1`
* Bump substreams-sink to `v0.5.0`

## v4.2.2

* Fix major bug when receiving empty `MapOutput`

## v4.2.1

* Bump substreams to v1.10.3 to support new manifest data like `protobuf:excludePaths`

## v4.2.0

* Added the --cursor-table and --history-table flags to allow running to sinks on the same database (be careful that you have no collision in table names)
* bumped substreams to v1.8.2, add some default network endpoints

## v4.1.0

* Bumped substreams to v1.7.3
* Enable gzip compression on substreams data on the wire

## v4.0.3

* Fix another case where 'infinite-retry' would not work and the program would stop on an error.
* Enable multiple Substreams authentication methods (API key, JWT), using flags `--api-key-envvar` and `--api-token-envvar`.
* Deprecates the use of `SF_API_TOKEN` environment variable, now use default `SUBSTREAMS_API_TOKEN` or set your own using `--api-token-envvar`.

## v4.0.2

* Fixed spurious error reporting when the sinker is terminating or has been canceled.

* Updated `substreams` dependency to latest version `v1.3.7`.

## v4.0.1

* Fixed the timestamp parsing in Clickhouse dialect.
* Fixed the schema in the tutorial for clickhouse.
* Add `--network` flag to override the default value in the manifest or spkg

## v4.0.0

* Bump version of `schema` dependency to fix errors with new Clickhouse versions now using `system.tables` table instead of `information_schema.tables` view.

## v4.0.0-rc.3

* Fix a critical bug breaking the reorg management when more than one row needs to be reverted.

## v4.0.0-rc.2

> :warning: This release candidate contains a critical bug in the reorg management and should not be used. Upgrade immediately to v4.0.0-rc.3

* Support more networks with default mappings (ex: solana, soon: optimism, soon: bitcoin)
* Add command "create-user <dsn> <username> <database>" to help creating more SQL users, read-only or otherwise
* add 'enabled' field under "DBTConfig"
* Removed PgwebFrontend and WireProtocolAccess fields from the SinkConfig message: they will now be deployed when on a development environment, so they are not mentionned here anymore.

## v4.0.0-rc.1

> :warning: This release candidate contains a critical bug in the reorg management and should not be used. Upgrade immediately to v4.0.0-rc.3

### Fixes

* Fix an issue preventing the `setup` command from running on a clickhouse backend because of the reorg settings.

## v4.0.0-beta

> :warning: This release candidate contains a critical bug in the reorg management and should not be used. Upgrade immediately to v4.0.0-rc.3

### Highlights

* This release brings support for managing reorgs in Postgres database, enabled by default when `--undo-buffer-size` to 0.

### Breaking changes

* A change in your SQL schema may be required to keep existing substreams:SQL integrations working:
  * The presence of a primary key (single key or composite) is now *MANDATORY* on every table.
  * The `sf.substreams.sink.database.v1.TableChange` message, generated inside substreams, must now exactly match its primary key with the one in the SQL schema.
  * You will need to re-run `setup` on your existing PostgreSQL databases to add the `substreams_history` table. You can use the new `--system-tables-only` flag to perform only that.

* Since reorgs management is not yet supported on Clickhouse, users will have to set `--undo-buffer-size` to a non-zero value (`12` was the previous default)

## Protodefs v1.0.4

* Added support for `rest_frontend` field with `enabled` boolean flag, aimed at this backend implementation: <https://github.com/semiotic-ai/sql-wrapper>

## v3.0.5

* Fixed regression: `run` command was incorrectly only processing blocks staying behind the "FinalBlocks" cliff.

## v3.0.4

* Fixed support for tables with primary keys misaligned with `database_changes`'s keys (fixing Clickhouse use case)

## Protodefs v1.0.3

* Added support for selecting engine `postgres` or `clickhouse` to sinkconfig protobuf definition

## v3.0.3

* Fixed missing uint8 and uint16 cast for clickhouse driver conversion

## v3.0.2

* Fixed default endpoint sfor networks 'goerli' and 'mumbai'
* Added `--postgraphile` flag to `setup`, which will add a @skip comment on cursor table so Postgraphile doesn't try to serve cursors (it resulted in a name collision with Postgraphile internal names)
* Fixed a bug with Clickhouse driver where different integer sizes need explicit conversion

## v3.0.1

### Fixed

* Fixed an issue where the schema encoded in the SinkConfig part of a manifest would not be encoded correctly, leading to garbled (base64) bytes being sent to the SQL server instead of the schema.

## v3.0.0

### Highlights

This release brings a major refactoring enabling support for multiple database drivers and not just Postgres anymore. Our first newly supported driver is [Clickhouse](https://clickhouse.com/#getting_started) which defines itself as *The fastest and most resource efficient open-source database for real-time apps and analytics*. In the future, further database driver could be supported like MySQL, MSSQL and any other that can talk the SQL protocol.

Now that we support multiple driver, keeping the `substreams-sink-postgres` didn't make sense anymore. As such, we have renamed the project from `substreams-sink-postgresql` to `substreams-sink-sql` since it now supports Clickhouse out of the box. The binary and Go modules have been renamed in consequence.

Another major change brought by this release is the usage of Substreams "Deployable Unit". What we call a "Deployable Unit" is a Substreams manifest that fully defines a deployment packaged as a single artifact. This change how the sink is operated; the SQL schema, output module and "Network" identifer are now passed in the "SinkConfig" section of the Substreams manifest instead of being accepted at command line.

Read the **Operators** section below to learn how to migrate to this new version.

#### Operators

Passing the `schema` and the `module_name` to the `run` and `setup` commands is no longer accepted via arguments, they need to be written to the `substreams.yaml` file.

Before:

```bash
substreams-sink-sql setup "psql://..." "path/to/schema.sql"
substreams-sink-sql run "psql://..." mainnet.eth.streamingfast.io:443 https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.5.1/substreams-eth-block-meta-v0.5.1.spkg db_out [<range>]
```

Now:

* Create a deployable unit file, let's call it `substreams.prod.yaml` with content:

```yaml
specVersion: v0.1.0
package:
  name: "<name>"
  version: v0.0.1

imports:
  sql: https://github.com/streamingfast/substreams-sink-sql/releases/download/protodefs-v1.0.1/substreams-sink-sql-protodefs-v1.0.1.spkg
  main: https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.5.1/substreams-eth-block-meta-v0.5.1.spkg

network: mainnet

sink:
  module: main:db_out
  type: sf.substreams.sink.sql.v1.Service
  config:
    schema: "./path/to/schema.sql"
```

In this `<name>` is the same name as what `<manifest>` defines was, `https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.5.1/substreams-eth-block-meta-v0.5.1.spkg` is the current manifest you deploy.

The `./path/to/schema.sql` would point to your schema file (path resolved relative to parent directory of `substreams.prod.yaml`).

The 'network: mainnet' will be used to resolve to an endpoint. You can configure each network to have its own endpoint via environment variables `SUBSTREAMS_ENDPOINTS_CONFIG_<NETWORK>` or override this mechanism completely by using the `--endpoint` flag. Most used networks have default endpoints.

* Setup your database:

```bash
substreams-sink-sql setup <dsn> substreams.prod.yaml
```

* Run the sink:

```bash
substreams-sink-sql run <dsn> substreams.prod.yaml
```

Similar changes have been applied to other commands as well.

## v2.5.4

### Added

* Added average flush duration to sink stats.

* Added log line when flush time to database is `> 5s` in `INFO` and in `WARN` if `> 30s`.

### Fixed

* Fixed `pprof` HTTP routes not properly registered.

### Changed

* Renamed metric Prometheus metric `substreams_sink_postgres_flushed_entries_count` to `substreams_sink_postgres_flushed_rows_count`, adjust your dashboard if needed and change it found a `Gauge` to a `Counter`.

## v2.5.3

* Refactored internal code to support multiple database drivers.

* **Experimental** `clickhouse` is now supported as a new `clickhouse` is now supported* Added driver abstraction

  You can connect to Clickhouse by using the following DSN:

  * Not encrypted: `clickhouse://<host>:9000/<database>?username=<user>&password=<password>`
  * Encrypted: `clickhouse://<host>:9440/<database>?secure=true&skip_verify=true&username=<user>&password=<password>`

  If you want to send custom args to the connection, you can use by sending as query params.

## v2.5.2

### Changed

* Bumped `logging` library to latest version which should fixed problem where containerized workload are not printing logs out in JSON format.

## v2.5.1

This is a bug fix release containing a fix for inserting rows into a table for which no primary key constraint exist. For now, we still requires internally that your provide an `id` in your `DatabaseChange` of your row, a future update will lift that limitations.

## v2.5.0

### Highlights

This releases brings improvements to reported progress message while your Substreams executes which should greatly enhanced progression tracking

> [!NOTE]
> Stay tuned, we are planning even more useful progression tracking now that we've updated progression data sent back to the client!

This releases also introduces a new mode to dump data in the database at high speed, useful for large amount of data insertion.

### Substreams Progress Messages

Bumped [substreams-sink](https://github.com/streamingfast/substreams-sink) [v0.3.1](https://github.com/streamingfast/substreams-sink/releases/tag/v0.3.1) and [substreams](https://github.com/streamingfast/substreams) to [v1.1.12](https://github.com/streamingfast/substreams/releases/tag/v1.1.12) to support the new progress message format. Progression now relates to **stages** instead of modules. You can get stage information using the `substreams info` command starting from version `v1.1.12`.

> [!IMPORTANT]
> This client only support progress messages sent from a server using `substreams` version `>=v1.1.12`

#### Changed Prometheus Metrics

* `substreams_sink_progress_message` removed in favor of `substreams_sink_progress_message_total_processed_blocks`
* `substreams_sink_progress_message_last_end_block` removed in favor of `substreams_sink_progress_message_last_block` (per stage)

#### Added Prometheus Metrics

* Added `substreams_sink_progress_message_last_contiguous_block` (per stage)
* Added `substreams_sink_progress_message_running_jobs`(per stage)

### New injection method

A new injection method has been added to this `substreams-sink-postgres` release. It's a 2 steps method that leverage `COPY FROM` SQL operations to inject at high speed a great quantity of data.

> [!NOTE]
> This method will be useful if you insert a lot of data into the database. If the standard ingestion speed satisfy your needs, continue to use it, the new feature is an advanced use case.

See the [High Throughput Injection section](https://github.com/streamingfast/substreams-sink-postgres/blob/develop/README.md#high-throughput-injection) of the `README.md` file to check how to use it.

### Added

* Added newer method of populating the database via CSV (thanks [@gusinacio](https://github.com/gusinacio)!).

  Newer commands:
  * `generate-csv`: Generates CSVs for each table
  * `inject-csv`: Injects generated CSV rows for `<table>`

## v2.4.0

### Changed

* gRPC `InvalidArgument` error(s) are not retried anymore like specifying and invalid start block or argument in your request.

* **Breaking** Flag shorthand `-p` for `--plaintext` has been re-assigned to Substreams params definition, to align with `substreams run/gui` on that aspect. There is no shorthand anymore for `--plaintext`.

  If you were using before `-p`, please convert to `--plaintext`.

  > **Note** We expect that this is affecting very few users as `--plaintext` is usually used only on developers machine.

### Added

* Added support for `--params, -p` (can be repeated multiple times) on the form `-p <module>=<value>`.

## v2.3.4

### Added

* Added logging of new `Session` received values (`linear_handoff_block`, `max_parallel_workers` and `resolved_start_block`).

* Added `--header, -H` (can be repeated multiple times) flag to pass extra headers to the server.

### Changed

* Now reporting available columns when an unknown column is encountered.

## v2.3.3

### Fixed

* Batches written to the database now respects the insertion ordering has received from your Substreams. This fixes for example auto-increment to be as defined on the chain.

## v2.3.2

### Fixed

* Fixed problem where string had unicode character and caused `pq: invalid message format`.

## v2.3.1

### Fixed

* The `substreams-sink-postgres setup` command has been fixed to use the correct schema defined by the DSN.

* The `cursors` table suggestion when the table is not found has been updated to be in-sync with table used in `substreams-sink-postgres setup`.

### Changed

* Now using Go Protobuf generate bindings from <https://github.com/streamingfast/substreams-sink-database-changes>.

## v2.3.0

### Added

* Added `Composite keys` support following the update in `substreams-database-change`

  The code was updated to use `oneOf` primary keys (pk and composite) to keep backward compatibility. Therefore, Substreams using older versions of `DatabaseChange` can still use newer versions of `postgres-sink` without problems. To use composite key, define your schema to use Postgres composite keys, update to latest version of `substreams-database-changes` and update your code to send a `CompositePrimaryKey` key object for the `primary_key` field of the `TableChange` message.

* Added escape to value in case the postgres data type is `BYTES`. We now escape the byte array.

### Fixed

* Added back support for old Substreams Database Change Protobuf package id `sf.substreams.database.v1.DatabaseChanges`.

## v2.2.1

### Changed

* Reduced the amount of allocations and escaping performed which should increase ingestion speed, this will be more visible for Substreams where a lot of entities and columns are processed.

### Fixed

* The `schema` is correctly respected now for the the `cursors` table.

## v2.2.0

### Highlights

#### Cursor Bug Fix

It appeared that the cursor was not saved properly until the first graceful shutdown of `substreams-sink-postgres`. Furthermore, the on exit save was actually wrong because it was saving the cursor without flushing accumulated data which is wrong (e.g. that we had N blocks in memory unflushed and a cursor, and we were saving this cursor to the database without having flushed the in memory logic).

This bug has been introduced in v2.0.0 by mistake which means if we synced a new database with v2.0.0+, there is a good chance your are actually missing some data in your database. It's highly recommended that you re-synchronize your database from scratch.

> **Note** If your are using the same `.spkg` that you are using right now, database ingestion from scratch should go at very high speed because you will be reading from previously cached output, so the bottleneck should be network and the database write performance.

#### Behavior on `.spkg` update

In the release, we change a big how cursor is associated to the `<module>`'s hash in the database and how it's stored.

Prior this version, when loading the cursor back from the database on restart, we were retrieving the cursor associated to the `<module>`'s hash received by `substreams-sink-postgres run`. The consequence of that is that if you change the `.spkg` version you were sinking with, on restart we would find no cursor since the module's hash of this new `.spkg` would have changed and which you mean a full sync back would be happening because we would start without a cursor.

This silent behavior is problematic because it could seen like the cursor was lost somehow while actually, we just picked up a new one from scratch because the `.spkg` changed.

This release brings in a new flag `substreams-sink-postgres run --on-module-hash-mistmatch=error` (default value shown) where it would control how we should react to a changes in the module's hash since last run.

* If `error` is used (default), it will exit with an error explaining the problem and how to fix it.
* If `warn` is used, it does the same as 'ignore' but it will log a warning message when it happens.
* If `ignore` is set, we pick the cursor at the highest block number and use it as the starting point. Subsequent updates to the cursor will overwrite the module hash in the database.

There is a possibility that multiple cursors exists in your database, hence why we pick the one with the highest block. If it's the case, you will be warned that multiple cursors exists. You can run `substreams-sink-postgres tools cursor cleanup <manifest> <module> --dsn=<dsn>` which will delete now useless cursors.

The `ignore` value can be used to change to a new `.spkg` while retaining the previous data in the database, the database schema will start to be different after a certain point where the new `.spkg` became active.

### Added

* Added `substreams-sink-postgres run --on-module-hash-mistmatch=error` to control how a change in module's hash should be handled.

### Changed

* Changed behavior of how cursor are retrieved on restart.

### Fixed

* Fixed `cursor` not being saved correctly until the binary exits.

* Fixed wrong handling of updating the cursor, we were not checking if a row was updated when doing the flush operation.

* Fixed a bug where it was possible if the sink was terminating to write a cursor for data not yet flushed. This was happening if the `substreams-sink-postgres run` was stopped before we ever written a cursor, which normally happens each 1000 blocks. We don't expect anybody to have been hit by this but if you are unsure, you should check data for the 1000 first blocks of you sink (for example from 11 000 000 to 11 001 000 if your module start block was 11 000 000).

## v2.1.0

### Changed

* Column's schema type that are not known by the `sql` library we know will now be transferred as-is to the database.

  There is a lot of column's for which the `sql` library we use have to Go representation for by default. This is the case for example for the `numeric` column's type. Previously, this would be reported directly as an error, Now, we pass the received value from your Substreams unmodified to the database engine. It will be your responsibility to send the data in the right format accepted by the database. We send the value as-is, without escaping and without sanitization, so this is a risk if you don't control the Substreams.

### Added

* When doing `substreams-sink-postgres run`, the `<manifest>` argument now accepts directory like `.`.

### Fixed

* Fixed timestamp received in RFC3339 format.

## v2.0.2

### Changed

* Diminish amount of allocations done to perform fields transformation.

### Fixed

* Fixed some places where escaping for either identifier or value was not done properly.

* Fixed double escaping of boolean values.

## v2.0.1

### Added

* Added proper escaping for table & column names to allow keyword column names to use keywords as column names such as `to` and `from` etc.

## v2.0.0

### Highlights

This release drops support for Substreams RPC protocol `sf.substreams.v1` and switch to Substreams RPC protocol `sf.substreams.rpc.v2`. As a end user, right now the transition is seamless. All StreamingFast endpoints have been updated to to support the legacy Substreams RPC protocol `sf.substreams.v1` as well as the newer Substreams RPC protocol `sf.substreams.rpc.v2`.

Support for legacy Substreams RPC protocol `sf.substreams.v1` is expected to end by June 6 2023. What this means is that you will need to update to at least this release if you are running `substreams-sink-postgres` in production. Otherwise, after this date, your current binary will stop working and will return errors that `sf.substreams.v1.Blocks` is not supported on the endpoint.

From a database and operator standpoint, this binary is **fully** backward compatible with your current schema. Updating to this binary will continue to sink just like if you used a prior release.

#### Retryable Errors

The errors coming from Postgres are **not** retried anymore and will stop the binary immediately.

#### Operators

If you were using environment variable to configure the binary, note that the environment prefix has changed from `SINK_` to `SINK_POSTGRES_`.

### Changed

* **Deprecated** The flag `--irreversible-only` is deprecated, use `--final-blocks-only` instead.

### Added

* Added command `substreams-sink-postgres tools --dsn <dsn> cursor read` to read the current cursors stored in your database.

* **Dangerous** Added command `substreams-sink-postgres tools --dsn <dsn> cursor write <module_hash> <cursor>` to update the cursor in your database for the given `<module_hash>`

    > **Warning** This is a destructive operation, be sure you understand the consequences of updating the cursor.

* **Dangerous** Added command `substreams-sink-postgres tools --dsn <dsn> cursor delete [<module_hash>|--all]` to delete the cursor associated with the given module's hash or all cursors if `--all` is used.

    > **Warning** This is a destructive operation, be sure you understand the consequences of updating the cursor.

## v1.0.0

### Highlights

This is the latest release before upgrading to Substreams RPC v2.

### Added

* Added `--infinite-retry` to never exit on error and retry indefinitely instead.

* Added `--development-mode` to run in development mode.

    > **Warning** You should use that flag for testing purposes, development mode drastically reduce performance you get from the server.

* Added `--irreversible-only` to only deal with final (irreversible) blocks.
