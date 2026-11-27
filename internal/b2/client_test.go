package b2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestObjectKeyFromURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw    string
		bucket string
		key    string
		ok     bool
	}{
		{"/static/uploads/products/35/a.jpg", "eletronicos-phoenix", "products/35/a.jpg", true},
		{"https://f005.backblazeb2.com/file/eletronicos-phoenix/products/5/p2016t-front.jpg", "eletronicos-phoenix", "products/5/p2016t-front.jpg", true},
		{"https://f005.backblazeb2.com/file/eletronicos-phoenix/products/5/p2016t-front.jpg", "", "products/5/p2016t-front.jpg", true},
		{"https://www.youtube.com/watch?v=abc", "eletronicos-phoenix", "", false},
		{"", "eletronicos-phoenix", "", false},
	}
	for _, tc := range cases {
		key, ok := ObjectKeyFromURL(tc.raw, tc.bucket)
		if ok != tc.ok || key != tc.key {
			t.Errorf("ObjectKeyFromURL(%q) = %q,%v want %q,%v", tc.raw, key, ok, tc.key, tc.ok)
		}
	}
}

func TestPublicURL_usesPublicBase(t *testing.T) {
	t.Parallel()
	c := New(Config{
		Bucket:     "eletronicos-phoenix",
		PublicBase: "https://f005.backblazeb2.com/file/eletronicos-phoenix",
	})
	got := c.PublicURL("products/35/a.jpg")
	want := "https://f005.backblazeb2.com/file/eletronicos-phoenix/products/35/a.jpg"
	if got != want {
		t.Fatalf("PublicURL = %q, want %q", got, want)
	}
}

func TestFromEnv_disabledWhenUnset(t *testing.T) {
	t.Setenv("B2_APPLICATION_KEY_ID", "")
	t.Setenv("B2_APPLICATION_KEY", "")
	t.Setenv("B2_BUCKET", "")
	t.Setenv("B2_PUBLIC_BASE", "")
	t.Setenv("B2_REGION", "")
	client, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if client != nil {
		t.Fatal("unset env must return a nil client")
	}
}

func TestFromEnv_errorsWhenPartial(t *testing.T) {
	t.Setenv("B2_APPLICATION_KEY_ID", "id")
	t.Setenv("B2_APPLICATION_KEY", "")
	t.Setenv("B2_BUCKET", "eletronicos-phoenix")
	if _, err := FromEnv(); err == nil {
		t.Fatal("partial B2 env must error")
	}
}

func TestFromEnv_derivesPublicBaseFromRegion(t *testing.T) {
	t.Setenv("B2_APPLICATION_KEY_ID", "id")
	t.Setenv("B2_APPLICATION_KEY", "secret")
	t.Setenv("B2_BUCKET", "eletronicos-phoenix")
	t.Setenv("B2_REGION", "us-east-005")
	t.Setenv("B2_PUBLIC_BASE", "")
	client, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("full env must return a client")
	}
	got := client.PublicURL("products/1/a.jpg")
	want := "https://f005.backblazeb2.com/file/eletronicos-phoenix/products/1/a.jpg"
	if got != want {
		t.Fatalf("PublicURL = %q, want %q", got, want)
	}
}

func TestClient_PutRemove(t *testing.T) {
	srv := fakeB2(t)
	client := New(Config{
		KeyID:      "key-id",
		AppKey:     "key-secret",
		Bucket:     "eletronicos-phoenix",
		Endpoint:   srv.URL,
		PublicBase: srv.URL + "/file/eletronicos-phoenix",
		HTTP:       srv.Client(),
	})

	ctx := context.Background()
	body := []byte{0xff, 0xd8, 0xff, 0xe0}
	url, err := client.Put(ctx, "products/35/cam.jpg", body, "image/jpeg")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if srv.lastUpload.fileName != "products/35/cam.jpg" {
		t.Errorf("uploaded name = %q", srv.lastUpload.fileName)
	}
	if srv.lastUpload.contentType != "image/jpeg" {
		t.Errorf("uploaded type = %q", srv.lastUpload.contentType)
	}
	if !bytes.Equal(srv.lastUpload.body, body) {
		t.Errorf("uploaded body = %v", srv.lastUpload.body)
	}
	wantURL := srv.URL + "/file/eletronicos-phoenix/products/35/cam.jpg"
	if url != wantURL {
		t.Errorf("Put URL = %q, want %q", url, wantURL)
	}

	if err := client.Remove(ctx, "products/35/cam.jpg"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if srv.deletedFileID != "file-1" {
		t.Errorf("deleted file id = %q", srv.deletedFileID)
	}
}

func TestClient_PutRejectsBadAuth(t *testing.T) {
	srv := fakeB2(t)
	client := New(Config{
		KeyID: "wrong", AppKey: "nope", Bucket: "eletronicos-phoenix",
		Endpoint: srv.URL, HTTP: srv.Client(),
	})
	_, err := client.Put(context.Background(), "a.jpg", []byte{0xff, 0xd8, 0xff}, "image/jpeg")
	if err == nil {
		t.Fatal("Put with bad credentials must fail")
	}
}

type fakeB2Server struct {
	*httptest.Server
	files         map[string]storedFile
	lastUpload    storedFile
	deletedFileID string
}

type storedFile struct {
	fileName    string
	fileID      string
	contentType string
	body        []byte
}

func fakeB2(t *testing.T) *fakeB2Server {
	t.Helper()
	s := &fakeB2Server{files: map[string]storedFile{}}
	mux := http.NewServeMux()

	mux.HandleFunc("/b2api/v2/b2_authorize_account", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "key-id" || pass != "key-secret" {
			http.Error(w, `{"code":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accountId":          "acc-1",
			"authorizationToken": "token-1",
			"apiUrl":             s.URL,
			"downloadUrl":        s.URL,
			"allowed":            map[string]any{"bucketId": "bucket-1", "bucketName": "eletronicos-phoenix"},
		})
	})
	mux.HandleFunc("/b2api/v2/b2_get_upload_url", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token-1" {
			http.Error(w, `{"code":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uploadUrl":          s.URL + "/b2upload",
			"authorizationToken": "upload-token",
		})
	})
	mux.HandleFunc("/b2upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "upload-token" {
			http.Error(w, `{"code":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		name, _ := url.QueryUnescape(r.Header.Get("X-Bz-File-Name"))
		file := storedFile{
			fileName:    name,
			fileID:      "file-1",
			contentType: r.Header.Get("Content-Type"),
			body:        body,
		}
		s.lastUpload = file
		s.files[name] = file
		_ = json.NewEncoder(w).Encode(map[string]any{"fileId": "file-1", "fileName": name})
	})
	mux.HandleFunc("/b2api/v2/b2_list_file_names", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prefix string `json:"prefix"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		var files []map[string]any
		for name, file := range s.files {
			if strings.HasPrefix(name, req.Prefix) {
				files = append(files, map[string]any{
					"fileName": file.fileName, "fileId": file.fileID,
				})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"files": files})
	})
	mux.HandleFunc("/b2api/v2/b2_delete_file_version", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FileID   string `json:"fileId"`
			FileName string `json:"fileName"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		s.deletedFileID = req.FileID
		delete(s.files, req.FileName)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}
