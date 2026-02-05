# 使用指南

## 快速开始

### 1. 编译项目

```bash
# 基本编译
make build

# 编译所有平台版本
make build-all

# 清理编译文件
make clean
```

### 2. 服务器巡检

#### 基本用法

```bash
./inspection-tool server \
  --host 192.168.1.100 \
  --user root \
  --password yourpassword \
  --port 22
```

#### 高级选项

```bash
./inspection-tool server \
  --host 192.168.1.100 \
  --user root \
  --password yourpassword \
  --port 22 \
  --output ./my-reports \
  --format yaml \
  --detailed
```

### 3. Kubernetes巡检

#### 基本用法

```bash
./inspection-tool k8s --kubeconfig ~/.kube/config
```

#### 指定命名空间

```bash
./inspection-tool k8s \
  --kubeconfig ~/.kube/config \
  --namespaces default,kube-system,production
```

#### 同时巡检Worker节点

```bash
./inspection-tool k8s \
  --kubeconfig ~/.kube/config \
  --inspect-workers \
  --ssh-user root \
  --ssh-password yourpassword
```

### 4. 综合巡检

```bash
./inspection-tool all \
  --kubeconfig ~/.kube/config \
  --hosts "192.168.1.10,192.168.1.11,192.168.1.12" \
  --ssh-user root \
  --ssh-password yourpassword
```

## 配置文件

可以使用配置文件来简化命令行参数:

```yaml
# configs/config.yaml
server:
  timeout: 30
  interval: 60
  thresholds:
    cpu:
      load_1min: 8.0
      usage_percent: 80.0
    memory:
      usage_percent: 85.0
    disk:
      usage_percent: 85.0

k8s:
  kubeconfig: "~/.kube/config"
  namespaces: []
  thresholds:
    node:
      cpu_usage_percent: 80.0
      memory_usage_percent: 85.0
```

## 巡检指标说明

### 服务器指标

#### CPU
- **load**: 1分钟、5分钟、15分钟平均负载
- **usage**: CPU使用率(user、system、idle、iowait等)
- **context_switches**: 上下文切换次数
- **run_queue**: 运行队列长度
- **blocked_tasks**: 阻塞任务数

#### 内存
- **total/used/free/available**: 内存总量和使用情况
- **swap**: 交换分区使用情况
- **cached/buffers**: 缓存和缓冲区
- **pressure**: 内存压力指标

#### 磁盘
- **usage**: 磁盘空间使用率
- **inodes**: Inode使用情况
- **io_stats**: IO统计(读写速率、IOPS、利用率)
- **io_errors**: IO错误计数

#### 网络
- **interfaces**: 各网络接口的收发流量和错误率
- **tcp_stats**: TCP连接状态统计
- **retransmits**: TCP重传统计

#### 系统
- **file_handles**: 文件句柄使用情况
- **process_count**: 进程和线程数
- **time_offset**: 时间偏差
- **kernel_params**: 关键内核参数

### Kubernetes指标

#### 集群
- **version**: 集群版本
- **node_count**: 节点数量
- **pod_count**: Pod总数
- **namespace_count**: 命名空间数量

#### 节点
- **ready**: 节点就绪状态
- **conditions**: 节点状态条件
- **resource_usage**: CPU、内存、Pod使用率
- **taints**: 节点污点

#### 控制平面
- **apiserver**: API Server健康状态
- **etcd**: etcd集群状态
- **controller**: Controller Manager状态
- **scheduler**: Scheduler状态

#### Pod
- **phase**: Pod状态
- **ready**: 就绪状态
- **restart_count**: 重启次数
- **resource_usage**: 资源使用情况

## 报告格式

### JSON格式

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "type": "server",
  "server_report": {
    "host": "192.168.1.100",
    "cpu": { ... },
    "memory": { ... },
    "issues": [
      {
        "level": "critical",
        "category": "memory",
        "message": "内存使用率过高: 92.50%",
        "suggestion": "释放内存或增加物理内存"
      }
    ]
  }
}
```

### YAML格式

```yaml
timestamp: 2024-01-01T00:00:00Z
type: server
server_report:
  host: 192.168.1.100
  cpu:
    ...
  memory:
    ...
  issues:
    - level: critical
      category: memory
      message: "内存使用率过高: 92.50%"
      suggestion: "释放内存或增加物理内存"
```

## 问题级别

- **critical**: 严重问题,需要立即处理
- **warning**: 警告,需要关注
- **info**: 信息性提示

## 常见问题

### 1. SSH连接失败

确保:
- 目标服务器SSH服务正常运行
- 防火墙允许SSH端口
- 用户名和密码正确
- 网络连通性良好

### 2. Kubernetes连接失败

确保:
- kubeconfig文件路径正确
- 集群API Server可访问
- 证书有效且未过期

### 3. Metrics Server不可用

如果K8s集群未安装Metrics Server,将无法获取节点和Pod的资源使用指标,但不影响其他巡检功能。

安装Metrics Server:
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## 最佳实践

1. **定期巡检**: 建议每天执行一次完整巡检
2. **保留历史**: 配置报告保留天数,便于趋势分析
3. **阈值调整**: 根据实际情况调整告警阈值
4. **自动化**: 使用cron或Jenkins等工具实现自动化巡检
5. **告警集成**: 将关键问题接入告警系统

## 自动化示例

### Cron定时任务

```bash
# 每天凌晨2点执行巡检
0 2 * * * /path/to/inspection-tool all --kubeconfig ~/.kube/config --ssh-user root --ssh-password pass > /var/log/inspection.log 2>&1
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    triggers {
        cron('0 2 * * *')
    }
    stages {
        stage('Inspection') {
            steps {
                sh '''
                    ./inspection-tool all \
                      --kubeconfig /path/to/kubeconfig \
                      --ssh-user root \
                      --ssh-password ${SSH_PASSWORD} \
                      --output ./reports
                '''
            }
        }
        stage('Archive') {
            steps {
                archiveArtifacts artifacts: 'reports/*.json', fingerprint: true
            }
        }
    }
}
```

## 贡献

欢迎提交Issue和Pull Request!
