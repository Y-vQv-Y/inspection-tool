package commands

import (
	"fmt"
	"inspection-tool/internal/server"
	"inspection-tool/internal/ssh"
	"inspection-tool/pkg/report"
	"inspection-tool/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

// ServerOptions 服务器巡检选项
type ServerOptions struct {
	Host     string
	User     string
	Password string
	Port     int
	Output   string
	Format   string
	Detailed bool
}

// NewServerCommand 创建服务器巡检命令
func NewServerCommand() *cobra.Command {
	opts := &ServerOptions{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "执行服务器巡检",
		Long:  `连接到远程服务器并执行资源巡检,包括CPU、内存、磁盘、网络等指标。`,
		Example: `  # 基本用法
  inspection-tool server --host 192.168.1.100 --user root --password yourpass

  # 指定端口和输出格式
  inspection-tool server --host 192.168.1.100 --user root --password yourpass --port 2222 --format yaml

  # 详细输出
  inspection-tool server --host 192.168.1.100 --user root --password yourpass --detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServerInspection(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Host, "host", "", "服务器地址(必需)")
	cmd.Flags().StringVar(&opts.User, "user", "root", "SSH用户名")
	cmd.Flags().StringVar(&opts.Password, "password", "", "SSH密码(必需)")
	cmd.Flags().IntVar(&opts.Port, "port", 22, "SSH端口")
	cmd.Flags().StringVar(&opts.Output, "output", "./reports", "报告输出目录")
	cmd.Flags().StringVar(&opts.Format, "format", "json", "报告格式(json/yaml)")
	cmd.Flags().BoolVar(&opts.Detailed, "detailed", true, "生成详细报告")

	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("password")

	return cmd
}

func runServerInspection(opts *ServerOptions) error {
	fmt.Println("========================================")
	fmt.Println("开始服务器巡检")
	fmt.Println("========================================")
	fmt.Printf("目标主机: %s:%d\n", opts.Host, opts.Port)
	fmt.Printf("用户: %s\n", opts.User)
	fmt.Println("========================================\n")

	// 验证配置
	if err := utils.ValidateConfig(opts.Host, opts.User, opts.Password, opts.Port); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 创建SSH客户端
	fmt.Println("正在连接服务器...")
	sshClient, err := ssh.NewClient(&ssh.Config{
		Host:     opts.Host,
		Port:     opts.Port,
		User:     opts.User,
		Password: opts.Password,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}
	defer sshClient.Close()

	// 测试连接
	if err := sshClient.TestConnection(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}
	fmt.Println("✓ 连接成功\n")

	// 创建巡检器
	inspector, err := server.NewInspector(sshClient)
	if err != nil {
		return fmt.Errorf("创建巡检器失败: %w", err)
	}
	defer inspector.Close()

	// 执行巡检
	fmt.Println("正在执行巡检...")
	serverReport, err := inspector.Inspect()
	if err != nil {
		return fmt.Errorf("巡检失败: %w", err)
	}
	fmt.Println("✓ 巡检完成\n")

	// 生成报告
	fmt.Println("正在生成报告...")
	generator := report.NewGenerator(opts.Format, opts.Output, opts.Detailed)
	reportPath, err := generator.GenerateServerReport(serverReport)
	if err != nil {
		return fmt.Errorf("生成报告失败: %w", err)
	}

	// 打印摘要
	report.PrintServerSummary(serverReport)

	fmt.Printf("\n✓ 报告已保存: %s\n", reportPath)

	// 根据问题数量返回退出码
	if len(serverReport.Issues) > 0 {
		criticalCount := 0
		for _, issue := range serverReport.Issues {
			if issue.Level == "critical" {
				criticalCount++
			}
		}
		if criticalCount > 0 {
			return fmt.Errorf("发现 %d 个严重问题", criticalCount)
		}
	}

	return nil
}
