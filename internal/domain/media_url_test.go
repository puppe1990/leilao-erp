package domain

import "testing"

func TestCatalogThumbURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/static/uploads/products/1/a.jpg", "/static/uploads/products/1/a.thumb.jpg"},
		{"/static/uploads/products/1/a.jpeg", "/static/uploads/products/1/a.thumb.jpg"},
		{"/static/uploads/products/1/a.thumb.jpg", "/static/uploads/products/1/a.thumb.jpg"},
		{"https://cdn.example/a.jpg", "https://cdn.example/a.jpg"},
		{"", ""},
		{"/static/uploads/products/1/a.mp4", "/static/uploads/products/1/a.mp4"},
	}
	for _, tc := range cases {
		if got := CatalogThumbURL(tc.in); got != tc.want {
			t.Errorf("CatalogThumbURL(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
