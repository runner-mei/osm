package osm

import (
	"fmt"
	"reflect"
)

func resultMaps(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	if ptr, ok := container.(*[]map[string]interface{}); ok {
		return resultMapsNoWrap(o, sql, sqlParams, ptr)
	}

	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Select()() all args must be use ptr"))
	}

	value := reflect.Indirect(pointValue)

	valueNew := []map[string]Data{}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	for rows.Next() {

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

		obj := make(map[string]Data, len(columns))

		for k, v := range fieldValueMap {
			vv := reflect.ValueOf(v).Elem().Interface()
			obj[k] = Data{d: vv}
		}

		valueNew = append(valueNew, obj)

		rowsCount++
	}

	value.Set(reflect.ValueOf(valueNew))

	return rowsCount, nil
}

func resultMapsNoWrap(o *osmBase, sql string, sqlParams []interface{}, container *[]map[string]interface{}) (int64, error) {
	valueNew := []map[string]interface{}{}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	for rows.Next() {
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

		obj := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if values[i] != nil {
				obj[col] = toInternal(values[i])
				fmt.Printf("%T %v\r\n", values[i], values[i])
			}
		}

		valueNew = append(valueNew, obj)
		rowsCount++
	}

	*container = valueNew
	return rowsCount, nil
}

func toInternal(v interface{}) interface{} {
	if value, ok := v.([]byte); ok {
		return string(value)
	}
	return v
}
