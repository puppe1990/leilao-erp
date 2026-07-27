package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// parseFormOrJSON populates r.Form from application/x-www-form-urlencoded
// or multipart, or from a JSON body (Inertia useForm default).
// After success, FormValue works the same for both clients.
func parseFormOrJSON(r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return err
		}
		_ = r.Body.Close()
		if len(body) == 0 {
			r.Form = make(url.Values)
			return nil
		}
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err != nil {
			return err
		}
		r.Form = make(url.Values, len(raw))
		r.PostForm = r.Form
		for k, v := range raw {
			r.Form.Set(k, jsonScalar(v))
		}
		return nil
	}
	return r.ParseForm()
}

func jsonScalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers decode as float64; keep integers clean when whole.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
