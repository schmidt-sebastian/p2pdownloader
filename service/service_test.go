package service_test

import (
	"testing"

	"github.com/schmidt-sebastian/p2pdownloader/service"
	"v.io/v23"
	"v.io/v23/rpc"
	_ "v.io/x/ref/runtime/factories/fake"
)

func TestGetChunk(t *testing.T) {
	var serverCall rpc.ServerCall
	ctx, shutdown := v23.Init()
	defer shutdown()
	p2pdownloader := service.Make(ctx, "tmp/downloader/test", "tmp/downloader")
	p2pdownloader.AddDownload("http://www.example.com")
	chunk, err := p2pdownloader.GetChunk(ctx, serverCall)

	if err != nil {
		t.Errorf("GetChunk returned error")
	}

	if chunk.Url != "http://www.example.com" {
		t.Errorf("GetChunk does not reference the correct URL")
	}

	// Try to get a second chunk, this should fail since example.com is small
	chunk, err = p2pdownloader.GetChunk(ctx, serverCall)

	if err == nil {
		t.Errorf("GetChunk returned a second chunk")
	}
}
