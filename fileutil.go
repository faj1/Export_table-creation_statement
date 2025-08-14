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

// GenerateFileName 生成汇总文件名
func (fu *FileUtil) GenerateFileName(databaseName string) string {
	filename := fu.OutputConfig.FilenameFormat
	filename = strings.ReplaceAll(filename, "{database}", databaseName)

	// 如果没有扩展名，添加.sql
	if !strings.HasSuffix(filename, ".sql") {
		filename += ".sql"
	}

	return filename
}

// SaveAllTablesDDL 保存所有表的DDL到一个文件
func (fu *FileUtil) SaveAllTablesDDL(databaseName string, tablesDDL map[string]string) error {
	err := fu.EnsureOutputDirectory()
	if err != nil {
		return err
	}

	filename := fu.GenerateFileName(databaseName)
	filepath := filepath.Join(fu.OutputConfig.Directory, filename)

	// 按表名排序，保证输出顺序一致
	var tableNames []string
	for tableName := range tablesDDL {
		tableNames = append(tableNames, tableName)
	}

	// 简单排序
	for i := 0; i < len(tableNames)-1; i++ {
		for j := i + 1; j < len(tableNames); j++ {
			if tableNames[i] > tableNames[j] {
				tableNames[i], tableNames[j] = tableNames[j], tableNames[i]
			}
		}
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`-- =============================================
-- 数据库表结构导出文件
-- =============================================
-- 数据库: %s
-- 导出时间: %s
-- 导出表数量: %d
-- 生成工具: export-table-ddl
-- =============================================

`, databaseName, time.Now().Format("2006-01-02 15:04:05"), len(tablesDDL)))

	// 添加目录
	content.WriteString("-- 表目录:\n")
	for i, tableName := range tableNames {
		content.WriteString(fmt.Sprintf("-- %d. %s\n", i+1, tableName))
	}
	content.WriteString("\n")

	// 添加每个表的DDL
	for i, tableName := range tableNames {
		ddl := tablesDDL[tableName]
		content.WriteString(fmt.Sprintf("\n-- ============================================\n"))
		content.WriteString(fmt.Sprintf("-- %d. 表名: %s\n", i+1, tableName))
		content.WriteString("-- ============================================\n\n")
		content.WriteString(ddl)
		content.WriteString("\n\n")
	}

	err = os.WriteFile(filepath, []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("保存DDL文件失败: %v", err)
	}

	fmt.Printf("已保存到文件: %s\n", filepath)
	return nil
}
