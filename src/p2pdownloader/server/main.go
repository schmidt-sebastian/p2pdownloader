package main

import (
	"p2pdownloader/service"
	"flag"
	"fmt"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/security"
	"v.io/x/ref/lib/signals"
	_ "v.io/x/ref/runtime/factories/roaming"
	"github.com/pborman/uuid"
	"p2pdownloader/protocol"
)

var (
	newDownload = flag.String("download", "", "HTTP location of new download")
	mountableRoot = flag.String("mountable", "tmp/downloader", "Global mountable identifier")
)

func runDownloadServer(ctx *context.T) error {
	localEndpoint := fmt.Sprintf("%v/%v", *mountableRoot, uuid.New())
	logic := service.Make(ctx, localEndpoint, *mountableRoot)

	ctx, _, err := v23.WithNewServer(ctx, localEndpoint,
		downloader.DownloaderServer(logic), security.AllowEveryone())

	if err != nil {
		ctx.Errorf("v23.WithNewServer failed with errro: %v", err)
		return nil
	}

	logic.Start()
	ctx.Infof("Successfully started server at %v", localEndpoint)

	if *newDownload != "" {
		logic.AddDownload(*newDownload)
	}

	<-signals.ShutdownOnSignals(ctx)
	logic.Shutdown()

	return nil
}

func main() {
	ctx, shutdown := v23.Init()
	defer shutdown()
	runDownloadServer(ctx)
}
