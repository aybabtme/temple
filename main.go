package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/codegangsta/cli"
)

var (
	version = "devel"
	appname = "temple"
	usage   = "renders Go template on the command line"

	srcFlag   = cli.StringFlag{Name: "src", Usage: "if specified, a source file to read the template from"}
	dstFlag   = cli.StringFlag{Name: "dst", Usage: "if specified, a destination file to write the rendered template to"}
	varFlag   = cli.StringSliceFlag{Name: "var", Usage: "key=values to interpolate in the source template"}
	debugFlag = cli.BoolFlag{Name: "d", Usage: "debug, prints debug information to stderr"}

	D            bool
	templateText string
	dst          io.WriteCloser = os.Stdout
	vars                        = map[string]string{}
)

func main() {
	log.SetPrefix(appname + ": ")
	log.SetFlags(0)

	if err := newApp().Run(os.Args); err != nil {
		log.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		log.Fatal(err)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = appname
	app.Author = "Antoine Grondin"
	app.Email = "antoinegrondin@gmail.com"
	app.Usage = usage
	app.Version = version

	app.Flags = []cli.Flag{srcFlag, dstFlag, varFlag, debugFlag}
	app.Before = func(ctx *cli.Context) error {

		D = ctx.GlobalBool(debugFlag.Name)

		var src io.Reader = os.Stdin
		if srcFile := ctx.GlobalString(srcFlag.Name); srcFile != "" {
			fd, err := os.Open(srcFile)
			if err != nil {
				return fmt.Errorf("can't open -src: %v", err)
			}
			src = fd
		}

		if dstFile := ctx.GlobalString(dstFlag.Name); dstFile != "" {
			fd, err := os.Open(dstFile)
			if err != nil {
				return fmt.Errorf("can't open -dst: %v", err)
			}
			dst = fd
		}

		var err error
		vars, err = parseVars(ctx.GlobalStringSlice(varFlag.Name))
		if err != nil {
			return err
		}

		templateRaw, err := ioutil.ReadAll(src)
		if err != nil {
			return fmt.Errorf("can't read template: %v", err)
		}
		templateText = string(templateRaw)
		return nil
	}
	app.Action = func(ctx *cli.Context) {
		tmpl, err := template.New("").Parse(templateText)
		if err != nil {
			log.Fatalf("can't parse template: %v", err)
		}
		if err := tmpl.Execute(dst, vars); err != nil {
			log.Fatalf("can't render template: %v", err)
		}
	}
	return app
}

func parseVars(vars []string) (map[string]string, error) {
	out := make(map[string]string, len(vars))
	for _, flag := range vars {
		kv := strings.Split(flag, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key=value, need 2 parts but has %d: -var %q", len(kv), flag)
		}
		key := kv[0]
		value := kv[1]
		if _, ok := out[key]; ok {
			return nil, fmt.Errorf("duplicated key: %q already has a value", key)
		}
		out[key] = value
		if D {
			log.Printf("%q=%q", key, value)
		}
	}
	return out, nil
}
