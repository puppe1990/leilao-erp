package handlers

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func TestShopBrandFromName_accessoriesAreNotMonitor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want string
	}{
		{"Carregador 30W com cabo Tipo C", "CARREGADOR"},
		{"Carregador 30W com cabo iPhone", "CARREGADOR"},
		{"Carregador completo Tipo C", "CARREGADOR"},
		{"Fonte 30W USB Tipo C", "CARREGADOR"},
		{"Fonte rápida USB", "CARREGADOR"},
		{"Cabo turbo USB para Tipo C 66W", "CABO"},
		{"Cabo Kaidi Tipo C para Tipo C", "CABO"},
		{"Monitor Dell P2016t 19,5\"", "DELL"},
		{"Samsung Galaxy S9+ preto", "SAMSUNG"},
	}
	for _, tc := range cases {
		if got := shopBrandFromName(tc.name); got != tc.want {
			t.Errorf("shopBrandFromName(%q)=%q want %q", tc.name, got, tc.want)
		}
	}
}

func TestShopCategory_accessoriesAreNotComBase(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want string
	}{
		{"Carregador 30W com cabo Tipo C", "acessorio"},
		{"Fonte 30W USB Tipo C", "acessorio"},
		{"Fonte rápida USB", "acessorio"},
		{"Cabo turbo USB para iPhone 66W", "acessorio"},
		{"Monitor Dell P2016t 19,5\"", "com-base"},
		{"Monitor LG E1941S-PN 18,5\" (sem base)", "sem-base"},
		{"Monitor Dell P2219H 22\" (defeito parte inferior da tela)", "defeito"},
	}
	for _, tc := range cases {
		got := shopCategory(models.Product{Name: tc.name})
		if got != tc.want {
			t.Errorf("shopCategory(%q)=%q want %q", tc.name, got, tc.want)
		}
	}
}
