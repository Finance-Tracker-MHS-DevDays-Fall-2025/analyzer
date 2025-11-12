package database

import (
	"context"
	"fmt"
)

type TableInfo struct {
	TableName string
	Columns   []ColumnInfo
}

type ColumnInfo struct {
	Name     string
	DataType string
	Nullable string
}

func (db *Database) GetTablesSchema(ctx context.Context) ([]TableInfo, error) {
	query := `
		SELECT 
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable
		FROM information_schema.columns c
		JOIN information_schema.tables t 
			ON c.table_name = t.table_name 
			AND c.table_schema = t.table_schema
		WHERE c.table_schema = 'public'
			AND t.table_type = 'BASE TABLE'
		ORDER BY c.table_name, c.ordinal_position
	`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	tablesMap := make(map[string]*TableInfo)

	for rows.Next() {
		var tableName, columnName, dataType, nullable string

		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := tablesMap[tableName]; !exists {
			tablesMap[tableName] = &TableInfo{
				TableName: tableName,
				Columns:   []ColumnInfo{},
			}
		}

		tablesMap[tableName].Columns = append(tablesMap[tableName].Columns, ColumnInfo{
			Name:     columnName,
			DataType: dataType,
			Nullable: nullable,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	tables := make([]TableInfo, 0, len(tablesMap))
	for _, table := range tablesMap {
		tables = append(tables, *table)
	}

	return tables, nil
}
