package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"picproc/mangle"
	"picproc/orient"
	"picproc/parallel"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/alecthomas/kong"
)

var cli struct {
	Workers int           `help:"Number of concurrent workers (if less than 1 use number of CPUs)" default:"1"`
	Orient  orient.CLICmd `cmd:"" help:"Sort files by orientation"`
	Mangle  mangle.CLICmd `cmd:"" help:"Mangle an image"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description("picproc processes pictures"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	workerPool := parallel.Start(cli.Workers)

	kctx.Bind(kctx.Selected().Name, workerPool.Do, workerPool.Wait)
	err := kctx.Run()

	workerPool.Cancel()
	workerPool.Wait(true)
	kctx.FatalIfErrorf(err)
}
