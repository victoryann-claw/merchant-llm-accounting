# 商户智能记账系统

LLM驱动的商户进销存管理，面向菜市场商贩（卖鱼、卖菜、卖粮油），帮助他们用自然语言记录进货、送货和收款。

## 核心特性

- 🎙️ **语音输入** - 按住说话，腾讯云ASR识别
- 📷 **图片识别** - 拍照识别进货单/送货单
- ✍️ **文字输入** - 直接打字描述交易
- 🤖 **LLM智能解析** - AI理解业务语境，自动识别商品、价格、对方
- 📊 **智能报表** - 每日/每周/每月自动汇总账单
- 🔔 **送货提醒** - 到点自动提醒，不漏单
- 🔒 **数据安全** - 商户数据完全隔离

## 技术栈

- **后端**: Golang 1.21+
- **数据库**: PostgreSQL 14+ / JSONB
- **LLM**: 支持通义千问、智谱GLM、豆包、MiniMax 等国内模型
- **语音**: 腾讯云 ASR
- **前端**: 微信小程序

## 快速开始

### Docker 部署（推荐）

```bash
# 克隆代码
git clone https://github.com/victoryann-claw/merchant-llm-accounting.git
cd merchant-llm-accounting

# 配置环境变量
cp .env.example .env
# 编辑 .env 填入你的 API Key

# 一键启动
docker-compose up -d
```

### 访问服务

```bash
curl http://localhost:8080/health
# 返回 {"status": "ok"}
```

### 开发模式

```bash
# 1. 安装 PostgreSQL
# 2. 创建数据库
createdb merchant_accounting

# 3. 配置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=merchant_accounting
export QWEN_API_KEY=你的Key

# 4. 运行
go run cmd/server/main.go
```

## 项目结构

```
merchant-llm-accounting/
├── cmd/server/          # 程序入口
├── internal/
│   ├── api/            # HTTP层（handler/middleware/router）
│   ├── llm/            # LLM适配器（通义千问/腾讯ASR等）
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据访问
│   └── model/          # 数据模型
├── pkg/db/             # 数据库连接
├── docs/               # 文档
│   ├── API文档.md
│   └── 部署文档.md
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `POST /api/v1/merchant` | 创建商户 |
| `POST /api/v1/record` | 文字提交记录（LLM解析） |
| `POST /api/v1/record/voice` | 语音提交记录 |
| `POST /api/v1/record/image` | 图片提交记录 |
| `GET /api/v1/records` | 查询记录列表 |
| `GET /api/v1/stats/today` | 今日统计 |
| `GET /api/v1/report/daily` | 日报 |
| `GET /api/v1/report/periodic` | 周报/月报 |
| `POST /api/v1/reminders` | 创建提醒 |
| `GET /health` | 健康检查 |

详见 [API文档](docs/API文档.md)

## 环境变量

| 变量 | 必填 | 说明 |
|------|------|------|
| `DB_HOST` | 是 | 数据库地址 |
| `DB_PORT` | 是 | 数据库端口 |
| `DB_USER` | 是 | 数据库用户 |
| `DB_PASSWORD` | 是 | 数据库密码 |
| `DB_NAME` | 是 | 数据库名 |
| `API_PORT` | 否 | API端口，默认8080 |
| `QWEN_API_KEY` | 是 | 通义千问API Key |
| `QWEN_VL_API_KEY` | 否 | 通义千问VL Key（图片识别） |
| `TENCENT_ASR_SECRET_ID` | 否 | 腾讯云SecretId（语音识别） |
| `TENCENT_ASR_SECRET_KEY` | 否 | 腾讯云SecretKey（语音识别） |

## 部署

详见 [部署文档](docs/部署文档.md)

## License

MIT
