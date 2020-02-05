package main

import (
	//log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"log"
	"os"
)

const usage = `这是一个获取最新release版本镜像的工具`

func main() {
	app := cli.NewApp()
	app.Name = "app"
	app.Usage = usage

	app.Commands = []cli.Command{
		searchCommand,
		releaseCommand,
	}

	app.Before = func(context *cli.Context) error {

		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
