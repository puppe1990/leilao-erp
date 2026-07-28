package domain

import "testing"

func TestNormalizeProductMediaURL_PhotoPathAndVideoURL(t *testing.T) {
	u, err := NormalizeProductMediaURL("photo", "/static/uploads/products/1/a.jpg")
	if err != nil || u != "/static/uploads/products/1/a.jpg" {
		t.Fatalf("photo path: %q %v", u, err)
	}
	u, err = NormalizeProductMediaURL("video", "https://www.youtube.com/watch?v=abc123")
	if err != nil || u == "" {
		t.Fatalf("youtube: %q %v", u, err)
	}
	u, err = NormalizeProductMediaURL("video", "https://youtu.be/abc123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeProductMediaURL("photo", "javascript:alert(1)"); err == nil {
		t.Fatal("want reject javascript")
	}
	if _, err := NormalizeProductMediaURL("video", "ftp://x"); err == nil {
		t.Fatal("want reject ftp")
	}
	if _, err := NormalizeProductMediaURL("audio", "/x"); err == nil {
		t.Fatal("want reject audio kind")
	}
	if _, err := NormalizeProductMediaURL("photo", ""); err == nil {
		t.Fatal("want reject empty")
	}
}

func TestIsProductMediaKind(t *testing.T) {
	if !IsProductMediaKind("photo") || !IsProductMediaKind("VIDEO") {
		t.Fatal("photo/video should be valid")
	}
	if IsProductMediaKind("gif") {
		t.Fatal("gif not a media kind")
	}
}
