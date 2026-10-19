package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var storageHTTP = &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
var bucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func validateStorageConfig(c Config) error {
	if c.StorageMode != "database" && c.StorageMode != "supabase" {
		return errors.New("STORAGE_MODE must be database or supabase")
	}
	if c.StorageMode == "database" && c.StorageKey == "" {
		return nil
	}
	u, e := url.Parse(c.SupabaseURL)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("Storage requires SUPABASE_URL as an HTTPS origin")
	}
	if !bucketName.MatchString(c.StorageBucket) {
		return errors.New("SUPABASE_STORAGE_BUCKET must be a lowercase bucket name")
	}
	if c.StorageKey == "" || strings.HasPrefix(c.StorageKey, "sb_publishable_") || c.StorageKey == c.SupabaseKey {
		return errors.New("Set a separate server-only SUPABASE_STORAGE_KEY (secret or legacy service_role key)")
	}
	return nil
}

// Privileged storage credentials never reach the browser. Callers must authorize
// project membership before invoking these helpers, including for downloads.
func (a *App) storageRequest(ctx context.Context, method, path string, data []byte) (*http.Response, error) {
	if a.cfg.StorageKey == "" {
		return nil, errors.New("storage credentials missing")
	}
	req, e := http.NewRequestWithContext(ctx, method, a.cfg.SupabaseURL+"/storage/v1/"+path, bytes.NewReader(data))
	if e != nil {
		return nil, e
	}
	req.Header.Set("apikey", a.cfg.StorageKey)
	if !strings.HasPrefix(a.cfg.StorageKey, "sb_secret_") {
		req.Header.Set("Authorization", "Bearer "+a.cfg.StorageKey)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if method == "DELETE" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("x-upsert", "false")
	res, e := storageHTTP.Do(req)
	if e != nil {
		return nil, errors.New("storage transport failed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		res.Body.Close()
		return nil, fmt.Errorf("storage returned HTTP %d", res.StatusCode)
	}
	return res, nil
}
func storageObjectPath(bucket, key string) (string, error) {
	if !bucketName.MatchString(bucket) || !strings.HasPrefix(key, "projects/") {
		return "", errors.New("invalid storage path")
	}
	parts := strings.Split(key, "/")
	for i, p := range parts {
		if p == "" || p == "." || p == ".." {
			return "", errors.New("invalid storage path")
		}
		parts[i] = url.PathEscape(p)
	}
	return bucket + "/" + strings.Join(parts, "/"), nil
}
func (a *App) storagePut(ctx context.Context, bucket, key string, data []byte) error {
	p, e := storageObjectPath(bucket, key)
	if e != nil {
		return e
	}
	res, e := a.storageRequest(ctx, "POST", "object/"+p, data)
	if e != nil {
		return e
	}
	res.Body.Close()
	return nil
}
func (a *App) storageGet(ctx context.Context, bucket, key string, size int64) ([]byte, error) {
	if size < 1 || size > 8<<20 {
		return nil, errors.New("invalid stored size")
	}
	p, e := storageObjectPath(bucket, key)
	if e != nil {
		return nil, e
	}
	res, e := a.storageRequest(ctx, "GET", "object/authenticated/"+p, nil)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	data, e := io.ReadAll(io.LimitReader(res.Body, size+1))
	if e != nil || int64(len(data)) != size {
		return nil, errors.New("incomplete storage download")
	}
	return data, nil
}
func (a *App) storageDelete(ctx context.Context, bucket, key string) error {
	if _, e := storageObjectPath(bucket, key); e != nil {
		return e
	}
	body, _ := json.Marshal(map[string]any{"prefixes": []string{key}})
	res, e := a.storageRequest(ctx, "DELETE", "object/"+bucket, body)
	if e != nil {
		return e
	}
	res.Body.Close()
	return nil
}
