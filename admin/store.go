package admin

import (
	"context"
	"fmt"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/samber/lo"
)

const (
	kilobyte = 1024
	megabyte = 1024 * 1024
	gigabyte = 1024 * 1024 * 1024
	terabyte = 1024 * 1024 * 1024 * 1024
)

type TableSizeData struct {
	Schema          string `json:"schema"`
	Table           string `json:"table"`
	RelSizeBytes    int64  `json:"rel_size_bytes"`
	RelSizePretty   string `json:"rel_size_pretty"`
	IndexSizeBytes  int64  `json:"index_size_bytes"`
	IndexSizePretty string `json:"index_size_pretty"`
	TotalSizeBytes  int64  `json:"total_size_bytes"`
	TotalSizePretty string `json:"total_size_pretty"`
	RowsEstimate    int64  `json:"rows_estimate"`
}

func GetTableSizes(ctx context.Context, dbtx db.DBTX) ([]*TableSizeData, error) {
	rows, err := db.New(dbtx).TableSizesAndRows(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Map(rows, func(row *db.TableSizesAndRowsRow, _ int) *TableSizeData {
		return &TableSizeData{
			Schema:          row.Schema,
			Table:           row.Table,
			RelSizeBytes:    row.RelSizeBytes,
			RelSizePretty:   formatByteSize(row.RelSizeBytes),
			IndexSizeBytes:  row.IndexSizeBytes,
			IndexSizePretty: formatByteSize(row.IndexSizeBytes),
			TotalSizeBytes:  row.TotalSizeBytes,
			TotalSizePretty: formatByteSize(row.TotalSizeBytes),
			RowsEstimate:    int64(row.RowsEstimate),
		}
	}), nil
}

func formatByteSize(size int64) string {
	if size < kilobyte {
		return fmt.Sprintf("%d B", size)
	}
	sizeFloat := float64(size)
	if size < megabyte {
		return fmt.Sprintf("%.2f KiB", sizeFloat/kilobyte)
	}
	if size < gigabyte {
		return fmt.Sprintf("%.2f MiB", sizeFloat/megabyte)
	}
	if size < terabyte {
		return fmt.Sprintf("%.2f GiB", sizeFloat/gigabyte)
	}
	return fmt.Sprintf("%.2f TiB", sizeFloat/terabyte)
}
