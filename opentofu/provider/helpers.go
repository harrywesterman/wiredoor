package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "not found:")
}

func parseID(id string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(id), 10, 64)
}

func parseCompositeID(id string) (int64, int64, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected <node_id>/<resource_id>")
	}
	nodeID, err := parseID(parts[0])
	if err != nil {
		return 0, 0, err
	}
	resourceID, err := parseID(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return nodeID, resourceID, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}
