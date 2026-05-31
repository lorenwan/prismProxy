package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemProxy 系统代理管理器
type SystemProxy struct {
	proxyAddr string
}

// NewSystemProxy 创建系统代理管理器
func NewSystemProxy(proxyAddr string) *SystemProxy {
	return &SystemProxy{
		proxyAddr: proxyAddr,
	}
}

// Enable 启用系统代理
func (sp *SystemProxy) Enable() error {
	switch runtime.GOOS {
	case "darwin":
		return sp.enableMacOS()
	case "linux":
		return sp.enableLinux()
	case "windows":
		return sp.enableWindows()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// Disable 禁用系统代理
func (sp *SystemProxy) Disable() error {
	switch runtime.GOOS {
	case "darwin":
		return sp.disableMacOS()
	case "linux":
		return sp.disableLinux()
	case "windows":
		return sp.disableWindows()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// IsEnabled 检查系统代理是否启用
func (sp *SystemProxy) IsEnabled() (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		return sp.isEnabledMacOS()
	case "linux":
		return sp.isEnabledLinux()
	case "windows":
		return sp.isEnabledWindows()
	default:
		return false, nil
	}
}

// macOS 实现
func (sp *SystemProxy) enableMacOS() error {
	// 获取当前网络服务
	service, err := sp.getMacOSNetworkService()
	if err != nil {
		return err
	}

	// 设置 HTTP 代理
	if err := exec.Command("networksetup", "-setwebproxy", service, "127.0.0.1", sp.getPort()).Run(); err != nil {
		return fmt.Errorf("设置 HTTP 代理失败: %v", err)
	}

	// 设置 HTTPS 代理
	if err := exec.Command("networksetup", "-setsecurewebproxy", service, "127.0.0.1", sp.getPort()).Run(); err != nil {
		return fmt.Errorf("设置 HTTPS 代理失败: %v", err)
	}

	// 设置 SOCKS 代理（可选）
	// exec.Command("networksetup", "-setsocksfirewallproxy", service, "127.0.0.1", sp.getPort()).Run()

	return nil
}

func (sp *SystemProxy) disableMacOS() error {
	service, err := sp.getMacOSNetworkService()
	if err != nil {
		return err
	}

	// 关闭 HTTP 代理
	exec.Command("networksetup", "-setwebproxystate", service, "off").Run()

	// 关闭 HTTPS 代理
	exec.Command("networksetup", "-setsecurewebproxystate", service, "off").Run()

	return nil
}

func (sp *SystemProxy) isEnabledMacOS() (bool, error) {
	service, err := sp.getMacOSNetworkService()
	if err != nil {
		return false, err
	}

	// 检查 HTTP 代理状态
	out, err := exec.Command("networksetup", "-getwebproxy", service).Output()
	if err != nil {
		return false, nil
	}

	return strings.Contains(string(out), "Enabled: Yes"), nil
}

func (sp *SystemProxy) getMacOSNetworkService() (string, error) {
	// 获取当前网络服务名称
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return "", fmt.Errorf("获取网络服务列表失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] { // 跳过第一行标题
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") {
			// 优先使用 Wi-Fi 或 Ethernet
			if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "Ethernet") {
				return line, nil
			}
		}
	}

	// 如果没找到，返回第一个有效服务
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") {
			return line, nil
		}
	}

	return "", fmt.Errorf("未找到网络服务")
}

// Linux 实现
func (sp *SystemProxy) enableLinux() error {
	// 尝试使用 gsettings (GNOME)
	if err := sp.enableLinuxGSettings(); err == nil {
		return nil
	}

	// 尝试使用环境变量
	return sp.enableLinuxEnv()
}

func (sp *SystemProxy) disableLinux() error {
	// 尝试使用 gsettings (GNOME)
	if err := sp.disableLinuxGSettings(); err == nil {
		return nil
	}

	// 尝试使用环境变量
	return sp.disableLinuxEnv()
}

func (sp *SystemProxy) isEnabledLinux() (bool, error) {
	// 检查 gsettings
	if enabled, err := sp.isEnabledLinuxGSettings(); err == nil {
		return enabled, nil
	}

	// 检查环境变量
	return sp.isEnabledLinuxEnv()
}

func (sp *SystemProxy) enableLinuxGSettings() error {
	// GNOME 桌面环境
	proxyHost := "127.0.0.1"
	proxyPort := sp.getPort()

	// 设置 HTTP 代理
	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", proxyHost).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "port", proxyPort).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "host", proxyHost).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "port", proxyPort).Run()

	return nil
}

func (sp *SystemProxy) disableLinuxGSettings() error {
	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	return nil
}

func (sp *SystemProxy) isEnabledLinuxGSettings() (bool, error) {
	out, err := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode").Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(out)) == "'manual'", nil
}

func (sp *SystemProxy) enableLinuxEnv() error {
	// 写入到 ~/.bashrc 或 ~/.profile
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envVars := fmt.Sprintf(`
# PrismProxy 系统代理
export http_proxy="http://127.0.0.1:%s"
export https_proxy="http://127.0.0.1:%s"
export no_proxy="localhost,127.0.0.1,::1"
`, sp.getPort(), sp.getPort())

	// 追加到 .bashrc
	bashrcPath := home + "/.bashrc"
	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(envVars); err != nil {
		return err
	}

	return nil
}

func (sp *SystemProxy) disableLinuxEnv() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bashrcPath := home + "/.bashrc"
	content, err := os.ReadFile(bashrcPath)
	if err != nil {
		return err
	}

	// 移除 PrismProxy 代理设置
	lines := strings.Split(string(content), "\n")
	var newLines []string
	skipNext := false

	for _, line := range lines {
		if strings.Contains(line, "# PrismProxy 系统代理") {
			skipNext = true
			continue
		}
		if skipNext && (strings.HasPrefix(line, "export http_proxy=") ||
			strings.HasPrefix(line, "export https_proxy=") ||
			strings.HasPrefix(line, "export no_proxy=")) {
			continue
		}
		skipNext = false
		newLines = append(newLines, line)
	}

	return os.WriteFile(bashrcPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func (sp *SystemProxy) isEnabledLinuxEnv() (bool, error) {
	// 检查当前环境变量
	httpProxy := os.Getenv("http_proxy")
	httpsProxy := os.Getenv("https_proxy")

	return httpProxy != "" || httpsProxy != "", nil
}

// Windows 实现
func (sp *SystemProxy) enableWindows() error {
	// 使用 PowerShell 设置注册表
	proxyServer := fmt.Sprintf("127.0.0.1:%s", sp.getPort())

	// 启用代理
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Set-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings' -Name ProxyEnable -Value 1; Set-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings' -Name ProxyServer -Value '%s'", proxyServer))

	return cmd.Run()
}

func (sp *SystemProxy) disableWindows() error {
	// 禁用代理
	cmd := exec.Command("powershell", "-Command",
		"Set-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings' -Name ProxyEnable -Value 0")

	return cmd.Run()
}

func (sp *SystemProxy) isEnabledWindows() (bool, error) {
	// 检查代理状态
	out, err := exec.Command("powershell", "-Command",
		"(Get-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings').ProxyEnable").Output()

	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(out)) == "1", nil
}

// 辅助函数
func (sp *SystemProxy) getPort() string {
	// 从 proxyAddr 中提取端口
	parts := strings.Split(sp.proxyAddr, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "8888" // 默认端口
}
