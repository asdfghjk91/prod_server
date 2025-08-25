package storage

import (
	"app/internal/domain/product/model"
	"app/pkg/common/logging"
	"context"

	sq "github.com/Masterminds/squirrel"
)

type ProductStorage struct {
	queryBuilder sq.StatementBuilderType
	client       PostgresSQLClient
	logger       *logging.Logger
}

func (s *ProductStorage) queryLogger(sql, table string, args []interface{}) *logging.Logger {
	return s.logger.ExtraFields(map[string]interface{}{
		"sql":   sql,
		"table": table,
		"args":  args,
	})
}

func NewProductStorage(client PostgresSQLClient, logger *logging.Logger) ProductStorage {
	return ProductStorage{
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		client:       client,
		logger:       logger,
	}
}

const (
	scheme = "public"
	table  = "product"
)

func (s *ProductStorage) All(ctx context.Context) ([]model.Product, error) {
	query := s.queryBuilder.Select("id").
		Column("name").
		Column("description").
		Column("image_id").
		Column("price").
		Column("currency_id").
		Column("rating").
		Column("created_at").
		Column("updated_at").
		From(scheme + "." + table)

	sql, args, err := query.ToSql()
	logger := s.queryLogger(sql, table, args)
	if err != nil {
		logger.Error("failed to create SQL Query due to error: %v", err)
		return nil, err
	}

	logger.Trace("do qurey")
	rows, err := s.client.Query(ctx, sql, args...)
	if err != nil {
		logger.Error("failed to do qurey error: %v", err)
		return nil, err
	}

	defer rows.Close()

	list := make([]model.Product, 0)

	for rows.Next() {
		p := model.Product{}
		if err = rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ImageID, &p.Price, &p.CurrencyID, &p.Rating,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			logger.Error("error scan")
		}
		list = append(list, p)
	}
	return list, nil
}
