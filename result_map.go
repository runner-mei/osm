package osm

import (
	"fmt"
	"reflect"
)

func resultMap(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	if ptr, ok := container.(*map[string]interface{}); ok {
		return resultMapNoWrap(o, sql, sqlParams, ptr)
	}

	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("Select()() all args must be use ptr")
	}

	value := reflect.Indirect(pointValue)

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	if rows.Next() {

		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		fieldValueMap := make(map[string]interface{}, len(columns))

		refs := make([]interface{}, len(columns))
		for i, col := range columns {
			var ref interface{}
			fieldValueMap[toGoName(col)] = &ref
			refs[i] = &ref
		}

		if err := rows.Scan(refs...); err != nil {
			return 0, err
		}

		valueNew := make(map[string]Data, len(refs))
		for k, v := range fieldValueMap {
			vv := reflect.ValueOf(v).Elem().Interface()
			valueNew[k] = Data{d: vv}
		}

		rowsCount++

		value.Set(reflect.ValueOf(valueNew))
	}

	return rowsCount, nil
}

func resultMapNoWrap(o *osmBase, sql string, sqlParams []interface{}, container *map[string]interface{}) (int64, error) {
	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	if rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		values := make([]interface{}, len(columns))
		refs := make([]interface{}, len(columns))
		for i := range columns {
			refs[i] = &values[i]
		}

		if err := rows.Scan(refs...); err != nil {
			return 0, err
		}

		valueNew := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if values[i] != nil {
				valueNew[col] = toInternal(values[i])
			}
		}

		rowsCount++

		*container = valueNew
	}

	return rowsCount, nil
}
