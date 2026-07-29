package handlers

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/session"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/store"
)

const minPasswordLen = 8

type ConfigHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewConfigHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ConfigHandler {
	return &ConfigHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *ConfigHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.store.FindUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name, err := h.store.CompanyName()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	waPhone, err := h.store.WhatsAppPhone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flash := ""
	switch r.URL.Query().Get("saved") {
	case "company":
		flash = "Nome da empresa atualizado."
	case "password":
		flash = "Senha alterada com sucesso."
	case "whatsapp":
		flash = "WhatsApp da loja atualizado."
	}

	_ = h.inertia.Render(w, r, "Config/Index", withCompany(h.store, inertia.Props{
		"site":          meta.ForRequest(h.site, r),
		"email":         user.Email,
		"companyName":   companyName(h.store),
		"companyForm":   name, // raw stored value (may be empty)
		"whatsappPhone": waPhone,
		"shopURL":       "/loja",
		"flash":         flash,
	}))
}

func (h *ConfigHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("company_name"))
	if name == "" {
		h.renderConfigErrors(w, r, inertia.ValidationErrors{
			"company_name": "Informe o nome da empresa",
		}, "")
		return
	}
	if utf8.RuneCountInString(name) > 80 {
		h.renderConfigErrors(w, r, inertia.ValidationErrors{
			"company_name": "Nome deve ter no máximo 80 caracteres",
		}, "")
		return
	}

	if err := h.store.SetCompanyName(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/config?saved=company", http.StatusSeeOther)
}

func (h *ConfigHandler) UpdateWhatsApp(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	phone := strings.TrimSpace(r.FormValue("whatsapp_phone"))
	if err := h.store.SetWhatsAppPhone(phone); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/config?saved=whatsapp", http.StatusSeeOther)
}

func (h *ConfigHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	current := r.FormValue("current_password")
	newPass := r.FormValue("new_password")
	confirm := r.FormValue("new_password_confirmation")

	errs := inertia.ValidationErrors{}
	if current == "" {
		errs["current_password"] = "Informe a senha atual"
	}
	if utf8.RuneCountInString(newPass) < minPasswordLen {
		errs["new_password"] = "Nova senha deve ter no mínimo 8 caracteres"
	}
	if newPass != confirm {
		errs["new_password_confirmation"] = "Confirmação não confere"
	}
	if len(errs) > 0 {
		h.renderConfigErrors(w, r, errs, "")
		return
	}

	user, err := h.store.FindUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !session.VerifyPassword(user.PasswordHash, current) {
		h.renderConfigErrors(w, r, inertia.ValidationErrors{
			"current_password": "Senha atual incorreta",
		}, "")
		return
	}

	hash, err := session.HashPassword(newPass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.UpdateUserPassword(userID, hash); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/config?saved=password", http.StatusSeeOther)
}

func (h *ConfigHandler) renderConfigErrors(w http.ResponseWriter, r *http.Request, errs inertia.ValidationErrors, flash string) {
	userID, _ := session.UserID(r)
	email := ""
	if user, err := h.store.FindUserByID(userID); err == nil {
		email = user.Email
	}
	rawName, _ := h.store.CompanyName()
	// Prefer submitted company name on validation errors for company form
	if v := strings.TrimSpace(r.FormValue("company_name")); v != "" {
		rawName = v
	}

	ctx := inertia.SetValidationErrors(r.Context(), errs)
	_ = h.inertia.Render(w, r.WithContext(ctx), "Config/Index", withCompany(h.store, inertia.Props{
		"site":        meta.ForRequest(h.site, r),
		"email":       email,
		"companyName": companyName(h.store),
		"companyForm": rawName,
		"flash":       flash,
	}))
}
