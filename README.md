# 快代理 CLI Test

快代理命令行工具的开源测试发行包，用于查询账户与订单、提取代理、管理白名单，以及执行已授权的新购订单和工单操作。

npm 包：`@zhuangzhuangzhao/kdl-agent-test`。安装后的命令：`kdl-agent-test`。
当前版本：`0.1.0-alpha.1`。

## 安装

需要 Node.js 22 或更高版本。npm 包包含 Go 编译好的可执行文件，安装后可以直接运行。

```bash
npm install -g @zhuangzhuangzhao/kdl-agent-test
kdl-agent-test --version
kdl-agent-test --help
```

也可以直接运行：

```bash
npx --yes --package=@zhuangzhuangzhao/kdl-agent-test kdl-agent-test --help
```

或安装到项目中：

```bash
npm install @zhuangzhuangzhao/kdl-agent-test
npx kdl-agent-test --help
```

平台覆盖 macOS 和 Linux，架构为 arm64 和 x64。对应平台的构建、Go 测试和打包安装测试由 [GitHub Actions](https://github.com/zhuangzhuangzhao921/kdl-agent-cli-test/actions) 执行。Windows 发行另行准备。

## 连接 Gateway

业务操作需要可访问的 KDL Agent Gateway，以及你自己的有效 Agent 凭证。
下面的 Bash 示例通过隐藏输入读取凭证，在当前终端中使用环境变量传入：

```bash
export KDL_AGENT_GATEWAY_URL="https://agent-gateway-api.kuaidaili.com"
export KDL_AGENT_CONFIG="$HOME/.kdl-agent-test/config.toml"
read -r -s -p "Agent token: " KDL_AGENT_TOKEN
printf '\n'
export KDL_AGENT_TOKEN
kdl-agent-test account funds --format json
unset KDL_AGENT_TOKEN
```

Gateway 地址必须对应你的凭证和已开通的服务环境。安装与本地模拟 HTTP 测试通过，不代表你的账户已开通，或线上业务能力已完成验收。

## 常用命令

```bash
kdl-agent-test account summary --format json
kdl-agent-test order list --format json
kdl-agent-test product list --format json
kdl-agent-test proxy --help
kdl-agent-test order whitelist --help
kdl-agent-test ticket --help
```

使用 `<子命令> --help` 查看参数。普通查询校验账户和订单权限；新购待付款订单、创建工单、获取订单密钥分别需要服务端开放相应授权。付款在官网完成。

`order secret get` 会显式输出订单密钥。请只在可信终端或已授权的 Agent 环境中使用，避免将输出写入公共日志。

## 配置与退出

本测试版支持以下环境变量：

| 变量 | 用途 |
| --- | --- |
| `KDL_AGENT_GATEWAY_URL` | Gateway 地址 |
| `KDL_AGENT_TOKEN` | 当前进程使用的 Agent 凭证 |
| `KDL_AGENT_CONFIG` | TOML 配置文件路径 |

`--config` 的路径优先级高于 `KDL_AGENT_CONFIG`；都未提供时，沿用当前目录的 `kdl-agent.toml`。文件内容支持 `gateway_url` 和 `token`，环境变量覆盖对应值。

`auth login` 目前通过参数接收地址和凭证，并以明文 TOML 保存。推荐上面的环境变量输入方式。`auth status` 显示本地配置状态，服务可用性需要通过一次实际查询确认。

`auth logout` 清除配置文件中保存的凭证；环境变量使用 `unset KDL_AGENT_TOKEN` 清除。服务端吊销在会员中心完成。

## 升级和卸载

```bash
npm install -g @zhuangzhuangzhao/kdl-agent-test@0.1.0-alpha.1
npm uninstall -g @zhuangzhuangzhao/kdl-agent-test
```

可通过指定已发布版本安装或回退。卸载移除 npm 安装的程序，用户配置由用户自行管理。

## 开发与验证

需要 Go 1.26.3 和 Node.js 22+：

```bash
git clone https://github.com/zhuangzhuangzhao921/kdl-agent-cli-test.git
cd kdl-agent-cli-test
npm ci
go test ./...
npm run build:local
npm test
npm run test:package
```

完整 npm 发行包：

```bash
npm run build
npm run verify:bundle
npm run test:package
npm pack
```

打包白名单包含启动入口、平台二进制、版本与 SHA256 清单、README 和许可证。
`test:package` 会将真实 tarball 安装到临时目录，测试帮助、版本、参数错误、模拟 Gateway 查询和凭证撤销响应。

维护者更新版本后，执行完整构建和安装测试，再使用 `npm publish --access public` 发布。

## 许可证

源码采用 [MIT License](LICENSE)。发行包包含 Go 运行时和依赖的许可证全文，见构建生成的 `THIRD_PARTY_NOTICES.txt`。
