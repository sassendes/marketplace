package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"mercatus/catalog/model"
)

type SearchIndex interface {
	Search(ctx context.Context, query string, limit int) ([]int64, error)
	Replace(ctx context.Context, docs []model.SearchDocument) error
	Upsert(ctx context.Context, doc model.SearchDocument) error
	Delete(ctx context.Context, id int64) error
}

type HTTPSearchIndex struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewHTTPSearchIndex(baseURL, token string) *HTTPSearchIndex {
	return &HTTPSearchIndex{baseURL: baseURL, token: token, client: &http.Client{Timeout: 3 * time.Second}}
}

func (s *HTTPSearchIndex) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("search engine %s: status %d", path, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (s *HTTPSearchIndex) Search(ctx context.Context, query string, limit int) ([]int64, error) {
	var resp struct {
		TotalIndexed int `json:"total_indexed"`
		Hits         []struct {
			ID int64 `json:"id"`
		} `json:"hits"`
	}
	if err := s.post(ctx, "/search", map[string]any{"q": query, "limit": limit}, &resp); err != nil {
		return nil, err
	}
	if resp.TotalIndexed == 0 {
		return nil, errors.New("search index is empty")
	}
	ids := make([]int64, 0, len(resp.Hits))
	for _, h := range resp.Hits {
		ids = append(ids, h.ID)
	}
	return ids, nil
}

func (s *HTTPSearchIndex) Replace(ctx context.Context, docs []model.SearchDocument) error {
	if docs == nil {
		docs = []model.SearchDocument{}
	}
	return s.post(ctx, "/index/replace", map[string]any{"documents": docs}, nil)
}

func (s *HTTPSearchIndex) Upsert(ctx context.Context, doc model.SearchDocument) error {
	return s.post(ctx, "/index", doc, nil)
}

func (s *HTTPSearchIndex) Delete(ctx context.Context, id int64) error {
	return s.post(ctx, "/index/delete", map[string]any{"id": id}, nil)
}
