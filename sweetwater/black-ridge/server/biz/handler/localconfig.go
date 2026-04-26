package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

type LocalConfigHandler struct {
	store *service.LocalConfigStore
}

func NewLocalConfigHandler(store *service.LocalConfigStore) *LocalConfigHandler {
	return &LocalConfigHandler{store: store}
}

func (h *LocalConfigHandler) GetConfig(_ context.Context, c *app.RequestContext) {
	cfg := h.store.Get()
	httputil.Success(c, cfg)
}

func (h *LocalConfigHandler) UpdateConfig(_ context.Context, c *app.RequestContext) {
	var req protocol.LocalAgentConfig
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	current := h.store.Get()
	if req.WorkingDir != "" && req.WorkingDir != current.WorkingDir {
		httputil.Error(c, http.StatusBadRequest, "working_dir is read-only")
		return
	}

	if req.Env != nil {
		if err := h.store.UpdateEnv(req.Env); err != nil {
			httputil.Error(c, http.StatusInternalServerError, "failed to update env: "+err.Error())
			return
		}
	}

	httputil.Success(c, h.store.Get())
}
