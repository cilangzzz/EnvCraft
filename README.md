# EnvCraft

**EnvCraft** 是一款高效、轻量化的配置文件迁移工具，旨在帮助开发者和运维人员快速迁移和同步软件配置、环境变量及注册表项。

## 项目简介

EnvCraft 提供了一套完整的配置迁移解决方案，支持多种配置类型的导出、导入和迁移操作。用户可以轻松地将开发环境的配置文件打包，在新环境中快速恢复，大幅减少手动配置时间。

### 核心功能

- **配置文件迁移**: 支持 JSON、YAML、XML、INI、TOML 等多种配置文件格式
- **环境变量迁移**: 支持 Windows 用户变量和系统变量的迁移
- **软件配置迁移**: 支持软件配置目录、数据文件的批量迁移
- **注册表迁移**: 支持 Windows 注册表项的导出和导入（仅 Windows）
- **预览模式**: 执行前可预览变更，评估影响
- **回滚机制**: 支持操作回滚，保障数据安全

## 项目结构

```
EnvCraft/
├── go-server/                    # Go 服务端代码
│   ├── cmd/                      # 应用入口
│   │   └── application/          # 主程序入口
│   ├── internal/                 # 内部模块
│   │   ├── backend_service/      # 后端服务
│   │   │   ├── handler/          # HTTP 处理器
│   │   │   ├── model/            # 数据模型
│   │   │   └── router/           # 路由配置
│   │   └── cfg/                  # 配置管理
│   └── pkg/                      # 公共包
│       ├── common/               # 通用工具
│       ├── constants/            # 常量定义
│       └── util/                 # 工具库
│           ├── migration/        # 迁移核心模块
│           │   ├── core/         # 核心接口和实现
│           │   │   └── strategies/  # 迁移策略实现
│           │   └── constants/    # 迁移相关常量
│           ├── downloader/       # 下载器模块
│           └── server_command/   # 服务器命令模块
└── README.md
```

## 快速开始

### 环境要求

- Go 1.22.6 或更高版本
- Windows 操作系统（注册表迁移功能需要）

### 安装

```bash
# 克隆仓库
git clone https://github.com/your-username/EnvCraft.git
cd EnvCraft

# 进入服务端目录
cd go-server

# 安装依赖
go mod download

# 编译
go build -o envcraft ./cmd/application
```

### 启动服务

```bash
# 默认启动
./envcraft

# 自定义参数启动
./envcraft -ip 0.0.0.0 -port 8080 -key your-secret-key -debug
```

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-ip` | 0.0.0.0 | 服务器监听 IP |
| `-port` | 8080 | 服务器监听端口 |
| `-key` | default-secret-key | 安全密钥 |
| `-debug` | false | 是否开启调试模式 |

## API 文档

服务启动后，可通过以下地址访问 Swagger API 文档：

```
http://localhost:8080/swagger/index.html
```

### 主要 API 接口

#### 迁移管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/migration/execute` | 执行迁移任务 |
| POST | `/api/v1/migration/dry-run` | 预览迁移任务 |
| POST | `/api/v1/migration/rollback` | 回滚迁移任务 |
| GET | `/api/v1/migration/tasks` | 获取任务列表 |
| GET | `/api/v1/migration/tasks/:task_id` | 获取任务详情 |
| GET | `/api/v1/migration/strategies` | 获取可用策略列表 |

