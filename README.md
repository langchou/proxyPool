# ProxyPool

一个基于 Go 的高性能代理池系统，支持自动抓取、验证和提供代理服务。

## 功能特点

- 支持多种代理类型（HTTP/HTTPS/SOCKS4/SOCKS5）
- 自动抓取免费代理
- 定时验证可用性
- RESTful API 接口
- 代理质量评分
- 安全特性
  - 基本认证 (Basic Auth)
  - API Key 认证
  - IP 访问频率限制
  - 自动封禁异常 IP
  - 所有安全特性可配置

## 快速开始

### 环境要求
- Redis 6.0+ (必需)

### 部署步骤

1. 首先需要运行 Redis 服务（以下方式二选一）：

   a. 使用 Docker 运行 Redis（推荐）：
   ```bash
   docker pull redis:latest
   docker run -d --name redis -p 6379:6379 redis:latest
   ```

   b. 直接安装 Redis：
   - Linux: `apt install redis-server` 或 `yum install redis`
   - macOS: `brew install redis`
   - Windows: 从 Redis 官网下载安装包

2. 运行 ProxyPool：
   ```bash
   # 解压下载的发布包
   unzip proxypool-{对应平台}.zip
   cd proxypool-{对应平台}
   
   # 运行服务
   ./proxypool-{对应平台}
   ```

### Redis 配置

在 `data/config.toml` 中配置 Redis 连接信息：

```toml
[redis]
host = "localhost"    # Redis 服务器地址
port = 6379          # Redis 端口
password = ""        # Redis 密码（如果有）
db = 0              # 使用的数据库编号
```

### API 使用

1. 获取单个代理
```bash
curl "http://localhost:8080/proxy"
```

2. 获取指定类型代理
```bash
curl "http://localhost:8080/proxy?type=http,https&count=5"
```

3. 获取高匿代理
```bash
curl "http://localhost:8080/proxy?anonymous=true"
```

### 认证方式

1. 基本认证 (Basic Auth)
```bash
# 使用用户名密码访问
curl -u admin:123456 "http://localhost:8080/proxy"

# 使用 base64 编码方式
curl -H "Authorization: Basic YWRtaW46MTIzNDU2" "http://localhost:8080/proxy"
```

2. API Key 认证
```bash
# 在请求头中携带 API Key
curl -H "X-API-Key: your-api-key" "http://localhost:8080/proxy"
```

### 访问限制
- 每个 IP 在指定时间窗口内有请求次数限制
- 超过限制后 IP 会被临时封禁
- 可以通过响应头查看限制情况：
  - X-RateLimit-Remaining: 剩余请求次数
  - X-RateLimit-Limit: 总请求限制
  - X-RateLimit-Reset: 重置时间

### 配置安全特性

在 `data/config.toml` 中配置：

```toml
[security]
# 基本认证
auth_enabled = false     # 是否启用认证
username = "admin"       # 认证用户名
password = "123456"      # 认证密码

# API Key认证
api_key_enabled = true   # 是否启用 API Key
api_keys = ["key1", "key2", "key3"]  # 允许的 API Key 列表

# 限流配置
rate_limit_enabled = true  # 是否启用限流
rate_limit = 100          # 每个时间窗口最大请求数
rate_window = 1           # 时间窗口（分钟）
ban_duration = 24         # 封禁时长（小时）
```

### 响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "ip": "1.2.3.4",
        "port": "8080",
        "type": "http",
        "anonymous": true,
        "speed_ms": 500,
        "score": 100
    }
}
```

### 配置说明

配置文件位于 `data/config.toml`，主要配置项：
- Redis 连接信息（必需配置）
- 代理验证参数
- 爬虫更新间隔
- 日志配置
- 安全特性配置

## 添加代理源

1. 在 `internal/crawler/sources` 目录下创建新的源文件，例如 `myproxy.go`：

```go
package sources

type MyProxySource struct {
    BaseSource
}

func NewMyProxySource() *MyProxySource {
    return &MyProxySource{
        BaseSource: BaseSource{name: "myproxy"},
    }
}

func (s *MyProxySource) Fetch() ([]*model.Proxy, error) {
    // 实现代理获取逻辑
    proxies := make([]*model.Proxy, 0)
    
    // ... 获取代理的具体实现 ...
    
    return proxies, nil
}
```

2. 在 `internal/crawler/crawler.go` 中注册新代理源：

```go
func NewManager(storage storage.Storage, validator *validator.Validator) *Manager {
    return &Manager{
        sources: []sources.Source{
            sources.NewOpenProxyListSource(),
            sources.NewMyProxySource(),  // 添加新代理源
        },
        storage:   storage,
        validator: validator,
    }
}
```

## 常见问题

1. Redis 连接失败
   - 检查 Redis 服务是否正常运行
   - 确认配置文件中的 Redis 连接信息是否正确
   - 确保 Redis 端口（默认6379）未被占用

2. 配置文件找不到
   - 确保 `data` 目录与程序在同一目录下
   - 确保 `data/config.toml` 文件存在且格式正确

## License

MIT License
