package internal

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Option struct {
	Alias string
	Force bool
	Like  bool
}

//GetConditionSQL 根据实体获取查询SQL
func GetConditionSQL(struc interface{}, db *sqlx.DB) (condition string, args []interface{}) {
	strucVal := reflect.ValueOf(struc).Elem()
	var count = strucVal.NumField()
	args = make([]interface{}, 0, count)

	for i := 0; i < count; i++ {

		tp := strucVal.Field(i).Type()
		fieldVal := strucVal.Field(i).Interface()
		strucTag := strucVal.Type().Field(i).Tag
		dbTag := strucTag.Get("db")
		options := strucTag.Get("select")
		if dbTag == "" {
			continue
		}
		var force, like bool
		if options != "" {
			for _, opt := range strings.Split(options, " ") {
				switch opt {
				case "force":
					force = true
				case "like": //模糊查询
					like = true
				default:
					dbTag = opt + "." + dbTag
				}
			}
		}
		if !tp.Comparable() {
			continue
		}
		if !force {
			zero := reflect.New(tp).Elem()
			if fieldVal == zero.Interface() {
				continue
			}
		}
		if like {
			condition += " AND  " + dbTag + " like ? "
			switch tp.Kind() {
			case reflect.String:
				args = append(args, "%"+fieldVal.(string)+"%")
			case reflect.Float32, reflect.Float64:
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

			}

			continue
		}

		condition += " AND  " + dbTag + " = ?"
		args = append(args, fieldVal)

	}
	if len(condition) > 4 {
		condition = "where " + condition[4:]
	}

	return condition, args
}

//TableName 获取struct的表名
func TableName(struc interface{}) string {
	var tableName []byte
	tableNameM := reflect.ValueOf(struc).MethodByName("TableName")

	if !tableNameM.IsValid() {
		name := reflect.TypeOf(struc).Elem().Name()
		// 65-90（A-Z），97-122（a-z）
		for i, key := range []byte(name) {
			if key < 90 {
				if i > 1 {
					tableName = append(tableName, byte('_'))
				}
				tableName = append(tableName, key+32)
				continue
			}
			tableName = append(tableName, key)
		}
	} else {
		tableName = tableNameM.Call(nil)[0].Bytes()
	}
	return string(tableName)
}

//InsertBatch 根据实体删除
func InsertBatch(slice interface{}, db *sqlx.DB, tx *sqlx.Tx) (err error) {

	strucPtr := reflect.ValueOf(slice).Index(0)
	strucVal := strucPtr.Elem()
	insertSQL := "insert into " + TableName(strucPtr.Interface())

	var cols string
	var values string
	var count = strucVal.NumField()
	for i := 0; i < count; i++ {

		tp := strucVal.Field(i).Type()
		fieldVal := strucVal.Field(i).Interface()
		strucTag := strucVal.Type().Field(i).Tag

		dbTag := strucTag.Get("db")
		option := strucTag.Get("option")

		if option == "select" || !tp.Comparable() {
			continue
		}

		zero := reflect.New(tp).Elem()
		if tp.Comparable() && fieldVal == zero.Interface() {
			continue
		}
		cols += " , " + dbTag + " "
		values += " , :" + dbTag + " "
	}
	if len(cols) > 2 {
		cols = "(" + cols[2:] + ") "
	}
	if values != "" {
		values = " VALUES (" + values[2:] + " ) "
	}
	insertSQL += cols + values
	if tx != nil {
		_, err = tx.NamedExec(insertSQL, slice)
		return err
	}
	_, err = db.NamedExec(insertSQL, slice)
	return err
}

//Insert 根据实体删除
func Insert(struc interface{}, db *sqlx.DB, tx *sqlx.Tx) (err error) {

	strucPtr := reflect.ValueOf(struc)
	strucVal := strucPtr.Elem()

	insertSQL := "insert into " + TableName(struc)

	var cols string
	var values string
	var count = strucVal.NumField()
	var args = make([]interface{}, 0, count)
	for i := 0; i < count; i++ {

		tp := strucVal.Field(i).Type()
		fieldVal := strucVal.Field(i).Interface()
		strucTag := strucVal.Type().Field(i).Tag

		dbTag := strucTag.Get("db")
		if dbTag == "" {
			continue
		}
		zero := reflect.New(tp).Elem()
		if tp.Comparable() && fieldVal == zero.Interface() {
			continue
		}
		// if fieldVal == zero.Interface() {
		// 	continue
		// }
		cols += " , " + dbTag + " "
		values += ", ?"
		args = append(args, fieldVal)
	}
	if len(cols) > 2 {
		cols = "(" + cols[2:] + ") "
	}
	if len(values) > 2 {
		values = " VALUES (" + values[1:] + " ) "
	}
	insertSQL += cols + values
	if tx != nil {
		if _, err = tx.Exec(insertSQL, args...); err != nil {
			tx.Rollback()
		}

	} else {
		_, err = db.Exec(insertSQL, args...)
	}
	return err
}

