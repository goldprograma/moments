package main

import (
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/tokensvc/cmd"
)

// 主函数
func main() {
	pkg.GetConfigFile("./", ".toml")
	cmd.Run("./config.toml")
}
