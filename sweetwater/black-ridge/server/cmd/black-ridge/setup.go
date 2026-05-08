package main

import (
	"context"
	"io/fs"
	"sync"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
	"github.com/charviki/sweetwater-black-ridge/internal/transport"
	"github.com/charviki/sweetwater-black-ridge/internal/webstatic"
	"net/http"
)

var staticFiles fs.FS = webstatic.Files

func newHTTPServer(cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger, gwmux *gwruntime.ServeMux) (*http.Server, *service.TemplateStore) {
	return transport.NewHTTPServer(transport.HTTPHandlerParams{
		Config:         cfg,
		TmuxService:    tmuxService,
		Logger:         logger,
		GWMux:          gwmux,
		StaticFiles:    staticFiles,
		AuthToken:      cfg.Server.AuthToken,
		AllowedOrigins: cfg.Server.AllowedOrigins,
	})
}

type backgroundRunner struct {
	name     string
	logger   logutil.Logger
	run      func(<-chan struct{})
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func newBackgroundRunner(name string, logger logutil.Logger, run func(<-chan struct{})) *backgroundRunner {
	return &backgroundRunner{
		name:   name,
		logger: logger,
		run:    run,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (r *backgroundRunner) ListenAndServe() error {
	defer close(r.doneCh)
	r.run(r.stopCh)
	return nil
}

func (r *backgroundRunner) Shutdown(ctx context.Context) error {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	select {
	case <-r.doneCh:
		return nil
	case <-ctx.Done():
		if r.logger != nil {
			r.logger.Warnf("[%s] shutdown timed out: %v", r.name, ctx.Err())
		}
		return nil
	}
}