//Get 根据实体查询单挑
func Get(struc interface{}, db *sqlx.DB) (err error) {

	strucPtr := reflect.ValueOf(struc)
	strucVal := strucPtr.Elem()

	deleteSQL := "select * from " + TableName(struc)
	var whereSQL string
	var count = strucVal.NumField()
	var args = make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		field := strucVal.Field(i)
		if !field.CanSet() {
			continue
		}

		tp := field.Type()

		dbTag := strucVal.Type().Field(i).Tag.Get("db")

		fieldVal := field.Interface()
		zero := reflect.New(tp).Elem()

		if dbTag == "" || (tp.Comparable() && fieldVal == zero.Interface()) {
			continue
		}
		if (tp.Kind() == reflect.Slice || tp.Kind() == reflect.Array || tp.Kind() == reflect.Map) && field.Len() == 0 {
			continue
		}
		whereSQL += " AND  " + dbTag + " = ?"
		args = append(args, fieldVal)
	}
	if len(whereSQL) > 4 {
		whereSQL = " where " + whereSQL[4:]
	}
	err = db.Get(struc, deleteSQL+whereSQL, args...)
	return err
}

//Delete 根据实体删除
func Delete(struc interface{}, db *sqlx.DB, tx *sqlx.Tx) (result sql.Result, err error) {

	strucPtr := reflect.ValueOf(struc)
	strucVal := strucPtr.Elem()

	deleteSQL := "delete from " + TableName(struc)
	var whereSQL string
	var count = strucVal.NumField()
	var args = make([]interface{}, 0, count)
	for i := 0; i < count; i++ {

		tp := strucVal.Field(i).Type()
		fieldVal := strucVal.Field(i).Interface()
		dbTag := strucVal.Type().Field(i).Tag.Get("db")

		zero := reflect.New(tp).Elem()

		if dbTag == "" || (tp.Comparable() && fieldVal == zero.Interface()) {
			continue
		}
		whereSQL += " AND  " + dbTag + " = ?"
		args = append(args, fieldVal)
	}
	if len(whereSQL) > 4 {
		whereSQL = " where " + whereSQL[4:]
	}

	if tx != nil {
		if result, err = tx.Exec(deleteSQL+whereSQL, args...); err != nil {
			tx.Rollback()
		}
	} else {
		result, err = db.Exec(deleteSQL+whereSQL, args...)
	}

	return result, err
}

//Update 根据实体获取更新
func Update(struc interface{}, db *sqlx.DB, tx *sqlx.Tx, whereCols ...string) error {
	var err error
	var colMp = make(map[string]struct{}, len(whereCols))
	//默认根据ID更新
	if len(whereCols) == 0 {
		colMp["id"] = struct{}{}
	}

	for _, col := range whereCols {
		colMp[col] = struct{}{}
	}
	strucPtr := reflect.ValueOf(struc)
	strucVal := strucPtr.Elem()

	updateSQL := "update " + TableName(struc) + " set "
	var whereSQL, setSQL string
	var ok bool
	var count = strucVal.NumField()
	var setArgs = make([]interface{}, 0, count)
	var whereArgs = make([]interface{}, 0, len(whereCols))
	for i := 0; i < count; i++ {
		tp := strucVal.Field(i).Type()
		fieldVal := strucVal.Field(i).Interface()
		strucTag := strucVal.Type().Field(i).Tag

		dbTag := strucTag.Get("db")
		options := strucTag.Get("update")
		var force = false
		for _, opt := range strings.Split(options, " ") {
			switch opt {
			case "force":
				force = true
			}
		}
		zero := reflect.New(tp).Elem()
		if dbTag == "" || (!force && tp.Comparable() && fieldVal == zero.Interface()) {
			continue
		}

		if _, ok = colMp[dbTag]; ok {
			whereSQL += " AND " + dbTag + " =?"
			whereArgs = append(whereArgs, fieldVal)
			continue

		}

		setSQL += " , " + dbTag + " = ?"
		setArgs = append(setArgs, fieldVal)

	}
	if whereSQL != "" {
		whereSQL = " where " + whereSQL[4:]
	}
	if setSQL == "" { // 无更新字段
		return nil
	}
	updateSQL += setSQL[2:] + whereSQL
	setArgs = append(setArgs, whereArgs...)
	if tx != nil {
		if _, err = tx.Exec(updateSQL, setArgs...); err != nil {
			tx.Rollback()
		}
	} else {
		_, err = db.Exec(updateSQL, setArgs...)
	}

	return err
}

//GetPageRows 分页查询
// func GetPageRows(pageRows *pkg.PageRows, query string, db *sqlx.DB, args ...interface{}) error {
// 	var err error
// 	countSQL := "select count(1) from ( " + query + " ) tmp "
// 	if pageRows.Page.Limit == 0 {
// 		pageRows.Page.Limit = 20
// 	}
// 	if err = db.Get(&pageRows.Page.TotalRows, countSQL, args...); err != nil {
// 		return err
// 	}

// 	if pageRows.Page.Order != "" && pageRows.Page.Sort != "" { //分页排序
// 		query += " Order By " + pageRows.Page.Order + " " + pageRows.Page.Sort
// 	}

// 	args = append(args, pageRows.Page.Offset, pageRows.Page.Limit)
// 	var rowsQuery = query + " limit ?,?"
// 	if err = db.Select(pageRows.Rows, rowsQuery, args...); err != nil {
// 		return err
// 	}
// 	return err
// }

//GetTotalRows 获取分页总数据
func GetTotalRows(query string, db *sqlx.DB) (int64, error) {
	var err error
	countSQL := "select count(1) from ( " + query + " ) tmp "
	var count int64
	if err = db.Get(&count, countSQL); err != nil {
		return count, err
	}
	return count, err
}
