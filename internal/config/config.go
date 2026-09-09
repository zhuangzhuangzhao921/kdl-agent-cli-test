// Package config 负责 CLI 配置加载、持久化与环境变量覆写。
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	// DefaultFilename 未指定 KDL_AGENT_CONFIG 时，相对 CWD 的默认配置文件名。
	DefaultFilename = "kdl-agent.toml"

	envConfigPath = "KDL_AGENT_CONFIG"
	envGatewayURL = "KDL_AGENT_GATEWAY_URL"
	envToken      = "KDL_AGENT_TOKEN"
)

// pathOverride 由根命令 --config 设置；优先级高于 KDL_AGENT_CONFIG 与 CWD 默认路径。
var pathOverride string

// File 持久化配置；token 不得写入日志或 stdout。
type File struct {
	GatewayURL string `toml:"gateway_url"`
	Token      string `toml:"token"`
}

// Resolved 合并文件与环境变量后的有效配置。
type Resolved struct {
	GatewayURL string
	Token      string
	ConfigPath string
}

// SetPathOverride 设置 CLI --config 指定的配置文件路径；传空字符串表示清除覆写。
func SetPathOverride(path string) {
	pathOverride = strings.TrimSpace(path)
}

// ResolvePath 返回将使用的配置文件绝对路径（文件可不存在）。
// 优先级：--config > KDL_AGENT_CONFIG > CWD/kdl-agent.toml。
func ResolvePath() (string, error) {
	if pathOverride != "" {
		abs, err := filepath.Abs(pathOverride)
		if err != nil {
			return "", fmt.Errorf("解析 --config 路径失败: %w", err)
		}
		return abs, nil
	}
	if v := strings.TrimSpace(os.Getenv(envConfigPath)); v != "" {
		abs, err := filepath.Abs(v)
		if err != nil {
			return "", fmt.Errorf("解析 KDL_AGENT_CONFIG 失败: %w", err)
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	return filepath.Join(cwd, DefaultFilename), nil
}

// Load 读取配置：TOML 文件 + 环境变量覆写（env 优先）。
func Load() (Resolved, error) {
	path, err := ResolvePath()
	if err != nil {
		return Resolved{}, err
	}

	var file File
	if raw, readErr := os.ReadFile(path); readErr == nil {
		if err := toml.Unmarshal(raw, &file); err != nil {
			return Resolved{}, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return Resolved{}, fmt.Errorf("读取配置文件 %s 失败: %w", path, readErr)
	}

	resolved := Resolved{
		GatewayURL: strings.TrimSpace(file.GatewayURL),
		Token:      strings.TrimSpace(file.Token),
		ConfigPath: path,
	}

	if v := strings.TrimSpace(os.Getenv(envGatewayURL)); v != "" {
		resolved.GatewayURL = v
	}
	if v := strings.TrimSpace(os.Getenv(envToken)); v != "" {
		resolved.Token = v
	}

	return resolved, nil
}

// Save 写入配置文件并设置权限 0600。
func Save(path string, file File) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("配置文件路径为空")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("创建配置目录 %s 失败: %w", dir, err)
	}
	raw, err := toml.Marshal(file)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("写入配置文件 %s 失败: %w", path, err)
	}
	return nil
}

// ValidateForAPI 检查调用 Gateway 前的最低要求。
func ValidateForAPI(r Resolved) error {
	if strings.TrimSpace(r.GatewayURL) == "" {
		return fmt.Errorf(
			"未配置 Gateway 地址\n  配置文件: %s\n  修复: 设置环境变量 KDL_AGENT_GATEWAY_URL，或运行 kdl-agent-test auth login --help",
			r.ConfigPath,
		)
	}
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf(
			"未配置 Agent 凭证\n  配置文件: %s\n  修复: 设置环境变量 KDL_AGENT_TOKEN，或运行 kdl-agent-test auth login --help",
			r.ConfigPath,
		)
	}
	return nil
}
