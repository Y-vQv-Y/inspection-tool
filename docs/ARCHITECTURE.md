# 项目架构

## 目录结构

```
inspection-tool/
├── cmd/                          # 命令行入口
│   ├── main.go                  # 主程序入口
│   └── commands/                # 子命令实现
│       ├── server.go           # 服务器巡检命令
│       ├── k8s.go              # K8s巡检命令
│       └── all.go              # 综合巡检命令
│
├── internal/                     # 内部实现(不对外导出)
│   ├── ssh/                     # SSH连接管理
│   │   └── client.go           # SSH客户端封装
│   ├── server/                  # 服务器巡检实现
│   │   ├── inspector.go        # 巡检核心逻辑
│   │   └── parser.go           # 指标解析器
│   └── k8s/                     # K8s巡检实现
│       └── inspector.go        # K8s巡检核心逻辑
│
├── pkg/                          # 可导出的公共包
│   ├── models/                  # 数据模型定义
│   │   └── models.go           # 所有数据结构
│   ├── report/                  # 报告生成
│   │   └── generator.go        # 报告生成器
│   └── utils/                   # 工具函数
│       └── helpers.go          # 辅助函数
│
├── configs/                      # 配置文件
│   └── config.yaml             # 默认配置
│
├── scripts/                      # 脚本文件
│   └── example.sh              # 示例脚本
│
├── docs/                         # 文档
│   └── USAGE.md                # 使用文档
│
├── go.mod                        # Go模块定义
├── go.sum                        # 依赖锁定
├── Makefile                      # 构建脚本
├── README.md                     # 项目说明
└── .gitignore                    # Git忽略文件
```

## 核心模块说明

### 1. SSH模块 (internal/ssh)

**职责**: 管理SSH连接,执行远程命令

**主要类型**:
- `Client`: SSH客户端
- `Config`: SSH配置

**主要方法**:
- `NewClient()`: 创建SSH连接
- `Execute()`: 执行命令
- `ExecuteWithTimeout()`: 带超时的命令执行
- `Close()`: 关闭连接

### 2. 服务器巡检模块 (internal/server)

**职责**: 执行服务器资源巡检

**主要文件**:
- `inspector.go`: 巡检主逻辑
- `parser.go`: 指标解析

**巡检内容**:
- CPU: load、使用率、上下文切换、阻塞任务等
- 内存: 使用率、可用内存、Swap、内存压力等
- 磁盘: 空间、Inode、IO统计、IO错误等
- 网络: 流量、错误率、TCP连接统计等
- 系统: 文件句柄、进程数、时间同步、内核参数等

**数据流**:
```
SSH Client → Execute Commands → Parse Output → Metrics → Analyze → Issues
```

### 3. Kubernetes巡检模块 (internal/k8s)

**职责**: 执行K8s集群巡检

**主要文件**:
- `inspector.go`: K8s巡检主逻辑

**巡检内容**:
- 集群信息: 版本、节点数、Pod数等
- 节点: 状态、资源使用、条件、污点等
- 控制平面: API Server、etcd、Controller、Scheduler
- Pod: 状态、重启次数、资源使用等

**数据流**:
```
K8s Client → API Calls → Parse Resources → Metrics → Analyze → Issues
```

### 4. 数据模型 (pkg/models)

**职责**: 定义所有数据结构

**主要类型**:
- `InspectionReport`: 完整巡检报告
- `ServerReport`: 服务器巡检报告
- `K8sReport`: K8s巡检报告
- `Issue`: 问题项
- 各类指标结构(CPU、内存、磁盘等)

### 5. 报告生成 (pkg/report)

**职责**: 生成和输出巡检报告

**支持格式**:
- JSON
- YAML

**输出方式**:
- 文件保存
- 终端打印摘要

### 6. 工具函数 (pkg/utils)

**职责**: 提供通用工具函数

**主要功能**:
- 构建巡检摘要
- 格式化输出
- 配置验证
- 百分比计算

## 工作流程

### 服务器巡检流程

```
1. 解析命令行参数
2. 创建SSH连接
3. 测试连接
4. 创建巡检器
5. 收集系统信息
   ├─ OS信息
   ├─ CPU指标
   ├─ 内存指标
   ├─ 磁盘指标
   ├─ 网络指标
   └─ 系统指标
6. 分析问题
7. 生成报告
8. 输出结果
```

### Kubernetes巡检流程

```
1. 解析命令行参数
2. 加载kubeconfig
3. 创建K8s客户端
4. 创建巡检器
5. 收集集群信息
   ├─ 集群基本信息
   ├─ 节点指标
   ├─ API Server状态
   ├─ etcd状态
   ├─ Controller状态
   ├─ Scheduler状态
   └─ Pod信息
6. 分析问题
7. 生成报告
8. 输出结果
```

### 综合巡检流程

```
1. 解析命令行参数
2. 执行K8s巡检
   └─ 获取节点列表
3. 执行服务器巡检
   ├─ 并发巡检多台服务器
   └─ 包括K8s worker节点
4. 合并报告
5. 构建综合摘要
6. 生成报告
7. 输出结果
```

## 并发设计

### 多服务器并发巡检

使用goroutine和信号量实现并发控制:

```go
semaphore := make(chan struct{}, 5) // 限制并发数为5

for _, host := range hosts {
    go func(h string) {
        semaphore <- struct{}{}        // 获取
        defer func() { <-semaphore }() // 释放
        
        // 执行巡检
        inspect(h)
    }(host)
}
```

## 错误处理

### 错误级别

1. **致命错误**: 返回错误,中止执行
   - SSH连接失败
   - K8s API连接失败
   - 报告生成失败

2. **警告错误**: 记录警告,继续执行
   - 单个指标收集失败
   - 部分节点巡检失败
   - Metrics Server不可用

3. **忽略错误**: 静默处理
   - 可选功能失败
   - 最佳努力型检查

### 错误传播

```go
// 致命错误
if err := criticalOperation(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// 警告错误
if err := optionalOperation(); err != nil {
    log.Warn("optional operation failed: %v", err)
    // 继续执行
}
```

## 性能优化

1. **并发执行**: 多服务器并发巡检
2. **超时控制**: 所有操作都有超时限制
3. **批量操作**: K8s API批量查询
4. **缓存复用**: SSH连接复用
5. **增量采样**: IO统计使用两次采样计算差值

## 安全考虑

1. **密码保护**: 
   - 命令行密码不记录到历史
   - 建议使用环境变量或密钥文件

2. **权限最小化**:
   - SSH只需要读权限
   - K8s只需要只读RBAC权限

3. **网络隔离**:
   - 支持指定SSH端口
   - 支持网络策略

## 扩展性

### 添加新的巡检指标

1. 在 `models.go` 中添加数据结构
2. 在 `inspector.go` 中添加收集逻辑
3. 在 `parser.go` 中添加解析逻辑
4. 在 `analyzeIssues()` 中添加分析逻辑

### 添加新的输出格式

1. 在 `report/generator.go` 中添加格式处理
2. 更新命令行参数说明

### 添加新的告警渠道

1. 创建新的 `alert` 包
2. 实现告警接口
3. 在配置文件中添加配置
4. 在报告生成后触发告警

## 最佳实践

1. **命令执行**: 使用超时和重试机制
2. **资源清理**: 使用defer确保资源释放
3. **日志记录**: 记录关键操作和错误
4. **单元测试**: 为核心逻辑编写测试
5. **文档维护**: 保持文档与代码同步
