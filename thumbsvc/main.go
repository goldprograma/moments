package main

import (
	"moments/pkg"
	"moments/thumbsvc/cmd"
)

// 主函数
func main() {
	pkg.GetConfigFile("./", ".toml")
	cmd.Run("./config.toml")
}
