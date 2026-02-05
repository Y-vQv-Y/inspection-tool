# Server & Kubernetes Inspection Tool

一个用于服务器基础资源和Kubernetes环境巡检的Go语言工具。

## 功能特性

### 服务器巡检
- CPU检测：负载、调度、性能、瓶颈问题
- 内存检测：内存压力、使用率
- 磁盘检测：磁盘IO、IO错误、读写瓶颈、iowait、Inode使用
- 网络检测：实时接收/发送流量、数据包错误、TCP连接信息
- 系统检测：文件句柄、内核参数、阻塞任务、时间偏差、上下文切换

### Kubernetes巡检
- 节点健康状态
- API Server状态
- etcd集群健康
- Pod资源使用和状态
- 控制平面组件状态
- 资源配额和限制

## 项目结构

```
inspection-tool/
├── cmd/                    # 命令行入口
├── internal/              # 内部实现
│   ├── server/           # 服务器巡检
│   ├── k8s/              # K8s巡检
│   └── ssh/              # SSH连接管理
├── pkg/                   # 可导出包
│   ├── models/           # 数据模型
│   ├── report/           # 报告生成
│   └── utils/            # 工具函数
├── configs/              # 配置文件
└── scripts/              # 脚本文件
```

## 使用方法

### 编译
```bash
go build -o inspection-tool cmd/main.go
```

### 服务器巡检
```bash
./inspection-tool server \
  --host 192.168.1.100 \
  --user root \
  --password yourpassword \
  --port 22 \
  --output report.json
```

### Kubernetes巡检
```bash
./inspection-tool k8s \
  --kubeconfig ~/.kube/config \
  --output k8s-report.json
```

### 混合巡检
```bash
./inspection-tool all \
  --kubeconfig ~/.kube/config \
  --output full-report.json
```

## 配置文件示例

见 `configs/config.yaml`

## 依赖

- Go 1.21+
- SSH客户端库
- Kubernetes客户端库
