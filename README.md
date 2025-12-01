# SigmaFlow - AI 量化交易机器人

SigmaFlow 是一个基于 AI 的量化交易机器人，使用 Gemini AI 进行市场分析，并通过 OKX 交易所执行交易。

## 功能特性

- 🤖 **AI 市场分析** - 使用 Gemini AI 进行技术分析和交易决策
- 📊 **技术指标** - MA, EMA, RSI, MACD, 布林带, ATR, 随机指标等
- 💹 **OKX 交易** - 支持 OKX 交易所现货交易
- ⏰ **定时调度** - 支持定时执行交易策略
- 🛡️ **风险控制** - 仓位管理、止损止盈、每日亏损限制
- 📢 **交易通知** - 支持 Telegram 通知
- 📈 **回测系统** - 策略回测功能
- 📝 **结构化日志** - JSON/Text 格式日志

## 快速开始

### 1. 安装

```bash
git clone https://github.com/TwoOnefour/SigmaFlow.git
cd SigmaFlow
go mod download
```

### 2. 配置

复制配置示例文件并填入你的 API 密钥：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置以下内容：
- OKX API 密钥
- Gemini API 密钥
- 交易参数
- 风险控制参数

### 3. 运行

单次执行模式：
```bash
go run cmd/main.go
```

调度器模式（需在 .env 中设置 `SCHEDULER_ENABLED=1`）：
```bash
go run cmd/main.go
```

## 项目结构

```
.
├── cmd/
│   └── main.go              # 程序入口
├── internal/
│   ├── backtest/            # 回测系统
│   │   └── engine.go        # 回测引擎
│   ├── client/
│   │   ├── exchange/
│   │   │   └── okx/         # OKX 交易所客户端
│   │   └── llm/
│   │       └── gemini/      # Gemini AI 客户端
│   ├── config/              # 配置管理
│   ├── logger/              # 日志模块
│   ├── model/               # 数据模型
│   ├── notify/              # 通知系统
│   ├── risk/                # 风险控制
│   ├── scheduler/           # 任务调度器
│   ├── service/
│   │   ├── llm/             # LLM 服务
│   │   └── trade/           # 交易服务
│   └── util/                # 工具函数（技术指标）
└── pkg/
    └── currency/            # 货币定义
```

## 技术指标

- **MA (移动平均线)** - 简单移动平均
- **EMA (指数移动平均)** - 指数加权移动平均
- **RSI (相对强弱指数)** - 衡量超买超卖
- **MACD** - 移动平均聚散指标
- **布林带** - 波动率通道
- **ATR (平均真实波幅)** - 波动率指标
- **随机指标 (KD)** - 动量指标

## 风险控制

- **最大仓位控制** - 限制单次交易的最大仓位比例
- **每日亏损限制** - 达到限制后停止交易
- **止损/止盈** - 自动止损止盈检测
- **仓位计算** - 基于风险的仓位大小计算

## 回测

使用内置回测引擎测试策略：

```go
import "okx/internal/backtest"

strategy := backtest.NewSimpleMAStrategy(5, 20)
engine := backtest.NewEngine(strategy, 10000, 0.001)
result, _ := engine.Run(candles, pair)
fmt.Println(result.Report())
```

## 开发

运行测试：
```bash
go test ./...
```

构建：
```bash
go build ./cmd/main.go
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `OKX_SIMULATE` | 模拟交易模式 | 0 |
| `GEMINI_API_KEY` | Gemini API 密钥 | - |
| `SCHEDULER_ENABLED` | 启用调度器 | 0 |
| `LOG_LEVEL` | 日志级别 | info |
| `RISK_MAX_POSITION_PCT` | 最大仓位比例 | 0.5 |

更多配置请参考 `.env.example`

## 免责声明

⚠️ **警告**: 加密货币交易具有高风险。本软件仅供学习和研究目的，不构成投资建议。使用本软件进行交易的风险由用户自行承担。

## License

MIT License
