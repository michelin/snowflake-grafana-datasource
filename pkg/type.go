package main

import (
	"database/sql"
)

type DataQueryResult struct {
	Tables []DataTable
}

type DataTable struct {
	Columns []*sql.ColumnType
	Rows    [][]interface{}
}
