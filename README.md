# NovelFlow

基于 AI Agent 的小说创作平台。

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| AI Agent | CloudWeGo Eino(v0.9.0-alpha.x+) |
| 语言模型 | Claude|
| 数据库 | MySQL (GORM) + MongoDB |
| 缓存 | Redis |
| 认证 | JWT (Access/Refresh Token) |
| 配置 | Viper + YAML + 环境变量 |

## 项目结构

```
novelflow/
├── agents/                  # AI Agent 引擎
│   ├── runner.go           # Agent 运行器（基于 Eino ADK）
│   ├── middleware.go        # 工具调用安全中间件
│   ├── session.go          # MongoDB 会话管理
│   ├── message.go          # 消息模型
│   └── utils.go            # 流式处理工具
├── backend/
│   ├── cmd/server/main.go  # 服务入口
│   └── internal/
│       ├── route.go         # 路由配置
│       ├── handler/         # HTTP 处理器
│       ├── service/         # 业务逻辑层
│       ├── middleware/      # 认证与 CORS 中间件
│       ├── response/        # 统一响应结构
│       └── servicecontext/  # 服务上下文（依赖管理）
├── cache/                   # Redis 客户端（JWT 黑名单）
├── config/                  # Viper 配置加载（支持环境变量覆盖）
├── database/
│   ├── mysql/              # 用户模型与 GORM 仓库
│   └── mongodb/            # MongoDB 连接（Agent 会话存储）
├── frontend/               # 前端（待开发）
└── cmd/                    # 其他命令入口（待开发）
```

## 功能模块

### 用户认证

- 注册 / 登录 / 登出 / 令牌刷新
- JWT Access Token + Refresh Token 双令牌机制
- Redis 令牌黑名单（登出即时失效）

### AI Agent

- 基于 CloudWeGo Eino `deep` 模式的 Agent 运行器
- 流式输出（支持思考过程展示）
- 工具调用安全中间件（错误不影响对话流程）
- 未知工具处理器（模型可自动修正）
- MongoDB 持久化会话管理
- 自动重试（限流 429 重试）

### API 路由

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/auth/register` | 用户注册 | 否 |
| POST | `/auth/login` | 用户登录 | 否 |
| POST | `/auth/refresh` | 刷新令牌 | 否 |
| POST | `/auth/logout` | 用户登出 | 是 |
| GET  | `/users/:id` | 获取用户 | 是 |
| PUT  | `/users/:id` | 更新用户 | 是 |
| DELETE | `/users/:id` | 删除用户 | 是 |
| GET | `/health` | 健康检查 | 否 |

## 快速开始

### 前置条件

- Go 1.26+
- MySQL 8.0+
- Redis 6.0+
- MongoDB 6.0+

### 配置

```bash
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 填入实际的数据库和 API 密钥
```

### 运行

```bash
go run backend/cmd/server/main.go
```

### 健康检查

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```
