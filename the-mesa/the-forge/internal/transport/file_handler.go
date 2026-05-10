package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// FileHandler 处理文件上传下载。
type FileHandler struct {
	fileSvc *service.FileService
}

// NewFileHandler 创建 FileHandler。
func NewFileHandler(fileSvc *service.FileService) *FileHandler {
	return &FileHandler{fileSvc: fileSvc}
}

// RegisterRoutes 注册文件相关路由到 mux。
func (h *FileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/files/upload", h.handleUpload)
	mux.HandleFunc("GET /api/v1/files/{key}", h.handleDownload)
}

func (h *FileHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 限制上传大小 32MB
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)

	if err := r.ParseMultipartForm(32 << 20); err != nil { //nolint:gosec
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key, err := h.fileSvc.Upload(r.Context(), header.Filename, data, contentType)
	if err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
}

func (h *FileHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	data, contentType, err := h.fileSvc.Download(r.Context(), key)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, key))
	_, _ = w.Write(data) //nolint:gosec
}
