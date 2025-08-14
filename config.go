package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode,omitempty"` // PostgreSQL专用
}

// OutputConfig 输出配置结构体
type OutputConfig struct {
	Directory      string `yaml:"directory"`
	FilenameFormat string `yaml:"filename_format"`
}

// Config 主配置结构体
type Config struct {
	Databases []DatabaseConfig `yaml:"databases"`
	Output    OutputConfig     `yaml:"output"`
}

// LoadConfig 从YAML文件加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// GetDSN 根据数据库类型生成数据源名称
func (db *DatabaseConfig) GetDSN() string {
	switch db.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.Username, db.Password, db.Host, db.Port, db.Database)
	case "postgres":
		sslmode := db.SSLMode
		if sslmode == "" {
			sslmode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			db.Host, db.Port, db.Username, db.Password, db.Database, sslmode)
	default:
		return ""
	}
}
