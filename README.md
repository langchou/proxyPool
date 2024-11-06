# ProxyPool

一个基于 Go 的高性能代理池系统，支持自动抓取、验证和提供代理服务。

## 功能特点

- 支持多种代理类型（HTTP/HTTPS/SOCKS4/SOCKS5）
- 自动抓取免费代理
- 定时验证可用性
- RESTful API 接口
- 代理质量评分

## 快速开始

### 环境要求
- Go 1.21+
- Redis 6.0+

### 安装运行

1. 克隆项目
```bash
git clone https://github.com/your-repo/proxy-pool.git
```

2. 编译运行
```bash
make build

make run
```


### API使用

1. 获取单个代理
``` bash
curl http://localhost:8080/proxy
```

2. 获取指定类型代理
``` bash
curl "http://localhost:8080/proxy?type=http,https&count=5"
```

3. 获取高匿代理
``` bash
curl "http://localhost:8080/proxy?anonymous=true"
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

## 配置说明

配置文件位于 `data/config.toml`，主要配置项：
- Redis 连接信息
- 代理验证参数
- 爬虫更新间隔
- 日志配置

## License

MIT License
