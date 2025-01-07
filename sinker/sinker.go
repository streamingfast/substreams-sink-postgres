package sinker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	sink "github.com/streamingfast/substreams-sink"
	pbdatabase "github.com/streamingfast/substreams-sink-database-changes/pb/sf/substreams/sink/database/v1"
	"github.com/streamingfast/substreams-sink-sql/db"
	pbsubstreamsrpc "github.com/streamingfast/substreams/pb/sf/substreams/rpc/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const BLOCK_FLUSH_INTERVAL_DISABLED = 0

type SQLSinker struct {
	*shutter.Shutter
	*sink.Sinker

	loader *db.Loader
	logger *zap.Logger
	tracer logging.Tracer

	lastSeenCursor     *sink.Cursor
	lastSeenFinalBlock uint64
	flushMutex         *sync.Mutex

	stats *Stats
}

func New(sink *sink.Sinker, loader *db.Loader, logger *zap.Logger, tracer logging.Tracer) (*SQLSinker, error) {
	return &SQLSinker{
		Shutter: shutter.New(),
		Sinker:  sink,

		loader: loader,
		logger: logger,
		tracer: tracer,

		flushMutex: &sync.Mutex{},

		stats: NewStats(logger),
	}, nil
}

func (s *SQLSinker) Run(ctx context.Context) {
	cursor, mistmatchDetected, err := s.loader.GetCursor(ctx, s.OutputModuleHash())
	if err != nil && !errors.Is(err, db.ErrCursorNotFound) {
		s.Shutdown(fmt.Errorf("unable to retrieve cursor: %w", err))
		return
	}

	// We write an empty cursor right away in the database because the flush logic
	// only performs an `update` operation so an initial cursor is required in the database
	// for the flush to work correctly.
	if errors.Is(err, db.ErrCursorNotFound) {
		if err := s.loader.InsertCursor(ctx, s.OutputModuleHash(), sink.NewBlankCursor()); err != nil {
			s.Shutdown(fmt.Errorf("unable to write initial empty cursor: %w", err))
			return
		}
	} else if mistmatchDetected {
		if err := s.loader.InsertCursor(ctx, s.OutputModuleHash(), cursor); err != nil {
			s.Shutdown(fmt.Errorf("unable to write new cursor after module mistmatch: %w", err))
			return
		}
	}

	s.Sinker.OnTerminating(func(err error) {
		if err == nil {
			s.logger.Info("sql sinker terminating, flushing remaining rows")
			err = s.flush(ctx)
		}
		s.stats.LogNow()
		s.logger.Info("sql sinker terminated", zap.Stringer("last_block_written", s.stats.lastBlock), zap.Error(err))
		s.Shutdown(err)
	})
	s.OnTerminating(s.Sinker.Shutdown)

	s.stats.OnTerminating(func(_ error) { s.stats.Close() })
	s.stats.OnTerminated(func(err error) { s.Shutdown(err) })

	logEach := 15 * time.Second
	if s.logger.Core().Enabled(zap.DebugLevel) {
		logEach = 5 * time.Second
	}

	s.stats.Start(logEach, cursor)

	s.logger.Info("starting sql sink",
		zap.Duration("stats_refresh_each", logEach),
		zap.Stringer("restarting_at", cursor.Block()),
		zap.String("database", s.loader.GetDatabase()),
		zap.String("schema", s.loader.GetSchema()),
	)
	s.Sinker.Run(ctx, cursor, s)
}

func (s *SQLSinker) HandleBlockScopedData(ctx context.Context, data *pbsubstreamsrpc.BlockScopedData, isLive *bool, cursor *sink.Cursor) error {
	output := data.Output

	if output.Name != s.OutputModuleName() {
		return fmt.Errorf("received data from wrong output module, expected to received from %q but got module's output for %q", s.OutputModuleName(), output.Name)
	}

	dbChanges := &pbdatabase.DatabaseChanges{}
	mapOutput := output.GetMapOutput()

	if mapOutput.String() != "" {
		if !mapOutput.MessageIs(dbChanges) && mapOutput.TypeUrl != "type.googleapis.com/sf.substreams.database.v1.DatabaseChanges" {
			return fmt.Errorf("mismatched message type: trying to unmarshal unknown type %q", mapOutput.MessageName())
		}

		// We do not use UnmarshalTo here because we need to parse an older proto type and
		// UnmarshalTo enforces the type check. So we check manually the `TypeUrl` above and we use
		// `Unmarshal` instead which only deals with the bytes value.
		if err := proto.Unmarshal(mapOutput.Value, dbChanges); err != nil {
			return fmt.Errorf("unmarshal database changes: %w", err)
		}

		// We lock here to ensure that blocks are always fully processed and state is updated before any flush can happen.
	s.flushMutex.Lock()
	if err := s.applyDatabaseChanges(dbChanges, data.Clock.Number, data.FinalBlockHeight); err != nil {
		s.flushMutex.Unlock()
			return fmt.Errorf("apply database changes: %w", err)
		}
	}

	s.lastSeenCursor = cursor
	s.lastSeenFinalBlock = data.FinalBlockHeight
	s.flushMutex.Unlock()

	if (s.batchBlockModulo(isLive) > 0 && data.Clock.Number%s.batchBlockModulo(isLive) == 0) || s.loader.FlushNeeded() {
		s.logger.Debug("flushing to database",
			zap.Stringer("block", cursor.Block()),
			zap.Bool("is_live", *isLive),
			zap.Bool("block_flush_interval_reached", s.batchBlockModulo(isLive) > 0 && data.Clock.Number%s.batchBlockModulo(isLive) == 0),
			zap.Bool("row_flush_interval_reached", s.loader.FlushNeeded()),
		)
		return s.flush(ctx)
	}

	return nil
}

