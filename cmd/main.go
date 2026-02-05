package main

import (
	"fmt"
	"os"

	"inspection-tool/cmd/commands"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "inspection-tool",
		Short: "服务器和Kubernetes巡检工具",
		Long: `一个用于服务器基础资源和Kubernetes环境巡检的工具。

支持功能:
- 服务器资源巡检(CPU、内存、磁盘、网络等)
- Kubernetes集群巡检(节点、Pod、控制平面等)
- 生成详细的巡检报告
- 问题分析和建议

示例:
  # 服务器巡检
  inspection-tool server --host 192.168.1.100 --user root --password pass

  # Kubernetes巡检
  inspection-tool k8s --kubeconfig ~/.kube/config

  # 混合巡检
  inspection-tool all --kubeconfig ~/.kube/config`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// 添加子命令
	rootCmd.AddCommand(commands.NewServerCommand())
	rootCmd.AddCommand(commands.NewK8sCommand())
	rootCmd.AddCommand(commands.NewAllCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
