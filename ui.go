package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UserInterface 用户交互界面
type UserInterface struct {
	scanner *bufio.Scanner
}

// NewUserInterface 创建新的用户交互界面
func NewUserInterface() *UserInterface {
	return &UserInterface{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// SelectDatabase 让用户选择数据库
func (ui *UserInterface) SelectDatabase(databases []DatabaseConfig) (*DatabaseConfig, error) {
	fmt.Println("\n=== 可用的数据库配置 ===")
	for i, db := range databases {
		fmt.Printf("[%d] %s (%s://%s:%d/%s)\n", i+1, db.Name, db.Type, db.Host, db.Port, db.Database)
	}

	for {
		fmt.Print("\n请选择数据库配置 (输入序号): ")
		if !ui.scanner.Scan() {
			return nil, fmt.Errorf("读取输入失败")
		}

		input := strings.TrimSpace(ui.scanner.Text())
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(databases) {
			fmt.Printf("无效选择，请输入 1-%d 之间的数字\n", len(databases))
			continue
		}

		return &databases[choice-1], nil
	}
}

// SelectTablesWithZeroOption 让用户选择要导出的表，支持输入0选择全部
func (ui *UserInterface) SelectTablesWithZeroOption(allTables []string) ([]string, error) {
	fmt.Printf("\n=== 数据库中的表 (共%d个) ===\n", len(allTables))

	// 显示所有表名，每行显示3个，并带序号
	for i, table := range allTables {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("[%d] %-20s", i+1, table)
	}
	fmt.Println()

	fmt.Println("\n=== 选择要导出的表 ===")
	fmt.Println("输入选项:")
	fmt.Println("  0 - 导出所有表")
	fmt.Println("  表序号 - 导出指定表 (多个表用逗号分隔，如: 1,3,5)")
	fmt.Println("  表名 - 直接输入表名 (多个表用逗号分隔)")

	for {
		fmt.Print("\n请输入选择: ")
		if !ui.scanner.Scan() {
			return nil, fmt.Errorf("读取输入失败")
		}

		input := strings.TrimSpace(ui.scanner.Text())
		if input == "" {
			fmt.Println("输入不能为空，请重新输入")
			continue
		}

		// 如果输入0，返回所有表
		if input == "0" {
			fmt.Printf("已选择导出所有 %d 个表\n", len(allTables))
			return allTables, nil
		}

		// 尝试解析为数字序号
		if selectedTables, err := ui.parseTableNumbers(input, allTables); err == nil {
			return ui.confirmExport(selectedTables)
		}

		// 尝试解析为表名
		if selectedTables, err := ui.parseTableNames(input, allTables); err == nil {
			return ui.confirmExport(selectedTables)
		}

		fmt.Println("无效输入，请重新输入")
		fmt.Println("提示: 输入0选择全部，或输入表序号/表名 (用逗号分隔)")
	}
}

// selectSpecificTables 让用户选择特定的表
func (ui *UserInterface) selectSpecificTables(allTables []string) ([]string, error) {
	fmt.Println("\n请输入要导出的表名，多个表名用逗号分隔:")
	fmt.Println("示例: users,orders,products")
	fmt.Print("\n表名: ")

	if !ui.scanner.Scan() {
		return nil, fmt.Errorf("读取输入失败")
	}

	input := strings.TrimSpace(ui.scanner.Text())
	if input == "" {
		return nil, fmt.Errorf("表名不能为空")
	}

	// 解析输入的表名
	requestedTables := strings.Split(input, ",")
	var validTables []string
	var invalidTables []string

	// 创建表名映射，方便查找
	tableMap := make(map[string]bool)
	for _, table := range allTables {
		tableMap[table] = true
	}

	// 验证输入的表名
	for _, table := range requestedTables {
		table = strings.TrimSpace(table)
		if table == "" {
			continue
		}

		if tableMap[table] {
			validTables = append(validTables, table)
		} else {
			invalidTables = append(invalidTables, table)
		}
	}

	// 显示验证结果
	if len(invalidTables) > 0 {
		fmt.Printf("\n警告: 以下表名不存在: %s\n", strings.Join(invalidTables, ", "))
	}

	if len(validTables) == 0 {
		return nil, fmt.Errorf("没有找到有效的表名")
	}

	fmt.Printf("\n将导出以下表 (%d个): %s\n", len(validTables), strings.Join(validTables, ", "))

	// 确认导出
	return ui.confirmExport(validTables)
}

// confirmExport 确认导出操作
func (ui *UserInterface) confirmExport(tables []string) ([]string, error) {
	fmt.Print("\n确认导出? (y/n): ")
	if !ui.scanner.Scan() {
		return nil, fmt.Errorf("读取输入失败")
	}

	input := strings.ToLower(strings.TrimSpace(ui.scanner.Text()))
	if input == "y" || input == "yes" {
		return tables, nil
	}

	return nil, fmt.Errorf("用户取消导出")
}

// ShowProgress 显示进度信息
func (ui *UserInterface) ShowProgress(current, total int, tableName string) {
	fmt.Printf("正在导出 [%d/%d]: %s\n", current, total, tableName)
}

// parseTableNumbers 解析表序号输入
func (ui *UserInterface) parseTableNumbers(input string, allTables []string) ([]string, error) {
	parts := strings.Split(input, ",")
	var selectedTables []string
	var invalidNumbers []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("无效的数字: %s", part)
		}

		if num < 1 || num > len(allTables) {
			invalidNumbers = append(invalidNumbers, part)
			continue
		}

		tableName := allTables[num-1]
		// 避免重复添加
		found := false
		for _, existing := range selectedTables {
			if existing == tableName {
				found = true
				break
			}
		}
		if !found {
			selectedTables = append(selectedTables, tableName)
		}
	}

	if len(invalidNumbers) > 0 {
		return nil, fmt.Errorf("无效的表序号: %s (有效范围: 1-%d)", strings.Join(invalidNumbers, ", "), len(allTables))
	}

	if len(selectedTables) == 0 {
		return nil, fmt.Errorf("没有选择任何有效的表")
	}

	fmt.Printf("已选择 %d 个表: %s\n", len(selectedTables), strings.Join(selectedTables, ", "))
	return selectedTables, nil
}

