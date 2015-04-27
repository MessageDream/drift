package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/MessageDream/drift/cmd"
	"github.com/MessageDream/drift/modules/setting"
)

const APP_VER = "0.0.1.0910 Beta"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	setting.AppVer = APP_VER
}

func main() {
	app := cli.NewApp()
	app.Name = "Gogs"
	app.Usage = "Go Git Service"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdWeb,
		cmd.CmdDump,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