func (s *SQLSinker) flush(ctx context.Context) error {

	// we haven't received any data yet, so nothing to do here
	if s.lastSeenCursor == nil {
		return nil
	}

	s.flushMutex.Lock()
	defer s.flushMutex.Unlock()

	flushStart := time.Now()
	rowFlushedCount, err := s.loader.Flush(ctx, s.OutputModuleHash(), s.lastSeenCursor, s.lastSeenFinalBlock)
	if err != nil {
		return fmt.Errorf("failed to flush at block %s: %w", s.lastSeenCursor.Block(), err)
	}

	flushDuration := time.Since(flushStart)
	if flushDuration > 5*time.Second {
		level := zap.InfoLevel
		if flushDuration > 30*time.Second {
			level = zap.WarnLevel
		}

		s.logger.Check(level, "flush to database took a long time to complete, could cause long sync time along the road").Write(zap.Duration("took", flushDuration))
	}

	FlushCount.Inc()
	FlushedRowsCount.AddInt(rowFlushedCount)
	FlushDuration.AddInt64(flushDuration.Nanoseconds())

	s.stats.RecordBlock(s.lastSeenCursor.Block())
	s.stats.RecordFlushDuration(flushDuration)

	return nil
}

func (s *SQLSinker) applyDatabaseChanges(dbChanges *pbdatabase.DatabaseChanges, blockNum, finalBlockNum uint64) error {
	for _, change := range dbChanges.TableChanges {
		if !s.loader.HasTable(change.Table) {
			return fmt.Errorf(
				"your Substreams sent us a change for a table named %s we don't know about on %s (available tables: %s)",
				change.Table,
				s.loader.GetIdentifier(),
				strings.Join(s.loader.GetAvailableTablesInSchema(), ", "),
			)
		}

		var primaryKeys map[string]string
		switch u := change.PrimaryKey.(type) {
		case *pbdatabase.TableChange_Pk:
			var err error
			primaryKeys, err = s.loader.GetPrimaryKey(change.Table, u.Pk)
			if err != nil {
				return err
			}
		case *pbdatabase.TableChange_CompositePk:
			primaryKeys = u.CompositePk.Keys
		default:
			return fmt.Errorf("unknown primary key type: %T", change.PrimaryKey)
		}

		changes := map[string]string{}
		for _, field := range change.Fields {
			changes[field.Name] = field.NewValue
		}

		var reversibleBlockNum *uint64
		if blockNum > finalBlockNum {
			reversibleBlockNum = &blockNum
		}

		switch change.Operation {
		case pbdatabase.TableChange_CREATE:
			err := s.loader.Insert(change.Table, primaryKeys, changes, reversibleBlockNum)
			if err != nil {
				return fmt.Errorf("database insert: %w", err)
			}
		case pbdatabase.TableChange_UPDATE:
			err := s.loader.Update(change.Table, primaryKeys, changes, reversibleBlockNum)
			if err != nil {
				return fmt.Errorf("database update: %w", err)
			}
		case pbdatabase.TableChange_DELETE:
			err := s.loader.Delete(change.Table, primaryKeys, reversibleBlockNum)
			if err != nil {
				return fmt.Errorf("database delete: %w", err)
			}
		default:
			//case database.TableChange_UNSET:
		}
	}
	return nil
}

func (s *SQLSinker) HandleBlockUndoSignal(ctx context.Context, data *pbsubstreamsrpc.BlockUndoSignal, cursor *sink.Cursor) error {
	return s.loader.Revert(ctx, s.OutputModuleHash(), cursor, data.LastValidBlock.Number)
}

func (s *SQLSinker) batchBlockModulo(isLive *bool) uint64 {
	if isLive == nil {
		panic(fmt.Errorf("liveness checker has been disabled on the Sinker instance, this is invalid in the context of 'substreams-sink-sql'"))
	}

	if *isLive {
		return uint64(s.loader.LiveBlockFlushInterval())
	}

	if s.loader.BatchBlockFlushInterval() > 0 {
		return uint64(s.loader.BatchBlockFlushInterval())
	}

	return BLOCK_FLUSH_INTERVAL_DISABLED
}
