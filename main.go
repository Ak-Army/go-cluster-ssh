package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/Ak-Army/go-cluster-ssh/cmd"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
)

func main() {
	l := initLogger()

	/*config.LoadVccConfig(AppName)
	if err != nil {
		log.Fatal(err)
	}*/
	l.SetField("version", Version)
	l.SetField("pid", fmt.Sprintf("%d", os.Getpid()))
	l.Info("start...")
	ctx := xlog.NewContext(context.Background(), l)

	c := cli.New("go-cluster-ssh", Version)
	cli.RootCommand().Authors = []string{"Ak-Army"}
	c.SetDefault("run")
	c.Run(ctx, os.Args)
}

func initLogger() xlog.Logger {
	xlog.SetLogger(xlog.NopLogger)
	multiOutput := xlog.MultiOutput{}
	multiOutput = append(multiOutput, xlog.NewConsoleOutput())
	level := xlog.LevelInfo
	for _, v := range os.Args {
		if v == "-v" {
			level = xlog.LevelDebug
		}
	}
	conf := xlog.Config{
		Level:  level,
		Output: multiOutput,
	}
	log.SetFlags(0)
	l := xlog.New(conf)
	xlog.SetLogger(l)
	log.SetOutput(l)

	return l
}
