package main

import (
	"database/sql"
)

type DataQueryResult struct {
	Tables []DataTable
}

// DataTable structure containing columns and rows
type DataTable struct {
	Columns []*sql.ColumnType
	Rows    [][]interface{}
}
