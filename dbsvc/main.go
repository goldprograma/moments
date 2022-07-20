package main

import (
	"gitlab.moments.im/dbsvc/cmd"
	"gitlab.moments.im/pkg"
)

// 主函数
func main() {
	pkg.GetConfigFile("./", ".toml")
	cmd.Run("./config.toml")
}
