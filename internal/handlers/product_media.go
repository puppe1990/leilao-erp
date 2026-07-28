package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

// AddMedia attaches a photo or video to a product.
// Form fields:
//   - kind: photo | video
//   - url: optional https:// or /static/... (used when no file)
//   - file: optional multipart upload (saved under /static/uploads/products/{id}/)
func (h *ProductsHandler) AddMedia(w http.ResponseWriter, r *http.Request, productID int64) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// fall back to urlencoded
		if err2 := parseFormOrJSON(r); err2 != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	kind := strings.TrimSpace(strings.ToLower(r.FormValue("kind")))
	if !domain.IsProductMediaKind(kind) {
		h.mediaError(w, r, productID, "Tipo deve ser foto ou vídeo")
		return
	}

	url := strings.TrimSpace(r.FormValue("url"))
	if file, hdr, err := r.FormFile("file"); err == nil {
		defer func() { _ = file.Close() }()
		saved, err := h.saveProductUpload(productID, kind, hdr.Filename, file)
		if err != nil {
			h.mediaError(w, r, productID, err.Error())
			return
		}
		url = saved
	}

	norm, err := domain.NormalizeProductMediaURL(kind, url)
	if err != nil {
		h.mediaError(w, r, productID, "URL/arquivo inválido: "+err.Error())
		return
	}

	if _, err := h.store.AddProductMedia(productID, store.ProductMediaInput{
		Kind: kind,
		URL:  norm,
	}); err != nil {
		h.mediaError(w, r, productID, err.Error())
		return
	}
	h.inertia.Redirect(w, r, fmt.Sprintf("/products/%d", productID), http.StatusSeeOther)
}

// DestroyMedia removes a media row (and local file if under uploads).
func (h *ProductsHandler) DestroyMedia(w http.ResponseWriter, r *http.Request, productID int64) {
	mediaID, _ := strconv.ParseInt(r.PathValue("mediaId"), 10, 64)
	if mediaID <= 0 {
		http.Error(w, "mídia inválida", http.StatusBadRequest)
		return
	}
	m, err := h.store.FindProductMedia(mediaID)
	if err != nil {
		h.mediaError(w, r, productID, "Mídia não encontrada")
		return
	}
	if m.ProductID != productID {
		h.mediaError(w, r, productID, "Mídia não pertence a este produto")
		return
	}
	if err := h.store.DeleteProductMedia(mediaID); err != nil {
		h.mediaError(w, r, productID, err.Error())
		return
	}
	h.tryRemoveLocalUpload(m.URL)
	h.inertia.Redirect(w, r, fmt.Sprintf("/products/%d", productID), http.StatusSeeOther)
}

func (h *ProductsHandler) mediaError(w http.ResponseWriter, r *http.Request, productID int64, msg string) {
	ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": msg})
	r = r.WithContext(ctx)
	h.Show(w, r, productID)
}

func (h *ProductsHandler) saveProductUpload(productID int64, kind, originalName string, src io.Reader) (string, error) {
	if h.staticDir == "" {
		return "", fmt.Errorf("upload indisponível (static dir vazio)")
	}
	ext := strings.ToLower(filepath.Ext(originalName))
	if kind == "photo" {
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		default:
			return "", fmt.Errorf("foto: use jpg, png, webp ou gif")
		}
	} else {
		switch ext {
		case ".mp4", ".webm", ".mov":
		default:
			// video uploads optional; still allow common video files
			if ext == "" {
				return "", fmt.Errorf("vídeo: envie arquivo com extensão ou use URL")
			}
		}
	}
	dir := filepath.Join(h.staticDir, "uploads", "products", strconv.FormatInt(productID, 10))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("criar pasta upload: %w", err)
	}
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), kind, ext)
	// sanitize name further
	name = strings.ReplaceAll(name, "..", "")
	destPath := filepath.Join(dir, name)
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = dst.Close() }()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return "/static/uploads/products/" + strconv.FormatInt(productID, 10) + "/" + name, nil
}

func (h *ProductsHandler) tryRemoveLocalUpload(mediaURL string) {
	if h.staticDir == "" || !strings.HasPrefix(mediaURL, "/static/uploads/") {
		return
	}
	rel := strings.TrimPrefix(mediaURL, "/static/")
	path := filepath.Join(h.staticDir, filepath.FromSlash(rel))
	// only delete inside uploads
	if !strings.Contains(filepath.ToSlash(path), "/uploads/") {
		return
	}
	_ = os.Remove(path)
}
