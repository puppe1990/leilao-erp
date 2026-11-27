// Package b2 talks to the Backblaze B2 Native API (v2): authorize, upload and
// delete. Tokens live in memory. Public shop URLs use the allPublic download host.
package b2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const defaultEndpoint = "https://api.backblazeb2.com"

const (
	putAttempts    = 3
	putBackoffBase = 250 * time.Millisecond
)

// Config is one application key scoped to a single bucket.
type Config struct {
	KeyID      string
	AppKey     string
	Bucket     string
	Endpoint   string
	PublicBase string
	HTTP       *http.Client
}

// Client uploads product media to a public B2 bucket.
type Client struct {
	cfg  Config
	http *http.Client

	mu   sync.Mutex
	auth *authSession
}

type authSession struct {
	accountID   string
	apiURL      string
	downloadURL string
	token       string
	bucketID    string
	bucketName  string
}

func New(cfg Config) *Client {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	httpClient := cfg.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Client{cfg: cfg, http: httpClient}
}

// FromEnv builds a client from B2_APPLICATION_KEY_ID, B2_APPLICATION_KEY and
// B2_BUCKET. All three empty disables B2 (nil, nil). A partial set is a boot
// error so a typo cannot silently keep uploads on disk.
func FromEnv() (*Client, error) {
	id := strings.TrimSpace(os.Getenv("B2_APPLICATION_KEY_ID"))
	key := strings.TrimSpace(os.Getenv("B2_APPLICATION_KEY"))
	bucket := strings.TrimSpace(os.Getenv("B2_BUCKET"))
	switch {
	case id == "" && key == "" && bucket == "":
		return nil, nil
	case id == "" || key == "" || bucket == "":
		return nil, fmt.Errorf("B2_APPLICATION_KEY_ID, B2_APPLICATION_KEY and B2_BUCKET must all be set")
	}
	return New(Config{
		KeyID:      id,
		AppKey:     key,
		Bucket:     bucket,
		Endpoint:   strings.TrimSpace(os.Getenv("B2_ENDPOINT")),
		PublicBase: publicBaseFromEnv(bucket),
	}), nil
}

func publicBaseFromEnv(bucket string) string {
	if base := strings.TrimRight(strings.TrimSpace(os.Getenv("B2_PUBLIC_BASE")), "/"); base != "" {
		return base
	}
	region := strings.TrimSpace(os.Getenv("B2_REGION"))
	suffix := strings.TrimPrefix(region, "us-east-")
	if suffix == "" || suffix == region {
		suffix = "005"
	}
	return fmt.Sprintf("https://f%s.backblazeb2.com/file/%s", suffix, bucket)
}

func (c *Client) PublicURL(key string) string {
	base := strings.TrimRight(c.cfg.PublicBase, "/")
	if base == "" {
		base = fmt.Sprintf("https://f005.backblazeb2.com/file/%s", c.cfg.Bucket)
	}
	return base + "/" + strings.TrimPrefix(key, "/")
}

// ObjectKeyFromURL maps a local /static/uploads path or a public B2 URL back
// to the object key (products/{id}/{file}).
func ObjectKeyFromURL(raw, bucket string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "/static/uploads/") {
		key := strings.TrimPrefix(s, "/static/uploads/")
		if key == "" || strings.Contains(key, "..") {
			return "", false
		}
		return key, true
	}
	const fileMarker = "/file/"
	i := strings.Index(s, fileMarker)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(fileMarker):]
	if q := strings.IndexAny(rest, "?#"); q >= 0 {
		rest = rest[:q]
	}
	if bucket != "" {
		prefix := bucket + "/"
		if !strings.HasPrefix(rest, prefix) {
			return "", false
		}
		rest = rest[len(prefix):]
	} else {
		slash := strings.Index(rest, "/")
		if slash < 0 {
			return "", false
		}
		rest = rest[slash+1:]
	}
	if rest == "" {
		return "", false
	}
	return rest, true
}

func (c *Client) Put(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	var lastErr error
	backoff := putBackoffBase
	for attempt := 1; attempt <= putAttempts; attempt++ {
		lastErr = c.putOnce(ctx, key, body, contentType)
		if lastErr == nil {
			return c.PublicURL(key), nil
		}
		if !isRetryableB2Error(lastErr) || attempt == putAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 3
	}
	return "", lastErr
}

// putOnce fetches a fresh upload URL and posts the file. Retrying with a new
// URL matters: a 503 "no tomes available" means that pod is out of capacity.
func (c *Client) putOnce(ctx context.Context, key string, body []byte, contentType string) error {
	sess, err := c.session(ctx)
	if err != nil {
		return err
	}
	var upload struct {
		UploadURL          string `json:"uploadUrl"`
		AuthorizationToken string `json:"authorizationToken"`
	}
	if err := c.post(ctx, "/b2_get_upload_url", map[string]any{"bucketId": sess.bucketID}, &upload); err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upload.UploadURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("b2 upload: %w", err)
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("Authorization", upload.AuthorizationToken)
	req.Header.Set("X-Bz-File-Name", encodeFileName(key))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Bz-Content-Sha1", "do_not_verify")
	return c.do(req, nil)
}

