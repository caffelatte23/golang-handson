// todo は cobra を使った TODO 管理 CLI のサンプルです。
package main

import (
	"os"

	"github.com/caffelatte23/golang-handson/handson/todo/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
