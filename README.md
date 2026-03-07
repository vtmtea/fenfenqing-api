# 分分清 API

打牌记分小程序后端 API 服务，使用 Go + Gin + GORM 开发。

## 技术栈

- **Go** 1.21+
- **Gin** - Web 框架
- **GORM** - ORM 库
- **MySQL** - 数据库

## 项目结构

```
fenfenqing-api/
├── cmd/
│   └── server/
│       └── main.go           # 程序入口
├── internal/
│   ├── config/               # 配置
│   ├── handler/              # API 处理器
│   ├── model/                # 数据模型
│   ├── router/               # 路由
│   └── middleware/           # 中间件
├── pkg/
│   └── response/             # 响应工具
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 5.7+

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置环境变量

```bash
# 服务器配置
export SERVER_PORT=8080
export SERVER_MODE=debug

# 数据库配置
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=fenfenqing
```

### 4. 创建数据库

```sql
CREATE DATABASE fenfenqing DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5. 运行服务

```bash
go run cmd/server/main.go
```

服务启动后访问 `http://localhost:8080/health` 检查状态。

## API 接口

### 房间管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/rooms | 获取房间列表 |
| POST | /api/rooms | 创建房间 |
| GET | /api/rooms/:id | 获取房间详情 |
| GET | /api/rooms/roomId/:roomId | 根据房间号获取 |
| DELETE | /api/rooms/:id | 删除房间 |

### 成员管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/rooms/:roomID/members | 获取成员列表 |
| POST | /api/rooms/:roomID/members | 添加成员 |
| DELETE | /api/rooms/:roomID/members/:memberID | 删除成员 |

### 分数管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/rooms/:roomID/scores | 获取分数记录 |
| POST | /api/rooms/:roomID/scores | 添加分数记录 |
| DELETE | /api/rooms/:roomID/scores/:scoreID | 删除分数记录 |

## 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 部署

### Docker 部署（推荐）

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### 构建运行

```bash
# 构建
go build -o fenfenqing-api cmd/server/main.go

# 运行
./fenfenqing-api
```

## License

MIT
