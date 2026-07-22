package benzstorage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const dbTimeout = 5 * time.Second

// Queue score constants — lower is better.
const (
	QueueUpTo30    = "up_to_30"
	QueueFrom30to60 = "from_30_to_60"
	QueueMoreThan60 = "more_than_60"

	scoreUpTo30    = 1
	scoreFrom30to60 = 2
	scoreMoreThan60 = 3
	scoreUnknown    = 4

	dayStartHour  = 6  // day period: 6:00–21:59
	dayEndHour    = 21
)

func queueToScore(queue string) int {
	switch queue {
	case QueueUpTo30:
		return scoreUpTo30
	case QueueFrom30to60:
		return scoreFrom30to60
	case QueueMoreThan60:
		return scoreMoreThan60
	default:
		return scoreUnknown
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// SnapshotInput is one station's state at collection time.
type SnapshotInput struct {
	AzsID     int
	Address   string
	AzsName   string
	IsWorking bool
	Has95     bool
	Queue     string
}

// PeriodBest holds the best hour found for a given time period.
type PeriodBest struct {
	Hour     int
	AvgQueue float64
	HasData  bool
}

// DayStats holds analytics for a single station on a given day of week.
// DayOfWeek follows Go's time.Weekday: 0=Sunday, 1=Monday … 6=Saturday.
// Has95 percentages are computed separately for each time period.
type DayStats struct {
	DayOfWeek     int
	DayHas95Pct   float64 // 0–100, has_95 share during day period (6:00–21:59)
	NightHas95Pct float64 // 0–100, has_95 share during night period (22:00–5:59)
	Day           PeriodBest
	Night         PeriodBest
}

// GasAnalytics is the aggregated analytics result for one gas station.
type GasAnalytics struct {
	AzsID           int
	Address         string
	AzsName         string
	OverallHas95Pct float64
	OverallAvgQueue float64
	DayStats        []DayStats // ordered Sun → Sat, only days with data
	TotalData       int
}

type Storage struct {
	db *sqlx.DB
}

// NewStorage opens (or creates) the SQLite database and initialises the schema.
// The parent directory of dbPath is created automatically if it does not exist.
func NewStorage(dbPath string) (*Storage, error) {
	if dir := filepath.Dir(dbPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating benz db directory: %w", err)
		}
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("connecting to benz database: %w", err)
	}

	s := &Storage{db: db}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := s.initSchema(ctx); err != nil {
		return nil, fmt.Errorf("initialising benz schema: %w", err)
	}
	return s, nil
}

func (s *Storage) initSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS gas_snapshots (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			azs_id      INTEGER NOT NULL,
			address     TEXT    NOT NULL,
			azs_name    TEXT    NOT NULL,
			is_working  INTEGER NOT NULL DEFAULT 0,
			has_95      INTEGER NOT NULL DEFAULT 0,
			queue       TEXT    NOT NULL DEFAULT '',
			queue_score INTEGER NOT NULL DEFAULT 4,
			day_of_week INTEGER NOT NULL,
			hour        INTEGER NOT NULL,
			recorded_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gs_azs      ON gas_snapshots(azs_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gs_analytics ON gas_snapshots(azs_id, day_of_week, hour)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// InsertSnapshots writes one snapshot per station in a single transaction.
func (s *Storage) InsertSnapshots(ctx context.Context, stations []SnapshotInput) error {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	hour := now.Hour()
	recordedAt := now.Unix()

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `INSERT INTO gas_snapshots
		(azs_id, address, azs_name, is_working, has_95, queue, queue_score, day_of_week, hour, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, st := range stations {
		score := scoreUnknown
		if st.IsWorking {
			score = queueToScore(st.Queue)
		}
		if _, err := tx.ExecContext(ctx, q,
			st.AzsID, st.Address, st.AzsName,
			boolToInt(st.IsWorking), boolToInt(st.Has95),
			st.Queue, score,
			dayOfWeek, hour, recordedAt,
		); err != nil {
			return fmt.Errorf("inserting snapshot azs_id=%d: %w", st.AzsID, err)
		}
	}

	return tx.Commit()
}

// GetTotalSnapshots returns the total number of stored snapshots.
func (s *Storage) GetTotalSnapshots(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM gas_snapshots`)
	return count, err
}

// — internal query result types —

type stationRow struct {
	AzsID     int     `db:"azs_id"`
	Address   string  `db:"address"`
	AzsName   string  `db:"azs_name"`
	AvgQueue  float64 `db:"avg_queue"`
	Has95Pct  float64 `db:"has95_pct"`
	TotalData int     `db:"total_data"`
}

type slotRow struct {
	DayOfWeek int     `db:"day_of_week"`
	Hour      int     `db:"hour"`
	AvgScore  float64 `db:"avg_score"`
}

type dayHas95Row struct {
	DayOfWeek     int     `db:"day_of_week"`
	DayHas95Pct   float64 `db:"day_has95_pct"`
	NightHas95Pct float64 `db:"night_has95_pct"`
}

// GetTop10 returns the top-10 stations ranked by has_95 availability (desc) then
// average queue score (asc), with per-day day/night optimal times.
func (s *Storage) GetTop10(ctx context.Context) ([]GasAnalytics, error) {
	const summaryQ = `
		SELECT
			azs_id,
			address,
			azs_name,
			COALESCE(AVG(CASE WHEN is_working = 1 THEN queue_score ELSE NULL END), 4) AS avg_queue,
			COALESCE(
				CAST(SUM(CASE WHEN is_working = 1 AND has_95 = 1 THEN 1 ELSE 0 END) AS FLOAT)
				* 100.0 / NULLIF(SUM(CASE WHEN is_working = 1 THEN 1 ELSE 0 END), 0),
				0
			) AS has95_pct,
			COUNT(*) AS total_data
		FROM gas_snapshots
		GROUP BY azs_id
		ORDER BY has95_pct DESC, avg_queue ASC
		LIMIT 10`

	var rows []stationRow
	if err := s.db.SelectContext(ctx, &rows, summaryQ); err != nil {
		return nil, fmt.Errorf("querying top-10 stations: %w", err)
	}

	result := make([]GasAnalytics, 0, len(rows))
	for _, r := range rows {
		dayStats, err := s.getDayStats(ctx, r.AzsID)
		if err != nil {
			return nil, fmt.Errorf("day stats for azs_id=%d: %w", r.AzsID, err)
		}
		result = append(result, GasAnalytics{
			AzsID:           r.AzsID,
			Address:         r.Address,
			AzsName:         r.AzsName,
			OverallHas95Pct: r.Has95Pct,
			OverallAvgQueue: r.AvgQueue,
			DayStats:        dayStats,
			TotalData:       r.TotalData,
		})
	}
	return result, nil
}

// getDayStats builds per-day-of-week analytics with separate day/night optimal hours.
func (s *Storage) getDayStats(ctx context.Context, azsID int) ([]DayStats, error) {
	// Best hour in day period (6–21) per day-of-week
	daySlots, err := s.getBestSlots(ctx, azsID, true)
	if err != nil {
		return nil, err
	}

	// Best hour in night period (22–5) per day-of-week
	nightSlots, err := s.getBestSlots(ctx, azsID, false)
	if err != nil {
		return nil, err
	}

	// has_95 percentage per day-of-week split by time period (day 6–21, night 22–5).
	// Both columns computed in a single query to avoid a second round-trip.
	const has95Q = `
		SELECT
			day_of_week,
			COALESCE(
				CAST(SUM(CASE WHEN (hour >= 6 AND hour <= 21) AND has_95 = 1 THEN 1 ELSE 0 END) AS FLOAT)
				* 100.0 / NULLIF(SUM(CASE WHEN hour >= 6 AND hour <= 21 THEN 1 ELSE 0 END), 0),
				0
			) AS day_has95_pct,
			COALESCE(
				CAST(SUM(CASE WHEN (hour >= 22 OR hour <= 5) AND has_95 = 1 THEN 1 ELSE 0 END) AS FLOAT)
				* 100.0 / NULLIF(SUM(CASE WHEN hour >= 22 OR hour <= 5 THEN 1 ELSE 0 END), 0),
				0
			) AS night_has95_pct
		FROM gas_snapshots
		WHERE azs_id = ? AND is_working = 1
		GROUP BY day_of_week
		ORDER BY day_of_week ASC`

	var has95Rows []dayHas95Row
	if err := s.db.SelectContext(ctx, &has95Rows, has95Q, azsID); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("querying has95 per day: %w", err)
	}

	has95ByDay := make(map[int]dayHas95Row, 7)
	for _, r := range has95Rows {
		has95ByDay[r.DayOfWeek] = r
	}

	// Collect all days that have any data
	daysWithData := make(map[int]bool)
	for d := range daySlots {
		daysWithData[d] = true
	}
	for d := range nightSlots {
		daysWithData[d] = true
	}
	for d := range has95ByDay {
		daysWithData[d] = true
	}

	// Build result in canonical order: Sun(0) … Sat(6)
	stats := make([]DayStats, 0, len(daysWithData))
	for dow := 0; dow <= 6; dow++ {
		if !daysWithData[dow] {
			continue
		}

		h95 := has95ByDay[dow]
		ds := DayStats{
			DayOfWeek:     dow,
			DayHas95Pct:   h95.DayHas95Pct,
			NightHas95Pct: h95.NightHas95Pct,
		}
		if slot, ok := daySlots[dow]; ok {
			ds.Day = PeriodBest{Hour: slot.Hour, AvgQueue: slot.AvgScore, HasData: true}
		}
		if slot, ok := nightSlots[dow]; ok {
			ds.Night = PeriodBest{Hour: slot.Hour, AvgQueue: slot.AvgScore, HasData: true}
		}
		stats = append(stats, ds)
	}
	return stats, nil
}

// getBestSlots returns the best hour per day-of-week for the day (isDayPeriod=true)
// or night period (isDayPeriod=false), keyed by day_of_week.
func (s *Storage) getBestSlots(ctx context.Context, azsID int, isDayPeriod bool) (map[int]slotRow, error) {
	var periodCond string
	if isDayPeriod {
		periodCond = fmt.Sprintf("AND hour >= %d AND hour <= %d", dayStartHour, dayEndHour)
	} else {
		periodCond = fmt.Sprintf("AND (hour >= %d OR hour <= %d)", dayEndHour+1, dayStartHour-1)
	}

	q := fmt.Sprintf(`
		SELECT day_of_week, hour, AVG(queue_score) AS avg_score
		FROM gas_snapshots
		WHERE azs_id = ? AND is_working = 1 %s
		GROUP BY day_of_week, hour
		ORDER BY day_of_week ASC, avg_score ASC`, periodCond)

	var rows []slotRow
	if err := s.db.SelectContext(ctx, &rows, q, azsID); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("querying best slots (day=%v): %w", isDayPeriod, err)
	}

	// Keep only the first (= best) slot per day-of-week
	best := make(map[int]slotRow, 7)
	for _, r := range rows {
		if _, exists := best[r.DayOfWeek]; !exists {
			best[r.DayOfWeek] = r
		}
	}
	return best, nil
}
