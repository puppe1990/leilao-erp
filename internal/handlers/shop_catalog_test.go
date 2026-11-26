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
		{"CB-23 Carregador 30W H'Maston com cabo Tipo C", "CARREGADOR"},
		{"CA31-2 Carregador 30W com cabo iPhone", "CARREGADOR"},
		{"HC-61 Carregador USB-A e Tipo C", "CARREGADOR"},
		{"CB-16 Fonte 30W H'Maston Tipo C", "FONTE"},
		{"CA-25 Fonte rápida 2x USB-A", "FONTE"},
		{"CB-17 Cabo turbo USB para Tipo C 66W", "CABO"},
		{"KD-27 Cabo Kaidi Tipo C para Tipo C", "CABO"},
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
