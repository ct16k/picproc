package main

import (
	"picproc/orient"

	"github.com/alecthomas/kong"
)

var cli struct {
	Orient orient.CLICmd `cmd:"" help:"Sort files by orientation"`
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
	kctx.Bind(kctx.Selected().Name)
	err := kctx.Run()
	kctx.FatalIfErrorf(err)
}
