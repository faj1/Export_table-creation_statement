package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileUtil 文件操作工具结构体
type FileUtil struct {
	OutputConfig *OutputConfig
}

// NewFileUtil 创建新的文件操作工具
func NewFileUtil(config *OutputConfig) *FileUtil {
	return &FileUtil{
		OutputConfig: config,
	}
}

// EnsureOutputDirectory 确保输出目录存在
func (fu *FileUtil) EnsureOutputDirectory() error {
	if _, err := os.Stat(fu.OutputConfig.Directory); os.IsNotExist(err) {
		err := os.MkdirAll(fu.OutputConfig.Directory, 0755)
		if err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}
		fmt.Printf("已创建输出目录: %s\n", fu.OutputConfig.Directory)
	}
	return nil
}

// GenerateFileName 生成文件名
func (fu *FileUtil) GenerateFileName(databaseName, tableName string) string {
	filename := fu.OutputConfig.FilenameFormat
	filename = strings.ReplaceAll(filename, "{database}", databaseName)
	filename = strings.ReplaceAll(filename, "{table}", tableName)

	// 如果没有扩展名，添加.sql
	if !strings.HasSuffix(filename, ".sql") {
		filename += ".sql"
	}

	return filename
}

// SaveDDL 保存DDL语句到文件
func (fu *FileUtil) SaveDDL(databaseName, tableName, ddl string) error {
	err := fu.EnsureOutputDirectory()
	if err != nil {
		return err
	}

	filename := fu.GenerateFileName(databaseName, tableName)
	filepath := filepath.Join(fu.OutputConfig.Directory, filename)

	// 在DDL前添加注释信息
	content := fmt.Sprintf(`-- 数据库: %s
-- 表名: %s
-- 导出时间: %s
-- 生成工具: export-table-ddl

%s
`, databaseName, tableName, time.Now().Format("2006-01-02 15:04:05"), ddl)

	err = os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("保存DDL文件失败: %v", err)
	}

	fmt.Printf("已保存: %s\n", filepath)
	return nil
}

// SaveAllTablesDDL 保存所有表的DDL到一个文件
func (fu *FileUtil) SaveAllTablesDDL(databaseName string, tablesDDL map[string]string) error {
	err := fu.EnsureOutputDirectory()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_all_tables_ddl.sql", databaseName)
	filepath := filepath.Join(fu.OutputConfig.Directory, filename)

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`-- 数据库: %s
-- 导出时间: %s
-- 生成工具: export-table-ddl
-- 包含所有表的建表语句

`, databaseName, time.Now().Format("2006-01-02 15:04:05")))

	for tableName, ddl := range tablesDDL {
		content.WriteString(fmt.Sprintf("\n-- ============================================\n"))
		content.WriteString(fmt.Sprintf("-- 表名: %s\n", tableName))
		content.WriteString(fmt.Sprintf("-- ============================================\n\n"))
		content.WriteString(ddl)
		content.WriteString("\n\n")
	}

	err = os.WriteFile(filepath, []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("保存所有表DDL文件失败: %v", err)
	}

	fmt.Printf("已保存所有表DDL: %s\n", filepath)
	return nil
}
