package main

import (
	"moments/forumsvc/cmd"
	"moments/pkg"
)

// 主函数
func main() {
	pkg.GetConfigFile("./", ".toml")
	cmd.Run("./config.toml")
}
