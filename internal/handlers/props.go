package handlers

import (
	"strings"

	"github.com/puppe1990/leilao-erp/internal/store"
	inertia "github.com/romsar/gonertia/v3"
)

const defaultCompanyName = "AuctionHQ"

func companyName(s store.Store) string {
	name, err := s.CompanyName()
	if err != nil {
		return defaultCompanyName
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultCompanyName
	}
	return name
}

// withCompany merges companyName into Inertia props for AppShell branding.
func withCompany(s store.Store, props inertia.Props) inertia.Props {
	if props == nil {
		props = inertia.Props{}
	}
	out := make(inertia.Props, len(props)+1)
	for k, v := range props {
		out[k] = v
	}
	out["companyName"] = companyName(s)
	return out
}
