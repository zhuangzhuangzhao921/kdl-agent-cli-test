// kdl-agent-test 是快代理 CLI 的测试发行入口。
package main

import (
	"os"

	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
