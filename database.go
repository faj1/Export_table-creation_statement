package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseConnection 数据库连接结构体
type DatabaseConnection struct {
	DB     *sql.DB
	Config *DatabaseConfig
}

// NewDatabaseConnection 创建新的数据库连接
func NewDatabaseConnection(config *DatabaseConfig) (*DatabaseConnection, error) {
	dsn := config.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	db, err := sql.Open(config.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return &DatabaseConnection{
		DB:     db,
		Config: config,
	}, nil
}

// Close 关闭数据库连接
func (dc *DatabaseConnection) Close() error {
	if dc.DB != nil {
		return dc.DB.Close()
	}
	return nil
}

// GetAllTables 获取数据库中所有表名
func (dc *DatabaseConnection) GetAllTables() ([]string, error) {
	var query string
	var tables []string

	switch dc.Config.Type {
	case "mysql":
		query = "SHOW TABLES"
	case "postgres":
		query = `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dc.Config.Type)
	}

	rows, err := dc.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询表名失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, fmt.Errorf("扫描表名失败: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历表名结果失败: %v", err)
	}

	return tables, nil
}

// GetTableDDL 获取指定表的建表语句
func (dc *DatabaseConnection) GetTableDDL(tableName string) (string, error) {
	var query string
	var ddl string

	switch dc.Config.Type {
	case "mysql":
		query = fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
		row := dc.DB.QueryRow(query)
		var tmpTableName string
		err := row.Scan(&tmpTableName, &ddl)
		if err != nil {
			return "", fmt.Errorf("获取MySQL表DDL失败: %v", err)
		}
	case "postgres":
		// PostgreSQL需要通过系统表构建DDL
		ddl, err := dc.getPostgreSQLTableDDL(tableName)
		if err != nil {
			return "", err
		}
		return ddl, nil
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s", dc.Config.Type)
	}

	return ddl, nil
}

// getPostgreSQLTableDDL 获取PostgreSQL表的DDL语句
func (dc *DatabaseConnection) getPostgreSQLTableDDL(tableName string) (string, error) {
	// 获取表结构信息
	query := `
	SELECT 
		column_name,
		data_type,
		is_nullable,
		column_default,
		character_maximum_length,
		numeric_precision,
		numeric_scale
	FROM information_schema.columns 
	WHERE table_name = $1 AND table_schema = 'public'
	ORDER BY ordinal_position`

	rows, err := dc.DB.Query(query, tableName)
	if err != nil {
		return "", fmt.Errorf("查询PostgreSQL表结构失败: %v", err)
	}
	defer rows.Close()

	var ddlParts []string
	ddlParts = append(ddlParts, fmt.Sprintf("CREATE TABLE %s (", tableName))

	var columns []string
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		var charMaxLength, numericPrecision, numericScale sql.NullInt64

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &charMaxLength, &numericPrecision, &numericScale)
		if err != nil {
			return "", fmt.Errorf("扫描PostgreSQL列信息失败: %v", err)
		}

		columnDef := fmt.Sprintf("    %s %s", columnName, dataType)

		// 添加长度/精度信息
		if charMaxLength.Valid {
			columnDef += fmt.Sprintf("(%d)", charMaxLength.Int64)
		} else if numericPrecision.Valid && numericScale.Valid {
			columnDef += fmt.Sprintf("(%d,%d)", numericPrecision.Int64, numericScale.Int64)
		}

		// 添加NOT NULL约束
		if isNullable == "NO" {
			columnDef += " NOT NULL"
		}

		// 添加默认值
		if columnDefault.Valid {
			columnDef += fmt.Sprintf(" DEFAULT %s", columnDefault.String)
		}

		columns = append(columns, columnDef)
	}

	ddlParts = append(ddlParts, strings.Join(columns, ",\n"))
	ddlParts = append(ddlParts, ");")

	return strings.Join(ddlParts, "\n"), nil
}
