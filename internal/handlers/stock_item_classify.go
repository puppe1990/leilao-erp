package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

// isAccessoryTitle flags stock items typically sold with a monitor (cables).
// Only "cabo" is used — matching "hdmi"/"vga" alone wrongly tags monitors
// like "Prizi … PZ0018HDMI" as accessories.
func isAccessoryTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	return strings.HasPrefix(t, "cabo") || strings.Contains(t, " cabo")
}

// parseAccessoryIDs reads accessory_ids from form (repeated keys and/or CSV).
func parseAccessoryIDs(r *http.Request) []int64 {
	var out []int64
	seen := map[int64]bool{}
	for _, raw := range r.Form["accessory_ids"] {
		for _, part := range strings.Split(raw, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
			if err != nil || id <= 0 || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
