# 后端代码风格规范

## Go

### 命名规范

- 包名：小写，单数（如 `traffic`）
- 导出类型：PascalCase（如 `Service`）
- 私有类型：camelCase（如 `serviceConfig`）
- 接口：PascalCase，以 `er` 结尾（如 `Repository`）

### 错误处理

- 使用 gRPC Status Code 返回错误
- 错误信息应包含上下文
- 使用 `fmt.Errorf` 包装错误，保留原始错误

### 包结构

```
internal/
├── [module]/        # 业务模块
│   ├── service.go   # 业务逻辑
│   ├── repo.go      # 数据访问
│   └── types.go     # 类型定义
├── grpc/            # gRPC 服务实现
└── storage/         # 存储层
```

---

## gRPC

### Proto 文件规范

- 语法：proto3
- 包名：小写（如 `prismproxy`）
- 服务名：PascalCase，以 `Service` 结尾
- 消息名：PascalCase
- 字段名：snake_case

### 服务定义

- 查询方法：`List` / `Get`
- 变更方法：`Create` / `Update` / `Delete`
- 订阅方法：`Subscribe`

---

## 数据库

### SQL 规范

- 表名：snake_case，复数（如 `traffic_items`）
- 字段名：snake_case
- 索引名：`idx_表名_字段`

### 迁移规范

- 迁移文件按版本号排序
- 每个迁移文件只包含一个变更
- 迁移必须可回滚

---

## 目录结构

```
internal/
├── traffic/         # 流量模块
├── rules/           # 规则模块
├── ai/              # AI 模块
├── grpc/            # gRPC 服务实现
├── storage/         # 存储层
├── proxy/           # HTTP 代理引擎
├── cert/            # 证书管理
├── codegen/         # 代码生成
├── collection/      # API 集合管理
├── debugger/        # 断点调试
├── diff/            # Diff 对比
├── environment/     # 环境变量
├── perf/            # 性能分析
├── rewrite/         # 请求重写
├── script/          # 脚本引擎
├── search/          # 搜索增强
└── websocket/       # WebSocket
```

---

## 安全规范

- 证书私钥不得日志输出
- 敏感数据必须脱敏存储
- SQL 查询使用参数化，防止注入
