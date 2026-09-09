package cmd

import (
	"fmt"
	"strings"

	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/internal/cmdutil"
	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/internal/config"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Agent 凭证与 Gateway 连接配置",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var gatewayURL, token string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "保存 Gateway 地址与 Agent Bearer 凭证",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := newFactory()
			path, err := config.ResolvePath()
			if err != nil {
				return fail(f, err, cmdutil.ExitConfig)
			}
			gatewayURL = strings.TrimSpace(gatewayURL)
			token = strings.TrimSpace(token)
			if gatewayURL == "" || token == "" {
				return fail(f, fmt.Errorf("需要 --gateway-url 与 --token"), cmdutil.ExitConfig)
			}
			if err := config.Save(path, config.File{
				GatewayURL: gatewayURL,
				Token:      token,
			}); err != nil {
				return fail(f, err, cmdutil.ExitConfig)
			}
			if strings.EqualFold(globalFormat, "json") {
				return writeJSON(f.Out, map[string]any{
					"config_path": path,
					"gateway_url": gatewayURL,
					"saved":       true,
				})
			}
			fmt.Fprintf(f.Out.Stdout, "%s %s\n", f.Out.Green("已写入配置:"), path)
			fmt.Fprintf(f.Out.Stdout, "Gateway: %s\n", gatewayURL)
			fmt.Fprintln(f.Out.Stdout, f.Out.Dim("凭证: (已保存，不在此显示)"))
			return nil
		},
	}
	cmd.Flags().StringVar(&gatewayURL, "gateway-url", "", "Gateway 基址，如 http://127.0.0.1:8080")
	cmd.Flags().StringVar(&token, "token", "", "Agent token: kdl_ag_<selector>.<secret>")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看当前配置（不显示 token 明文）",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := newFactory()
			cfg, err := config.Load()
			if err != nil {
				return fail(f, err, cmdutil.ExitConfig)
			}
			if strings.EqualFold(globalFormat, "json") {
				return writeJSON(f.Out, map[string]any{
					"config_path":      cfg.ConfigPath,
					"gateway_url":      cfg.GatewayURL,
					"token_configured": strings.TrimSpace(cfg.Token) != "",
				})
			}
			fmt.Fprintf(f.Out.Stdout, "配置文件: %s\n", cfg.ConfigPath)
			if cfg.GatewayURL != "" {
				fmt.Fprintf(f.Out.Stdout, "Gateway:  %s\n", cfg.GatewayURL)
			} else {
				fmt.Fprintln(f.Out.Stdout, "Gateway:  (未配置)")
			}
			if strings.TrimSpace(cfg.Token) != "" {
				fmt.Fprintln(f.Out.Stdout, "凭证:     已配置")
			} else {
				fmt.Fprintln(f.Out.Stdout, "凭证:     未配置")
			}
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "清除本地保存的 Agent 凭证",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := newFactory()
			cfg, err := config.Load()
			if err != nil {
				return fail(f, err, cmdutil.ExitConfig)
			}
			if err := config.Save(cfg.ConfigPath, config.File{
				GatewayURL: cfg.GatewayURL,
				Token:      "",
			}); err != nil {
				return fail(f, err, cmdutil.ExitConfig)
			}
			fmt.Fprintf(f.Out.Stdout, "已清除凭证: %s\n", cfg.ConfigPath)
			return nil
		},
	}
}

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "账户只读 Capability",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "funds",
		Short: "查询账户余额（account.funds.read）",
		RunE:  runGET("/v1/account/funds"),
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "summary",
		Short: "查询账户摘要（account.summary.read）",
		RunE:  runGET("/v1/account/summary"),
	})
	return cmd
}
