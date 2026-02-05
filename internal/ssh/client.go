package ssh

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client SSH客户端
type Client struct {
	conn   *ssh.Client
	config *Config
}

// Config SSH配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

// NewClient 创建SSH客户端
func NewClient(config *Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Client{
		conn:   conn,
		config: config,
	}, nil
}

// Execute 执行命令
func (c *Client) Execute(cmd string) (string, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// ExecuteWithTimeout 带超时的命令执行
func (c *Client) ExecuteWithTimeout(cmd string, timeout time.Duration) (string, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 创建管道
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// 读取输出(带超时)
	type result struct {
		output string
		err    error
	}

	resultChan := make(chan result, 1)
	go func() {
		outBytes, _ := io.ReadAll(stdout)
		errBytes, _ := io.ReadAll(stderr)
		output := string(outBytes) + string(errBytes)
		err := session.Wait()
		resultChan <- result{output: output, err: err}
	}()

	select {
	case res := <-resultChan:
		if res.err != nil {
			return res.output, fmt.Errorf("command failed: %w", res.err)
		}
		return res.output, nil
	case <-time.After(timeout):
		session.Signal(ssh.SIGKILL)
		return "", fmt.Errorf("command timeout after %v", timeout)
	}
}

// CopyFile 复制文件到远程服务器
func (c *Client) CopyFile(localPath, remotePath string, content []byte) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 使用scp协议上传文件
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", len(content), remotePath)
		w.Write(content)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("scp failed: %w", err)
	}

	return nil
}

// TestConnection 测试连接
func (c *Client) TestConnection() error {
	_, err := c.Execute("echo test")
	return err
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetHost 获取主机地址
func (c *Client) GetHost() string {
	return c.config.Host
}

// IsAlive 检查连接是否存活
func (c *Client) IsAlive() bool {
	if c.conn == nil {
		return false
	}

	// 尝试建立一个新会话来测试连接
	session, err := c.conn.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}

// GetLocalIP 获取本地IP(用于排除巡检流量)
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}
