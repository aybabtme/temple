package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"

	"github.com/codegangsta/cli"
)

var (
	version = "devel"
	appname = "temple"
	usage   = "renders Go template on the command line"

	debugFlag = cli.BoolFlag{Name: "d", Usage: "debug, prints debug information to stderr"}

	D bool
)

func main() {
	log.SetPrefix(appname + ": ")
	log.SetFlags(0)

	if err := newApp().Run(os.Args); err != nil {
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

	app.Flags = []cli.Flag{debugFlag}
	app.Before = func(ctx *cli.Context) error {
		D = ctx.GlobalBool(debugFlag.Name)
		return nil
	}

	app.Commands = []cli.Command{
		fileCmd(),
		treeCmd(),
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
			log.Printf("var %q=%q", key, value)
		}
	}
	return out, nil
}

func fileCmd() cli.Command {

	varFlag := cli.StringSliceFlag{Name: "var", Usage: "key=values to interpolate in the source template"}
	srcFlag := cli.StringFlag{Name: "src", Usage: "if specified, a source file to read the template from"}
	dstFlag := cli.StringFlag{Name: "dst", Usage: "if specified, a destination file to write the rendered template to"}

	var (
		templateText string
		vars                        = map[string]string{}
		dst          io.WriteCloser = os.Stdout
	)

	checkFlags := func(ctx *cli.Context) error {
		var err error
		vars, err = parseVars(ctx.StringSlice(varFlag.Name))
		if err != nil {
			return err
		}

		var src io.Reader = os.Stdin
		if srcFile := ctx.String(srcFlag.Name); srcFile != "" {
			fd, err := os.Open(srcFile)
			if err != nil {
				return fmt.Errorf("can't open -src: %v", err)
			}
			src = fd
		}

		if dstFile := ctx.String(dstFlag.Name); dstFile != "" {
			fd, err := os.Open(dstFile)
			if err != nil {
				return fmt.Errorf("can't open -dst: %v", err)
			}
			dst = fd
		}

		templateRaw, err := ioutil.ReadAll(src)
		if err != nil {
			return fmt.Errorf("can't read template: %v", err)
		}
		templateText = string(templateRaw)
		return nil
	}

	action := func(ctx *cli.Context) {
		defer dst.Close()
		tmpl, err := template.New("").Parse(templateText)
		if err != nil {
			log.Fatalf("can't parse template: %v", err)
		}
		if err := tmpl.Execute(dst, vars); err != nil {
			log.Fatalf("can't render template: %v", err)
		}
	}
	return cli.Command{
		Flags:  []cli.Flag{varFlag, srcFlag, dstFlag},
		Name:   "file",
		Usage:  "render a single file",
		Before: checkFlags,
		Action: action,
	}
}

func treeCmd() cli.Command {

	varFlag := cli.StringSliceFlag{Name: "var", Usage: "key=values to interpolate in the source template"}
	srcFlag := cli.StringFlag{Name: "src", Usage: "path to a tree where templates are found"}
	dstFlag := cli.StringFlag{Name: "dst", Usage: "the root where the rendered tree will be put"}
	overwriteFlag := cli.BoolFlag{Name: "overwrite", Usage: "if specified, will overwrite existing files in the destination"}

	var (
		src       string
		dst       string
		vars      = map[string]string{}
		files     = []struct{ src, dst string }{}
		overwrite = false
	)

	checkFlags := func(ctx *cli.Context) error {
		var err error
		vars, err = parseVars(ctx.StringSlice(varFlag.Name))
		if err != nil {
			return err
		}

		src = ctx.String(srcFlag.Name)
		if src == "" {
			return fmt.Errorf("flag %q is necessary", srcFlag.Name)
		}
		dst = ctx.String(dstFlag.Name)
		if dst == "" {
			return fmt.Errorf("flag %q is necessary", dstFlag.Name)
		}
		overwrite = ctx.Bool(overwriteFlag.Name)

		return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return err
			}
			tgt, err := filepath.Rel(src, path)
			if err != nil {
				return fmt.Errorf("%q is not relative to %q", path, src)
			}
			_, err = os.Stat(filepath.Join(dst, tgt))
			switch {
			case os.IsNotExist(err):
			case err == nil:
				// the destination file exists
				if !overwrite {
					return fmt.Errorf("can't proceed, action would overwrite %q and %q flag was not provided",
						filepath.Join(dst, tgt), overwriteFlag.Name)
				}
				if D {
					log.Printf("will overwrite %q", filepath.Join(dst, tgt))
				}
			default:
				return fmt.Errorf("unknown error: %v", err)
			}
			files = append(files, struct{ src, dst string }{
				src: filepath.Join(src, tgt),
				dst: filepath.Join(dst, tgt),
			})
			return nil
		})
	}

	action := func(ctx *cli.Context) {

		wg := new(sync.WaitGroup)
		sem := make(chan struct{}, runtime.NumCPU())

		for _, pair := range files {
			wg.Add(1)
			go func(src, dst string) {
				sem <- struct{}{}
				defer func() { wg.Done(); <-sem }()

				if D {
					log.Printf("rendering %q", dst)
				}

				srcDirFi, err := os.Stat(filepath.Dir(src))
				if err != nil {
					log.Fatalf("can't stat dir of %q: %v", src, err)
				}
				if _, err := os.Stat(filepath.Dir(dst)); os.IsNotExist(err) {
					err := os.MkdirAll(filepath.Dir(dst), srcDirFi.Mode().Perm())
					if err != nil {
						log.Fatalf("can't mkdir of %q: %v", dst, err)
					}
				} else if err != nil {
					log.Fatalf("can't stat dir of %q: %v", dst, err)
				}

				fi, err := os.Stat(src)
				if err != nil {
					log.Fatalf("can't stat %q: %v", src, err)
				}
				templateText, err := ioutil.ReadFile(src)
				if err != nil {
					log.Fatalf("can't read content of %q: %v", src, err)
				}

				tmpl, err := template.New("").Parse(string(templateText))
				if err != nil {
					log.Fatalf("can't parse template %q: %v", src, err)
				}

				flags := os.O_CREATE | os.O_WRONLY
				if overwrite {
					flags |= os.O_TRUNC
				}
				fd, err := os.OpenFile(dst, flags, fi.Mode())
				if err != nil {
					log.Fatalf("can't open file %q: %v", dst, err)
				}
				if err := tmpl.Execute(fd, vars); err != nil {
					log.Fatalf("can't render template on file %q: %v", dst, err)
				}
				if err := fd.Close(); err != nil {
					log.Fatalf("can't close file %q: %v", dst, err)
				}
			}(pair.src, pair.dst)
		}
		wg.Wait()
		if D {
			log.Printf("rendered %d files", len(files))
		}
	}
	return cli.Command{
		Flags:       []cli.Flag{varFlag, srcFlag, dstFlag, overwriteFlag},
		Name:        "tree",
		Usage:       "render a tree of files",
		Description: strings.TrimSpace(``),
		Before:      checkFlags,
		Action:      action,
	}
}
