package extracteddata

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/solher/forklift/files"
	"github.com/solher/hunterio-test/lib/pgutil"
)

var ErrNotFound = errors.New("extracted data not found")

// NewPostgresRepository returns a Postgres backed repository.
func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{
		db: db,
	}
}

type postgresRepository struct {
	db *pgxpool.Pool
}

// Search allows object searching.
type Search struct {
	URL string `db:"url"`
}

func (r *postgresRepository) Insert(ctx context.Context, extractedData *ExtractedData) (*ExtractedData, error) {
	if extractedData.URL == "" {
		return nil, errors.New("url cannot be empty")
	}

	cpy := *extractedData
	extractedData = &cpy

	extractedData.CreatedAt = time.Now().UTC()
	extractedData.UpdatedAt = extractedData.CreatedAt

	if err := r.db.QueryRow(ctx, files.File("insert.tmpl.sql"), pgutil.ToNamedArgs(extractedData)).Scan(&extractedData.ID); err != nil {
		return nil, err
	}
	return extractedData, nil
}

func (r *postgresRepository) Find(ctx context.Context, search Search) (extractedDataList []ExtractedData, err error) {
	rows, err := r.db.Query(ctx, files.Template("find.lazy.sql", search), pgutil.ToNamedArgs(search))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[ExtractedData])
}

func (r *postgresRepository) GetByURL(ctx context.Context, url string) (*ExtractedData, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	extractedDataList, err := r.Find(ctx, Search{URL: url})
	if err != nil {
		return nil, err
	}
	if len(extractedDataList) == 0 {
		return nil, ErrNotFound
	}
	return &extractedDataList[0], nil
}

func (r *postgresRepository) UpdateByID(ctx context.Context, id uint64, extractedData *ExtractedData) error {
	if id == 0 {
		return errors.New("id cannot be empty")
	}
	if extractedData.URL == "" {
		return errors.New("url cannot be empty")
	}

	cpy := *extractedData
	extractedData = &cpy
	extractedData.ID = id

	_, err := r.db.Exec(ctx, files.File("update_by_id.tmpl.sql"), pgutil.ToNamedArgs(extractedData))
	return err
}