#### 导入导出

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/migration/export` | 导出配置文件 |
| POST | `/api/v1/migration/import` | 导入配置文件 |

## 迁移类型

### 1. 环境变量迁移 (env_variable)

支持迁移 Windows 环境变量，包括用户变量和系统变量。

```json
{
  "name": "环境变量迁移",
  "type": "env_variable",
  "source": {
    "type": "user",
    "variables": {
      "JAVA_HOME": "C:\\Program Files\\Java\\jdk-17",
      "NODE_PATH": "C:\\Program Files\\nodejs"
    }
  },
  "target": {
    "type": "user"
  }
}
```

### 2. 配置文件迁移 (config_file)

支持多种配置文件格式的迁移，包括 JSON、YAML、XML、INI、TOML。

```json
{
  "name": "配置文件迁移",
  "type": "config_file",
  "source": {
    "type": "local",
    "path": "C:\\Users\\...\\.idea\\config.xml",
    "format": "xml"
  },
  "target": {
    "type": "local",
    "path": "D:\\backup\\idea_config.xml",
    "backup": true
  }
}
```

### 3. 软件配置迁移 (software)

支持软件配置目录、数据文件的批量迁移。

```json
{
  "name": "软件配置迁移",
  "type": "software",
  "source": {
    "path": "C:\\Users\\...\\.vscode"
  },
  "target": {
    "path": "D:\\backup\\.vscode",
    "backup": true,
    "merge_mode": "overwrite"
  }
}
```

### 4. 注册表迁移 (registry)

支持 Windows 注册表项的导出和导入（仅 Windows）。

```json
{
  "name": "注册表迁移",
  "type": "registry",
  "source": {
    "path": "HKCU\\Software\\MyApp"
  },
  "target": {
    "path": "D:\\backup\\myapp.reg"
  }
}
```

## 导出包格式

导出的配置文件采用标准 JSON 格式，便于分享和版本控制：

```json
{
  "metadata": {
    "version": "1.0",
    "export_id": "export_123",
    "export_time": "2024-01-01T12:00:00Z",
    "source_type": "config_file",
    "original_format": "xml",
    "original_path": "C:\\Users\\...\\.idea\\config.xml",
    "checksum": "sha256:abc123...",
    "tags": ["ide", "java"],
    "description": "IDEA 配置文件导出",
    "app_info": {
      "name": "IntelliJ IDEA",
      "version": "2024.1",
      "category": "IDE"
    }
  },
  "content": {
    "data": {
      "key": "value"
    },
    "raw_content": "base64...",
    "format_specific_data": {}
  }
}
```

## 开发指南

### 添加新的迁移策略

1. 在 `pkg/util/migration/core/strategies/` 目录下创建新的策略文件
2. 实现 `MigrationStrategy` 接口：

```go
type MigrationStrategy interface {
    Name() string
    Type() MigrationType
    Description() string
    Validate(config *MigrationConfig) error
    Execute(ctx context.Context, config *MigrationConfig) (*MigrationResult, error)
    Rollback(ctx context.Context, config *MigrationConfig) error
    DryRun(ctx context.Context, config *MigrationConfig) (*MigrationPreview, error)
    Export(ctx context.Context, config *MigrationConfig) (*ExportResult, error)
    Import(ctx context.Context, config *MigrationConfig) (*ImportResult, error)
    ValidateExport(config *MigrationConfig) error
    ValidateImport(config *MigrationConfig) error
}
```

3. 在 `init()` 函数中注册策略：

```go
func init() {
    if err := core.RegisterStrategy(&MyStrategy{}); err != nil {
        panic(fmt.Sprintf("failed to register my strategy: %v", err))
    }
}
```

### 项目依赖

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Swagger](https://github.com/swaggo/swag) - API 文档
- [Zap](https://go.uber.org/zap) - 日志库

## 未来规划

### 服务器平台（计划中）

我们将开发一个在线配置分享平台，主要功能包括：

- **配置共享**: 用户可以上传自己的配置文件，分享给其他开发者
- **快速下载**: 其他用户可以快速浏览和下载需要的配置模板
- **版本管理**: 支持配置文件的版本控制和更新追踪
- **分类标签**: 按软件类型、用途等分类管理配置
- **社区评价**: 用户可对配置进行评分和评论

### 其他规划

- [ ] 支持 macOS 和 Linux 环境配置迁移
- [ ] 增加配置文件差异对比功能
- [ ] 支持加密存储敏感配置
- [ ] 提供命令行工具 (CLI)
- [ ] 开发桌面客户端

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 联系方式

如有问题或建议，欢迎提交 Issue。