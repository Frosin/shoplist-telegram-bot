package bugetstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

const (
	dbTimeout  = time.Second * 5
	bugetDB    = "buget"
	categoryDB = "category"
	noteDB     = "note"

	fundsBudgetID = -1
)

var (
	txOptions = sql.TxOptions{Isolation: sql.LevelSerializable}
)

type Budget struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	Created int64  `db:"created"`
}

type Category struct {
	ID          int    `db:"id"`
	BugetID     int    `db:"buget_id"`
	Title       string `db:"title"`
	Current     int    `db:"current"`
	CashCurrent int    `db:"cash_current"`
	Target      int    `db:"target"`
}

type PaymentMethod byte

const (
	PaymentMethodCard PaymentMethod = 0
	PaymentMethodCash PaymentMethod = 1
)

type Note struct {
	ID            int           `db:"id"`
	CategoryID    int           `db:"category_id"`
	Sum           int           `db:"sum"`
	Title         string        `db:"title"`
	PaymentMethod PaymentMethod `db:"payment_method"`
	Created       int64         `db:"created"`
}

type Storage struct {
	db     *sqlx.DB
	dumper *helpers.Dumper
}

func (m PaymentMethod) String() string {
	if m == 0 {
		return "к"
	}

	return "н"
}

func NewStorage(dumpFn helpers.DumpFn) (Storage, error) {
	budgetPath := viper.GetString("SHOPLIST-BOT_BUGETPATH")
	db, err := sqlx.Connect("sqlite3", budgetPath)
	if err != nil {
		return Storage{}, fmt.Errorf("connecting to database: %w", err)
	}

	dumper := helpers.NewDumper(dumpFn, nil)
	dumper.Start()

	return Storage{
		db:     db,
		dumper: dumper,
	}, nil
}

func (s Storage) InsertBudget(ctx context.Context, title string) error {
	query := `INSERT INTO buget (title, created) VALUES (?, ?)`

	_, err := s.db.ExecContext(ctx, query, title, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("inserting budget: %w", err)
	}

	s.dumper.ScheduleUpdate()
	return nil
}

func (s Storage) GetBudget(ctx context.Context, ID int) (Budget, error) {
	var budget Budget
	query := `SELECT id, title, created FROM buget WHERE id = ?`

	err := s.db.GetContext(ctx, &budget, query, ID)
	if err != nil {
		return Budget{}, fmt.Errorf("getting budget: %w", err)
	}

	return budget, nil
}

func (s Storage) GetLastBudgets(ctx context.Context, num uint) ([]Budget, error) {
	var budgets []Budget
	query := `SELECT id, title, created FROM buget ORDER BY created DESC LIMIT ?`

	err := s.db.SelectContext(ctx, &budgets, query, num)
	if err != nil {
		return nil, fmt.Errorf("getting last budgets: %w", err)
	}

	return budgets, nil
}

func (s Storage) InsertCategory(ctx context.Context, category Category) error {

	query := `
		INSERT INTO category (buget_id, title, current, cash_current, target) 
		VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		category.BugetID, category.Title,
		category.Current, category.CashCurrent, category.Target,
	)
	if err != nil {
		return fmt.Errorf("inserting category: %w", err)
	}

	s.dumper.ScheduleUpdate()
	return nil
}

func (s Storage) InsertFund(ctx context.Context, category Category) error {
	query := `
		INSERT INTO category (buget_id, title, current, target) 
		VALUES (?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		fundsBudgetID, category.Title, category.Current, 0,
	)
	if err != nil {
		return fmt.Errorf("inserting fund: %w", err)
	}

	s.dumper.ScheduleUpdate()
	return nil
}

