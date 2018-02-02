package migrator

import "fmt"

type Sqlite3 struct {
	BaseDialect
}

func NewSqlite3Dialect() *Sqlite3 {
	d := Sqlite3{}
	d.BaseDialect.dialect = &d
	d.BaseDialect.driverName = SQLITE
	return &d
}

func (db *Sqlite3) SupportEngine() bool {
	return false
}

func (db *Sqlite3) Quote(name string) string {
	return "`" + name + "`"
}

func (db *Sqlite3) QuoteStr() string {
	return "`"
}

func (db *Sqlite3) AutoIncrStr() string {
	return "AUTOINCREMENT"
}

func (db *Sqlite3) SqlType(c *Column) string {
	switch c.Type {
	case DB_Date, DB_DateTime, DB_TimeStamp, DB_Time:
		return DB_DateTime
	case DB_TimeStampz:
		return DB_Text
	case DB_Char, DB_Varchar, DB_NVarchar, DB_TinyText, DB_Text, DB_MediumText, DB_LongText:
		return DB_Text
	case DB_Bit, DB_TinyInt, DB_SmallInt, DB_MediumInt, DB_Int, DB_Integer, DB_BigInt, DB_Bool:
		return DB_Integer
	case DB_Float, DB_Double, DB_Real:
		return DB_Real
	case DB_Decimal, DB_Numeric:
		return DB_Numeric
	case DB_TinyBlob, DB_Blob, DB_MediumBlob, DB_LongBlob, DB_Bytea, DB_Binary, DB_VarBinary:
		return DB_Blob
	case DB_Serial, DB_BigSerial:
		c.IsPrimaryKey = true
		c.IsAutoIncrement = true
		c.Nullable = false
		return DB_Integer
	default:
		return c.Type
	}
}

func (db *Sqlite3) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}

func (db *Sqlite3) DropIndexSql(tableName string, index *Index) string {
	quote := db.Quote
	//var unique string
	idxName := index.XName(tableName)
	return fmt.Sprintf("DROP INDEX %v", quote(idxName))
}
