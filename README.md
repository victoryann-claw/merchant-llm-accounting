# 商户智能记账系统

LLM驱动的商户进销存管理，面向菜市场商贩（卖鱼、卖菜、卖粮油），帮助他们用自然语言记录进货、送货和收款。

## 核心特性

- 🎙️ **自然语言输入** - 说话就能记账，不用手动填表
- 🤖 **LLM智能解析** - AI理解业务语境，自动识别商品、价格、对方
- 📊 **智能报表** - 每日/每周/每月自动汇总账单
- 🔔 **送货提醒** - 到点自动提醒，不漏单
- 🔒 **数据安全** - 商户数据完全隔离

## 技术栈

- **后端**: Golang
- **数据库**: PostgreSQL + JSONB
- **LLM**: 支持通义千问、智谱GLM、豆包、MiniMax 等国内模型
- **前端**: 微信小程序

## 快速开始

### 前置依赖

- Go 1.21+
- PostgreSQL 14+

### 环境变量

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=merchant_accounting
export QWEN_API_KEY=your-api-key  # 通义千问API Key
export API_PORT=8080
```

### 数据库初始化

```bash
createdb merchant_accounting
```

### 启动服务

```bash
go mod download
go run cmd/server/main.go
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `POST /api/v1/merchant` | 创建商户 |
| `POST /api/v1/record` | 提交记录（LLM解析） |
| `GET /api/v1/records` | 查询记录列表 |
| `GET /api/v1/records/:id` | 获取单条记录 |
| `GET /health` | 健康检查 |

## 项目结构

```
merchant-llm-accounting/
├── cmd/server/          # 入口
├── internal/
│   ├── api/            # HTTP层
│   ├── llm/            # LLM适配器
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据访问
│   └── model/          # 数据模型
├── pkg/db/             # 数据库连接
└── migrations/         # 数据库迁移
```

## License

MIT