func (c *Client) Remove(ctx context.Context, key string) error {
	file, err := c.findFile(ctx, key)
	if err != nil {
		return err
	}
	return c.post(ctx, "/b2_delete_file_version", map[string]any{
		"fileName": file.FileName, "fileId": file.FileID,
	}, nil)
}

type b2File struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
}

func (c *Client) findFile(ctx context.Context, key string) (b2File, error) {
	sess, err := c.session(ctx)
	if err != nil {
		return b2File{}, err
	}
	var payload struct {
		Files []b2File `json:"files"`
	}
	if err := c.post(ctx, "/b2_list_file_names", map[string]any{
		"bucketId": sess.bucketID, "prefix": key, "maxFileCount": 1,
	}, &payload); err != nil {
		return b2File{}, err
	}
	for _, file := range payload.Files {
		if file.FileName == key {
			return file, nil
		}
	}
	return b2File{}, fmt.Errorf("b2 object %q not found", key)
}

func (c *Client) session(ctx context.Context) (*authSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.auth != nil {
		return c.auth, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.cfg.Endpoint, "/")+"/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		return nil, fmt.Errorf("b2 authorize: %w", err)
	}
	req.SetBasicAuth(c.cfg.KeyID, c.cfg.AppKey)

	var payload struct {
		AccountID          string `json:"accountId"`
		AuthorizationToken string `json:"authorizationToken"`
		APIURL             string `json:"apiUrl"`
		DownloadURL        string `json:"downloadUrl"`
		Allowed            struct {
			BucketID   string `json:"bucketId"`
			BucketName string `json:"bucketName"`
		} `json:"allowed"`
	}
	if err := c.do(req, &payload); err != nil {
		return nil, err
	}
	if payload.AuthorizationToken == "" || payload.APIURL == "" {
		return nil, fmt.Errorf("b2 authorize: empty response")
	}
	bucketID := payload.Allowed.BucketID
	bucketName := payload.Allowed.BucketName
	if bucketName == "" {
		bucketName = c.cfg.Bucket
	}
	if bucketID == "" {
		id, err := c.lookupBucketID(ctx, payload.APIURL, payload.AuthorizationToken, payload.AccountID)
		if err != nil {
			return nil, err
		}
		bucketID = id
	}
	c.auth = &authSession{
		accountID:   payload.AccountID,
		apiURL:      strings.TrimRight(payload.APIURL, "/"),
		downloadURL: strings.TrimRight(payload.DownloadURL, "/"),
		token:       payload.AuthorizationToken,
		bucketID:    bucketID,
		bucketName:  bucketName,
	}
	return c.auth, nil
}

func (c *Client) lookupBucketID(ctx context.Context, apiURL, token, accountID string) (string, error) {
	body, err := json.Marshal(map[string]any{"accountId": accountID, "bucketName": c.cfg.Bucket})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(apiURL, "/")+"/b2api/v2/b2_list_buckets", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	var payload struct {
		Buckets []struct {
			BucketID   string `json:"bucketId"`
			BucketName string `json:"bucketName"`
		} `json:"buckets"`
	}
	if err := c.do(req, &payload); err != nil {
		return "", err
	}
	for _, b := range payload.Buckets {
		if b.BucketName == c.cfg.Bucket {
			return b.BucketID, nil
		}
	}
	return "", fmt.Errorf("b2 bucket %q not found", c.cfg.Bucket)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	sess, err := c.session(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("b2 %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sess.apiURL+"/b2api/v2"+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("b2 %s: %w", path, err)
	}
	req.Header.Set("Authorization", sess.token)
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("b2 %s: %w", req.URL.Path, err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return &statusError{path: req.URL.Path, code: res.StatusCode, snippet: snippet}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// statusError keeps the HTTP status so callers can tell transient failures
// (503 no tomes available) from permanent ones (400 bad request).
type statusError struct {
	path    string
	code    int
	snippet []byte
}

func (e *statusError) Error() string {
	return fmt.Sprintf("b2 %s: HTTP %d: %s", e.path, e.code, bytes.TrimSpace(e.snippet))
}

func isRetryableB2Error(err error) bool {
	var se *statusError
	if !errors.As(err, &se) {
		return false
	}
	switch se.code {
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return true
	}
	return se.code >= 500
}

func encodeFileName(key string) string {
	escaped := strings.ReplaceAll(url.QueryEscape(key), "+", "%20")
	return strings.ReplaceAll(escaped, "%2F", "/")
}