// parseTableNames 解析表名输入
func (ui *UserInterface) parseTableNames(input string, allTables []string) ([]string, error) {
	parts := strings.Split(input, ",")
	var selectedTables []string
	var invalidTables []string

	// 创建表名映射，方便查找
	tableMap := make(map[string]bool)
	for _, table := range allTables {
		tableMap[table] = true
	}

	for _, part := range parts {
		tableName := strings.TrimSpace(part)
		if tableName == "" {
			continue
		}

		if tableMap[tableName] {
			// 避免重复添加
			found := false
			for _, existing := range selectedTables {
				if existing == tableName {
					found = true
					break
				}
			}
			if !found {
				selectedTables = append(selectedTables, tableName)
			}
		} else {
			invalidTables = append(invalidTables, tableName)
		}
	}

	if len(invalidTables) > 0 {
		return nil, fmt.Errorf("表不存在: %s", strings.Join(invalidTables, ", "))
	}

	if len(selectedTables) == 0 {
		return nil, fmt.Errorf("没有选择任何有效的表")
	}

	fmt.Printf("已选择 %d 个表: %s\n", len(selectedTables), strings.Join(selectedTables, ", "))
	return selectedTables, nil
}

// ShowSummary 显示导出摘要
func (ui *UserInterface) ShowSummary(successCount, totalCount int, outputDir string) {
	fmt.Println("\n=== 导出完成 ===")
	fmt.Printf("成功导出: %d/%d 个表\n", successCount, totalCount)
	fmt.Printf("输出目录: %s\n", outputDir)

	if successCount < totalCount {
		fmt.Printf("失败: %d 个表导出失败\n", totalCount-successCount)
	}
}