func (s Storage) UpdateCategory(ctx context.Context, categoryID int, current, cashCurrent int) error {
	tx, err := s.db.BeginTxx(ctx, &txOptions)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `UPDATE category SET current = ?, cash_current = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, current, cashCurrent, categoryID)
	if err != nil {
		return fmt.Errorf("updating category: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	s.dumper.ScheduleUpdate()
	return nil
}

func (s Storage) GetBudgetCategories(ctx context.Context, budgetID int) ([]Category, error) {
	var categories []Category
	query := `
		SELECT id, buget_id, title, current, cash_current, target 
		FROM category WHERE buget_id = ?`

	err := s.db.SelectContext(ctx, &categories, query, budgetID)
	if err != nil {
		return nil, fmt.Errorf("getting budget categories: %w", err)
	}

	return categories, nil
}

func (s Storage) GetFunds(ctx context.Context) ([]Category, error) {
	var categories []Category
	query := `SELECT id, buget_id, title, current FROM category WHERE buget_id = ?`

	err := s.db.SelectContext(ctx, &categories, query, fundsBudgetID)
	if err != nil {
		return nil, fmt.Errorf("getting funds: %w", err)
	}

	return categories, nil
}

func (s Storage) GetCategory(ctx context.Context, ID int) (Category, error) {
	var category Category
	query := `
		SELECT id, buget_id, title, current, cash_current, target 
		FROM category WHERE id = ?`

	err := s.db.GetContext(ctx, &category, query, ID)
	if err != nil {
		return Category{}, fmt.Errorf("getting category: %w", err)
	}

	return category, nil
}

func (s Storage) GetFund(ctx context.Context, ID int) (Category, error) {
	return s.GetCategory(ctx, ID)
}

func (s Storage) InsertNote(ctx context.Context, note Note) error {
	tx, err := s.db.BeginTxx(ctx, &txOptions)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `INSERT INTO note (category_id, title, sum, payment_method, created) VALUES (?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, query,
		note.CategoryID, note.Title, note.Sum, note.PaymentMethod, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("inserting note: %w", err)
	}

	// Обновляем баланс категории
	if err := s.updateCategoryBalance(ctx, tx, note.CategoryID, note.Sum, note.PaymentMethod); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	s.dumper.ScheduleUpdate()
	return nil
}

func (s Storage) updateCategoryBalance(ctx context.Context, tx *sqlx.Tx, categoryID int, sum int, paymentType PaymentMethod) error {
	var category Category
	err := tx.GetContext(ctx, &category, "SELECT current, cash_current FROM category WHERE id = ?", categoryID)
	if err != nil {
		return fmt.Errorf("getting category for update: %w", err)
	}

	newCurrent := category.Current + sum

	newCashCurrent := category.CashCurrent
	if paymentType == PaymentMethodCash {
		newCashCurrent += sum
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE category SET current = ?, cash_current = ? WHERE id = ?",
		newCurrent, newCashCurrent, categoryID,
	)
	if err != nil {
		return fmt.Errorf("updating category balance: %w", err)
	}

	return nil
}

func (s Storage) GetCategoryNotes(ctx context.Context, categoryID int) ([]Note, error) {
	var notes []Note
	query := `SELECT id, category_id, title, sum, payment_method, created FROM note WHERE category_id = ?`

	err := s.db.SelectContext(ctx, &notes, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("getting category notes: %w", err)
	}

	return notes, nil
}

func (s Storage) GetAllBudgets(ctx context.Context) ([]Budget, error) {
	var budgets []Budget
	query := `SELECT id, title, created FROM buget ORDER BY created DESC`

	err := s.db.SelectContext(ctx, &budgets, query)
	if err != nil {
		return nil, fmt.Errorf("getting all budgets: %w", err)
	}

	return budgets, nil
}

func (s Storage) GetBudgetNotes(ctx context.Context, budgetID int) ([]Note, error) {
	var notes []Note
	query := `
		SELECT n.id, n.category_id, n.title, n.sum, n.payment_method, n.created 
		FROM note n
		INNER JOIN category c ON n.category_id = c.id 
		WHERE c.buget_id = ?`

	err := s.db.SelectContext(ctx, &notes, query, budgetID)
	if err != nil {
		return nil, fmt.Errorf("getting budget notes: %w", err)
	}

	return notes, nil
}
