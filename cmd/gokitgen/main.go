package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/l-vitaly/gokitgen/pkg/generators"
	"github.com/l-vitaly/gokitgen/pkg/parser"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "s",
		},
		cli.StringFlag{
			Name: "p",
		},
	}
	app.Before = func(c *cli.Context) error {
		path, err := filepath.Abs(c.String("p"))
		if err != nil {
			return err
		}
		basePath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		result, err := new(parser.Parser).Parse(basePath, c.String("s"))
		if err != nil {
			return err
		}
		c.App.Metadata["result"] = result
		c.App.Metadata["path"] = path
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "transport",
			Aliases: []string{"t"},
			Usage:   "",
			Subcommands: []cli.Command{
				{
					Name: "http",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "zipkin",
						},
						cli.BoolFlag{
							Name: "logger",
						},
						cli.BoolFlag{
							Name: "greq",
						},
						cli.BoolFlag{
							Name: "gresp",
						},
						cli.BoolFlag{
							Name: "c",
						},
					},
					Action: func(c *cli.Context) error {
						transportGenerator := generators.NewHTTPTransport(
							generators.HTTPGeneratorZipkin(c.Bool("zipkin")),
							generators.HTTPGeneratorClient(c.Bool("c")),
							generators.HTTPGeneratorLogger(c.Bool("logger")),
							generators.HTTPGeneratorGenericRequest(c.Bool("greq")),
							generators.HTTPGeneratorGenericResponse(c.Bool("gresp")),
						)
						data, err := transportGenerator.Generate(c.App.Metadata["result"].(parser.Result))
						if err != nil {
							return err
						}

						savePath := c.App.Metadata["path"].(string)
						if err := ioutil.WriteFile(savePath+"/http.go", data, 0755); err != nil {
							return err
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "endpoint",
			Aliases: []string{"e"},
			Usage:   "",
			Action: func(c *cli.Context) error {
				transportGenerator := generators.NewEndpoint()
				data, err := transportGenerator.Generate(c.App.Metadata["result"].(parser.Result))
				if err != nil {
					return err
				}
				savePath := c.App.Metadata["path"].(string)
				if err := ioutil.WriteFile(savePath+"/endpoints.go", data, 0755); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:    "logging",
			Aliases: []string{"lg"},
			Usage:   "",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "st",
				},
			},
			Action: func(c *cli.Context) error {
				transportGenerator := generators.NewLogging(
					generators.LoggingGeneratorEnableStackTrace(c.Bool("st")),
				)
				data, err := transportGenerator.Generate(c.App.Metadata["result"].(parser.Result))
				if err != nil {
					return err
				}
				savePath := c.App.Metadata["path"].(string)
				if err := ioutil.WriteFile(savePath+"/logging.go", data, 0755); err != nil {
					return err
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
