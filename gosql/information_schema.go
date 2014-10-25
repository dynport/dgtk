package gosql

import (
	"fmt"
	"reflect"
	"strings"
)

func Columns(db Dbi) ([]*Column, error) {
	names, err := columnNames(db, &Column{})
	if err != nil {
		return nil, err
	}
	q := "SELECT " + strings.Join(names, ", ") + " FROM information_schema.columns"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		t := &Column{}
		rows.Scan(&t.TableCatalog, &t.TableSchema, &t.TableName, &t.ColumnName, &t.OrdinalPosition, &t.ColumnDefault, &t.IsNullable, &t.DataType, &t.CharacterMaximumLength, &t.CharacterOctetLength, &t.NumericPrecision, &t.NumericPrecisionRadix, &t.NumericScale, &t.DatetimePrecision, &t.IntervalType, &t.IntervalPrecision, &t.CharacterSetCatalog, &t.CharacterSetSchema, &t.CharacterSetName, &t.CollationCatalog, &t.CollationSchema, &t.CollationName, &t.DomainCatalog, &t.DomainSchema, &t.DomainName, &t.UdtCatalog, &t.UdtSchema, &t.UdtName, &t.ScopeCatalog, &t.ScopeSchema, &t.ScopeName, &t.MaximumCardinality, &t.DtdIdentifier, &t.IsSelfReferencing, &t.IsIdentity, &t.IdentityGeneration, &t.IdentityStart, &t.IdentityIncrement, &t.IdentityMaximum, &t.IdentityMinimum, &t.IdentityCycle, &t.IsGenerated, &t.GenerationExpression, &t.IsUpdatable)
		err = rows.Scan()
		if err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

type tableOptions struct {
	TableSchema    string
	TableName      string
	IncludeColumns bool
}

type tableOpt func(*tableOptions)

func (t *tableOptions) apply(funcs ...tableOpt) {
	for _, f := range funcs {
		f(t)
	}
}

func WithName(s string) tableOpt {
	return func(o *tableOptions) {
		o.TableName = s
	}
}

var PublicTables = func(o *tableOptions) {
	o.TableSchema = "public"
}

func Tables(db Dbi, funcs ...tableOpt) ([]*Table, error) {
	opt := &tableOptions{}
	opt.apply(funcs...)

	names, err := columnNames(db, &Table{})
	if err != nil {
		return nil, err
	}
	q := "SELECT " + strings.Join(names, ", ") + " FROM information_schema.tables"
	v := []interface{}{}
	w := []string{}
	if opt.TableSchema != "" {
		w = append(w, "table_schema = $1")
		v = append(v, opt.TableSchema)
	}
	if opt.TableName != "" {
		w = append(w, fmt.Sprintf("table_name = $%d", len(w)+1))
		v = append(v, opt.TableName)
	}
	if len(w) > 0 {
		q += " WHERE " + strings.Join(w, " AND ")
	}
	rows, err := db.Query(q, v...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Table{}
	return out, UnmarshalRows(rows, &out)
}

func columnNames(db Dbi, i interface{}) ([]string, error) {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	out := []string{}
	for i := 0; i < t.NumField(); i++ {
		column := t.Field(i).Tag.Get("sql")
		if column != "" {
			out = append(out, column)
		}
	}
	return out, nil
}

type Table struct {
	TableCatalog              string  `sql:"table_catalog"`
	TableSchema               string  `sql:"table_schema"`
	TableName                 string  `sql:"table_name"`
	TableType                 string  `sql:"table_type"`
	SelfReferencingColumnName *string `sql:"self_referencing_column_name"`
	ReferenceGeneration       *string `sql:"reference_generation"`
	UserDefinedTypeCatalog    *string `sql:"user_defined_type_catalog"`
	UserDefinedTypeSchema     *string `sql:"user_defined_type_schema"`
	UserDefinedTypeName       *string `sql:"user_defined_type_name"`
	IsInsertableInto          string  `sql:"is_insertable_into"`
	IsTyped                   string  `sql:"is_typed"`
	CommitAction              *string `sql:"commit_action"`
}

type Column struct {
	TableCatalog           string `sql:"table_catalog"`            // information_schema.sql_identifier
	TableSchema            string `sql:"table_schema"`             // information_schema.sql_identifier
	TableName              string `sql:"table_name"`               // information_schema.sql_identifier
	ColumnName             string `sql:"column_name"`              // information_schema.sql_identifier
	OrdinalPosition        int    `sql:"ordinal_position"`         // information_schema.cardinal_number
	ColumnDefault          string `sql:"column_default"`           // information_schema.character_data
	IsNullable             bool   `sql:"is_nullable"`              // information_schema.yes_or_no
	DataType               string `sql:"data_type"`                // information_schema.character_data
	CharacterMaximumLength int    `sql:"character_maximum_length"` // information_schema.cardinal_number
	CharacterOctetLength   int    `sql:"character_octet_length"`   // information_schema.cardinal_number
	NumericPrecision       int    `sql:"numeric_precision"`        // information_schema.cardinal_number
	NumericPrecisionRadix  int    `sql:"numeric_precision_radix"`  // information_schema.cardinal_number
	NumericScale           int    `sql:"numeric_scale"`            // information_schema.cardinal_number
	DatetimePrecision      int    `sql:"datetime_precision"`       // information_schema.cardinal_number
	IntervalType           string `sql:"interval_type"`            // information_schema.character_data
	IntervalPrecision      int    `sql:"interval_precision"`       // information_schema.cardinal_number
	CharacterSetCatalog    string `sql:"character_set_catalog"`    // information_schema.sql_identifier
	CharacterSetSchema     string `sql:"character_set_schema"`     // information_schema.sql_identifier
	CharacterSetName       string `sql:"character_set_name"`       // information_schema.sql_identifier
	CollationCatalog       string `sql:"collation_catalog"`        // information_schema.sql_identifier
	CollationSchema        string `sql:"collation_schema"`         // information_schema.sql_identifier
	CollationName          string `sql:"collation_name"`           // information_schema.sql_identifier
	DomainCatalog          string `sql:"domain_catalog"`           // information_schema.sql_identifier
	DomainSchema           string `sql:"domain_schema"`            // information_schema.sql_identifier
	DomainName             string `sql:"domain_name"`              // information_schema.sql_identifier
	UdtCatalog             string `sql:"udt_catalog"`              // information_schema.sql_identifier
	UdtSchema              string `sql:"udt_schema"`               // information_schema.sql_identifier
	UdtName                string `sql:"udt_name"`                 // information_schema.sql_identifier
	ScopeCatalog           string `sql:"scope_catalog"`            // information_schema.sql_identifier
	ScopeSchema            string `sql:"scope_schema"`             // information_schema.sql_identifier
	ScopeName              string `sql:"scope_name"`               // information_schema.sql_identifier
	MaximumCardinality     int    `sql:"maximum_cardinality"`      // information_schema.cardinal_number
	DtdIdentifier          string `sql:"dtd_identifier"`           // information_schema.sql_identifier
	IsSelfReferencing      bool   `sql:"is_self_referencing"`      // information_schema.yes_or_no
	IsIdentity             bool   `sql:"is_identity"`              // information_schema.yes_or_no
	IdentityGeneration     string `sql:"identity_generation"`      // information_schema.character_data
	IdentityStart          string `sql:"identity_start"`           // information_schema.character_data
	IdentityIncrement      string `sql:"identity_increment"`       // information_schema.character_data
	IdentityMaximum        string `sql:"identity_maximum"`         // information_schema.character_data
	IdentityMinimum        string `sql:"identity_minimum"`         // information_schema.character_data
	IdentityCycle          bool   `sql:"identity_cycle"`           // information_schema.yes_or_no
	IsGenerated            string `sql:"is_generated"`             // information_schema.character_data
	GenerationExpression   string `sql:"generation_expression"`    // information_schema.character_data
	IsUpdatable            bool   `sql:"is_updatable"`             // information_schema.yes_or_no
}
