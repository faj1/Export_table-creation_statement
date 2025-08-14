package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// 命令行参数
	var configFile = flag.String("config", "config.yaml", "配置文件路径")
	var help = flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 加载配置
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if len(config.Databases) == 0 {
		log.Fatal("配置文件中没有找到数据库配置")
	}

	// 创建用户界面
	ui := NewUserInterface()

	// 让用户选择数据库
	selectedDB, err := ui.SelectDatabase(config.Databases)
	if err != nil {
		log.Fatalf("选择数据库失败: %v", err)
	}

	fmt.Printf("\n正在连接数据库: %s\n", selectedDB.Name)

	// 连接数据库
	dbConn, err := NewDatabaseConnection(selectedDB)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	fmt.Println("数据库连接成功!")

	// 获取所有表
	fmt.Println("正在获取表列表...")
	allTables, err := dbConn.GetAllTables()
	if err != nil {
		log.Fatalf("获取表列表失败: %v", err)
	}

	if len(allTables) == 0 {
		fmt.Println("数据库中没有找到任何表")
		return
	}

	// 让用户选择要导出的表
	selectedTables, err := ui.SelectTablesWithZeroOption(allTables)
	if err != nil {
		log.Fatalf("选择表失败: %v", err)
	}

	// 创建文件操作工具
	fileUtil := NewFileUtil(&config.Output)

	// 导出表DDL
	fmt.Printf("\n开始导出 %d 个表的建表语句到一个文件...\n", len(selectedTables))

	successCount := 0
	tablesDDL := make(map[string]string)

	for i, tableName := range selectedTables {
		ui.ShowProgress(i+1, len(selectedTables), tableName)

		ddl, err := dbConn.GetTableDDL(tableName)
		if err != nil {
			fmt.Printf("错误: 获取表 %s 的DDL失败: %v\n", tableName, err)
			continue
		}

		tablesDDL[tableName] = ddl
		successCount++
	}

	// 保存所有表的DDL到一个文件
	if successCount > 0 {
		err = fileUtil.SaveAllTablesDDL(selectedDB.Database, tablesDDL)
		if err != nil {
			fmt.Printf("错误: 保存DDL文件失败: %v\n", err)
		} else {
			fmt.Printf("成功保存 %d 个表的建表语句\n", successCount)
		}
	}

	// 显示导出摘要
	ui.ShowSummary(successCount, len(selectedTables), config.Output.Directory)
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("数据库表结构导出工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s [选项]\n", os.Args[0])
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -config string")
	fmt.Println("        配置文件路径 (默认: config.yaml)")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("功能:")
	fmt.Println("  - 支持MySQL和PostgreSQL数据库")
	fmt.Println("  - 可配置多个数据库连接")
	fmt.Println("  - 支持导出全部表或指定表的建表语句")
	fmt.Println("  - 自动生成带注释的SQL文件")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s                    # 使用默认配置文件\n", os.Args[0])
	fmt.Printf("  %s -config db.yaml   # 使用指定配置文件\n", os.Args[0])
}
