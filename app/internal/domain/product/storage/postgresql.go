package storage

import sq "github.com/Masterminds/squirrel"

type ProductStorage struct {
	queryBuilder sq.StatementBuilderType
	client       PostgresSQLClient
}
