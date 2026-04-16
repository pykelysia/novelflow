# NovelFlow Backend

基于 Gin 框架的 RESTful API 后端服务

## 技术栈

- **Web 框架**: Gin
- **配置管理**: Viper
- **认证**: 双 JWT（访问令牌 + 刷新令牌）
- **数据库**: 外部实现（MySQL）
- **缓存**: 外部实现（Redis）
- **Go 版本**: 1.26

## 项目结构

```
backend/
├── cmd/server/main.go              # 入口文件
├── config/
│   ├── config.go                   # 配置加载逻辑
│   └── config.yaml                 # 配置文件
├── internal/
│   ├── handler/                    # 处理器层
│   │   ├── auth.go                 # 认证处理
│   │   └── user.go                 # 用户 CRUD 处理
│   ├── middleware/                 # 中间件
│   │   ├── auth.go                # JWT 认证中间件
│   │   └── cors.go                # 跨域中间件
│   ├── model/                      # 数据模型
│   │   └── user.go                # 用户模型
│   ├── repository/                # 数据访问层（TODO: 外部实现 MySQL）
│   │   └── user.go                # 用户数据操作
│   ├── service/                   # 业务逻辑层
│   │   ├── auth.go                # 认证业务逻辑
│   │   └── user.go                # 用户业务逻辑
│   └── response/                  # 统一响应格式
│       └── response.go            # 响应结构定义
├── pkg/
│   └── jwt/                       # JWT 工具包
│       └── jwt.go                 # JWT 生成和验证
├── go.mod
└── go.sum
```

## API 路由

### 认证接口（无需登录）

| 方法   | 路径              | 描述           |
|--------|-------------------|----------------|
| POST   | /auth/register    | 用户注册       |
| POST   | /auth/login       | 用户登录       |
| POST   | /auth/refresh     | 刷新令牌       |
| POST   | /auth/logout      | 用户登出       |

### 用户接口（需要登录）

| 方法   | 路径              | 描述           |
|--------|-------------------|----------------|
| GET    | /users            | 获取用户列表   |
| GET    | /users/:id        | 获取单个用户   |
| PUT    | /users/:id        | 更新用户       |
| DELETE | /users/:id        | 删除用户       |

### 其他

| 方法   | 路径              | 描述           |
|--------|-------------------|----------------|
| GET    | /health           | 健康检查       |

## 配置

配置文件位于 `config/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

# TODO: MySQL 配置（外部实现）
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  dbname: "novelflow"

# TODO: Redis 配置（外部实现）
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  access_secret: "your-access-secret-key"
  refresh_secret: "your-refresh-secret-key"
  access_expire: 3600        # 1小时
  refresh_expire: 604800     # 7天
```

## 安装和运行

1. 安装依赖：

```bash
cd backend
go mod tidy
```

2. 配置数据库和 Redis（外部实现）

3. 运行服务：

```bash
go run cmd/server/main.go
```

服务将在 `http://0.0.0.0:8080` 启动。

## TODO 标记清单

- [ ] MySQL 数据库连接和初始化（internal/repository/user.go）
- [ ] Redis 缓存连接和初始化
- [ ] 在 Repository 层实现数据库 CRUD 操作
- [ ] 在 Cache 层实现令牌黑名单（登出功能）
- [ ] 添加数据库迁移脚本
- [ ] 完善单元测试覆盖

## 响应格式

所有 API 响应都使用统一的 JSON 格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

错误响应示例：

```json
{
  "code": 400,
  "message": "bad request"
}
```

## 双 JWT 认证机制

1. **访问令牌（Access Token）**：有效期 1 小时，用于 API 访问验证
2. **刷新令牌（Refresh Token）**：有效期 7 天，用于获取新的访问令牌
3. **流程**：
   - 登录时返回 access_token 和 refresh_token
   - access_token 过期后，使用 refresh_token 换取新令牌
   - refresh_token 过期后需要重新登录
   - 登出时将令牌加入黑名单（TODO: 使用 Redis 存储）

## 许可证

MIT License
