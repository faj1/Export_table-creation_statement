# 数据库表结构导出工具

这是一个用Go语言编写的数据库表结构导出工具，可以读取数据库的所有表或者指定表的建表语句并保存到本地文件。

## 功能特性

- 🚀 支持MySQL和PostgreSQL数据库
- 📝 支持多个数据库配置
- 🎯 支持导出全部表或指定表
- 📁 自动生成带注释的SQL文件
- 🔢 灵活的表选择方式（序号、表名、全选）
- 📊 详细的导出进度显示

## 安装和使用

### 1. 环境要求

- Go 1.21 或更高版本
- MySQL 或 PostgreSQL 数据库

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

编辑 `config.yaml` 文件，配置你的数据库连接信息：

```yaml
databases:
  - name: "MySQL本地数据库"
    type: "mysql"
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "test"
    
  - name: "PostgreSQL本地数据库"
    type: "postgres"
    host: "localhost"
    port: 5432
    username: "postgres"
    password: "password"
    database: "test"
    sslmode: "disable"

output:
  directory: "./output"
  filename_format: "{database}_{table}_ddl.sql"
```

### 4. 编译程序

```bash
go build -o export-ddl
```

### 5. 运行程序

```bash
./export-ddl
```

或者指定配置文件：

```bash
./export-ddl -config my-config.yaml
```

### 6. 使用说明

1. **选择数据库**: 程序会显示配置文件中的所有数据库，输入序号选择
2. **选择表**: 程序连接数据库后会显示所有表，你可以：
   - 输入 `0` - 导出所有表
   - 输入表序号 - 如 `1,3,5` 导出指定序号的表
   - 输入表名 - 如 `users,orders` 导出指定名称的表

## 输出文件

程序会在输出目录中生成一个汇总文件：

**汇总文件**: `{数据库名}_tables_ddl.sql` - 包含所有选中表的建表语句

文件包含：
- 数据库基本信息和导出时间
- 表的目录索引
- 按表名排序的完整建表语句
- 清晰的分隔符和注释

## 命令行选项

```bash
./export-ddl [选项]

选项:
  -config string
        配置文件路径 (默认: config.yaml)
  -help
        显示帮助信息
```

## 支持的数据库

### MySQL
- 使用 `SHOW CREATE TABLE` 获取完整的DDL语句
- 支持所有MySQL表结构特性

### PostgreSQL
- 通过系统表构建DDL语句
- 包含列定义、数据类型、约束等信息

## 示例

### 配置示例

```yaml
databases:
  - name: "生产环境MySQL"
    type: "mysql"
    host: "prod.example.com"
    port: 3306
    username: "readonly_user"
    password: "secure_password"
    database: "production"

output:
  directory: "./backup/ddl"
  filename_format: "prod_tables_ddl.sql"
```

### 运行示例

```bash
$ ./export-ddl

=== 可用的数据库配置 ===
[1] MySQL本地数据库 (mysql://localhost:3306/test)
[2] PostgreSQL本地数据库 (postgres://localhost:5432/test)

请选择数据库配置 (输入序号): 1

正在连接数据库: MySQL本地数据库
数据库连接成功!
正在获取表列表...

=== 数据库中的表 (共5个) ===
[1] users              [2] orders             [3] products           
[4] categories         [5] order_items        

=== 选择要导出的表 ===
输入选项:
  0 - 导出所有表
  表序号 - 导出指定表 (多个表用逗号分隔，如: 1,3,5)
  表名 - 直接输入表名 (多个表用逗号分隔)

请输入选择: 0
已选择导出所有 5 个表

开始导出 5 个表的建表语句到一个文件...
正在导出 [1/5]: users
正在导出 [2/5]: orders
正在导出 [3/5]: products
正在导出 [4/5]: categories
正在导出 [5/5]: order_items
已保存到文件: ./output/test_tables_ddl.sql
成功保存 5 个表的建表语句

=== 导出完成 ===
成功导出: 5/5 个表
输出目录: ./output
```

## 注意事项

1. 确保数据库用户有足够的权限读取表结构
2. PostgreSQL用户需要对 `information_schema` 有读取权限
3. 大型数据库可能需要较长时间来获取所有表的DDL
4. 建议在生产环境使用只读账户
