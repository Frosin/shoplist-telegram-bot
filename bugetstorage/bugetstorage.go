package bugetstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

const (
	dbTimeout  = time.Second * 5
	bugetDB    = "buget"
	categoryDB = "category"
	noteDB     = "note"
)

var (
	txOptions = sql.TxOptions{Isolation: sql.LevelSerializable}
)

type Buget struct {
	ID      int
	Title   string
	Created int64
}

type Category struct {
	ID      int
	BugetID int
	Title   string
	Current int64
	Target  int64
}

type Note struct {
	ID         int
	CategoryID int
	Sum        int
	Title      string
	Created    int64
}

type Storage struct {
	db *sqlx.DB
}

func NewStorage() (Storage, error) {
	bugetPath := viper.GetString("SHOPLIST-BOT_BUGETPATH")
	db, err := sqlx.Connect("sqlite3", bugetPath)
	if err != nil {
		return Storage{}, err
	}
	return Storage{
		db: db,
	}, nil
}

func (s Storage) InsertBuget(ctx context.Context, title string) error {
	q, args, err := squirrel.
		Insert(bugetDB).
		Columns("title", "created").
		Values(title, time.Now().Unix()).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &txOptions)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(q, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Storage) GetBuget(ctx context.Context, ID int) (Buget, error) {
	q, args, err := squirrel.
		Select("id", "title", "created").
		From(bugetDB).
		Where(squirrel.Eq{"id": ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return Buget{}, err
	}
	buget := Buget{}
	row := s.db.QueryRowContext(ctx, q, args...)
	err = row.Scan(&buget.ID, &buget.Title, &buget.Created)
	if err != nil {
		return Buget{}, err
	}
	return buget, nil
}

func (s Storage) GetLastBugets(ctx context.Context, num uint64) ([]Buget, error) {
	q, args, err := squirrel.
		Select("id", "title", "created").
		From(bugetDB).
		OrderBy("created DESC").
		Limit(num).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	bugets := []Buget{}
	buget := Buget{}
	rows, err := s.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&buget.ID, &buget.Title, &buget.Created)
		if err != nil {
			return nil, err
		}
		bugets = append(bugets, buget)
	}
	return bugets, nil
}

func (s Storage) InsertCategory(ctx context.Context, category Category) error {
	q, args, err := squirrel.
		Insert(categoryDB).
		Columns(
			"buget_id", "title",
			"current", "target",
		).
		Values(
			category.BugetID, category.Title,
			category.Current, category.Target,
		).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &txOptions)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(q, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Storage) UpdateCategory(ctx context.Context, categoryID int, sum int) error {
	q, args, err := squirrel.
		Update(categoryDB).
		Set("current", sum).
		Where(squirrel.Eq{"id": categoryID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &txOptions)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(q, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Storage) GetBugetCategories(ctx context.Context, bugetID int) ([]Category, error) {
	q, args, err := squirrel.
		Select("id", "buget_id", "title", "current", "target").
		From(categoryDB).
		Where(squirrel.Eq{"buget_id": bugetID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	categories := []Category{}
	category := Category{}
	rows, err := s.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(
			&category.ID, &category.BugetID,
			&category.Title, &category.Current,
			&category.Target,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (s Storage) GetCategory(ctx context.Context, ID int) (Category, error) {
	q, args, err := squirrel.
		Select("id", "buget_id", "title", "current", "target").
		From(categoryDB).
		Where(squirrel.Eq{"id": ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return Category{}, err
	}
	category := Category{}
	row := s.db.QueryRowContext(ctx, q, args...)
	err = row.Scan(
		&category.ID, &category.BugetID,
		&category.Title, &category.Current,
		&category.Target,
	)
	if err != nil {
		return Category{}, err
	}
	return category, nil
}

func (s Storage) InsertNote(ctx context.Context, note Note) error {
	q, args, err := squirrel.
		Insert(noteDB).
		Columns(
			"category_id",
			"title",
			"sum",
			"created",
		).
		Values(
			note.CategoryID,
			note.Title,
			note.Sum,
			time.Now().Unix(),
		).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &txOptions)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(q, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Storage) GetCategoryNotes(ctx context.Context, categoryID int) ([]Note, error) {
	q, args, err := squirrel.
		Select("id", "category_id", "title", "sum", "created").
		From(noteDB).
		Where(squirrel.Eq{"category_id": categoryID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	notes := []Note{}
	note := Note{}
	rows, err := s.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(
			&note.ID, &note.CategoryID,
			&note.Title, &note.Sum, &note.Created,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s Storage) GetBugetNotes(ctx context.Context, bugetID int) ([]Note, error) {

	subQ := fmt.Sprintf("category_id in (select id from %s where buget_id=%d)", categoryDB, bugetID)

	q, args, err := squirrel.
		Select("id", "category_id", "title", "sum", "created").
		From(noteDB).
		Where(subQ).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	notes := []Note{}
	note := Note{}
	rows, err := s.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(
			&note.ID, &note.CategoryID,
			&note.Title, &note.Sum, &note.Created,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}
