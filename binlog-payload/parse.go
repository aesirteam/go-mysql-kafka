package blp

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/schema"
)

// DATA_FORMAT the data format of timestamp
//const DATE_FORMAT = "2006-01-02T15:04:05.000+08:00"

func parseRowMap(columns *[]schema.TableColumn, row []interface{}) *map[string]interface{} {
	rowMap := make(map[string]interface{})

	nCol := len(*columns)
	if len(row) < nCol {
		nCol = len(row)
	}

	for colId := 0; colId < nCol; colId++ {
		if row[colId] != nil && ((*columns)[colId].RawType == "json" || (*columns)[colId].RawType == "text") {
			rowMap[(*columns)[colId].Name] = string(row[colId].([]uint8))
		} else {
			rowMap[(*columns)[colId].Name] = row[colId]
		}
	}
	return &rowMap
}

func parseColumns(columns *[]schema.TableColumn) *map[string]schema.TableColumn {
	metaMap := make(map[string]schema.TableColumn)

	nCol := len(*columns)

	for colId := 0; colId < nCol; colId++ {
		metaMap[(*columns)[colId].Name] = (*columns)[colId]
	}
	return &metaMap
}

func ParseCanalPayload(e *canal.RowsEvent) *CanalPayload {
	var columnChanged []string
	var payload = &CanalPayload{
		EventType: strings.ToUpper(e.Action),
		Db:        e.Table.Schema,
		Table:     e.Table.Name,
		// MysqlType: make(map[string]string, len(e.Table.Columns)),
	}

	for pk := range e.Table.PKColumns {
		payload.PKColumn = append(payload.PKColumn, e.Table.GetPKColumn(pk).Name)
	}

	// for _, col := range e.Table.Columns {
	// 	payload.MysqlType[col.Name] = col.RawType
	// }

	if e.Action == canal.InsertAction {
		for _, row := range e.Rows {
			payload.Rows = append(payload.Rows, *parseRowMap(&e.Table.Columns, row))
		}
	} else if e.Action == canal.DeleteAction {
		for _, row := range e.Rows {
			payload.Rows = append(payload.Rows, *parseRowMap(&e.Table.Columns, row))
		}

	} else if e.Action == canal.UpdateAction {
		for i := 0; i < len(e.Rows); i += 2 {
			pre := e.Rows[i]
			post := e.Rows[i+1]

			beforeUpdate := *parseRowMap(&e.Table.Columns, pre)
			afterUpdate := *parseRowMap(&e.Table.Columns, post)

			if len(columnChanged) == 0 {
				for col := range afterUpdate {
					if afterUpdate[col] == nil || reflect.TypeOf(afterUpdate[col]).Comparable() {
						if afterUpdate[col] != beforeUpdate[col] {
							columnChanged = append(columnChanged, col)
						}
					} else {
						if !reflect.DeepEqual(afterUpdate[col], beforeUpdate[col]) {
							columnChanged = append(columnChanged, col)
						}
					}
				}
			}

			preUpdate := make(map[string]interface{})
			for _, c := range columnChanged {
				preUpdate[c] = beforeUpdate[c]
			}

			payload.Es = time.Unix(int64(e.Header.Timestamp), 0).UnixNano() / 1e6
			payload.Olds = append(payload.Olds, preUpdate)
			payload.Rows = append(payload.Rows, afterUpdate)
		}
	}

	return payload
}
