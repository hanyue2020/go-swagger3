package app

import (
	"strings"

	"github.com/urfave/cli"
)

type args struct {
	flags []cli.Flag

	modulePath    string
	domain        string
	basicInfoPath string
	handlerPath   string
	output        string
	debug         bool
	strict        bool
	generateYaml  bool
}

func LoadArgs(c *cli.Context) *args {
	basicInfoPath := c.GlobalString("basic-info-path")
	if basicInfoPath == "" {
		basicInfoPath = c.GlobalString("main-file-path")
	}
	appArgs := args{
		flags:         flags,
		modulePath:    c.GlobalString("module-path"),
		domain:        c.GlobalString("domain"),
		basicInfoPath: basicInfoPath,
		handlerPath:   c.GlobalString("handler-path"),
		output:        c.GlobalString("output"),
		debug:         c.GlobalBool("debug"),
		strict:        c.GlobalBool("strict"),
		generateYaml:  c.GlobalBool("generate-yaml"),
	}
	if appArgs.generateYaml && strings.HasSuffix(appArgs.output, ".json") {
		appArgs.output = strings.TrimSuffix(appArgs.output, ".json") + ".yaml"
	}
	return &appArgs

}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "module-path",
		Value: "",
		Usage: "go-swagger3 will search @comment under the module",
	},
	cli.StringFlag{
		Name:  "domain",
		Value: "",
		Usage: "auto-detect the Go module and domain/<name> handler paths",
	},
	cli.StringFlag{
		Name:  "basic-info-path",
		Value: "",
		Usage: "go-swagger3 will start to search @comment from this path file, default auto-detect main.go",
	},
	cli.StringFlag{
		Name:  "main-file-path",
		Value: "",
		Usage: "alias of --basic-info-path; useful when a project contains multiple service mains under different domains",
	},
	cli.StringFlag{
		Name:  "handler-path",
		Value: "",
		Usage: "go-swagger3 only search handleFunc comments under the path",
	},
	cli.StringFlag{
		Name:  "output",
		Value: "oas.json",
		Usage: "output file",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
	cli.BoolFlag{
		Name:  "strict",
		Usage: "convert go parsing warnings to fatal errors",
	},
	cli.BoolFlag{
		Name:  "schema-without-pkg",
		Usage: "create schemas without package name append to the name",
	},
	cli.BoolFlag{
		Name:  "generate-yaml",
		Usage: "generate yaml spec if true",
	},
}
