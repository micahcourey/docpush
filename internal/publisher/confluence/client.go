package confluence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Client is a Confluence Data Center REST API client.
type Client struct {
	baseURL    string
	pat        string
	httpClient *http.Client
}

// NewClient creates a new Confluence REST API client.
// baseURL should be like "https://your-domain.atlassian.net" or "https://confluence.example.com".
// pat is a Personal Access Token for authentication.
func NewClient(baseURL, pat string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		pat:        pat,
		httpClient: &http.Client{},
	}
}

// PageResponse represents a Confluence page from the REST API.
type PageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Body struct {
		Storage struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
	Links struct {
		Base   string `json:"base"`
		WebUI  string `json:"webui"`
		TinyUI string `json:"tinyui"`
	} `json:"_links"`
}

// SearchResponse represents a Confluence content search result.
type SearchResponse struct {
	Results []PageResponse `json:"results"`
	Size    int            `json:"size"`
}

// GetPage retrieves a Confluence page by ID with body and version expanded.
func (c *Client) GetPage(ctx context.Context, pageID string) (*PageResponse, error) {
	url := fmt.Sprintf("%s/rest/api/content/%s?expand=version,body.storage", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET page %s: status %d: %s", pageID, resp.StatusCode, string(body))
	}

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &page, nil
}

// CreatePage creates a new page under the given parent in the given space.
func (c *Client) CreatePage(ctx context.Context, spaceKey, parentID, title, body string) (*PageResponse, error) {
	payload := map[string]interface{}{
		"type":  "page",
		"title": title,
		"space": map[string]string{"key": spaceKey},
		"body": map[string]interface{}{
			"storage": map[string]string{
				"value":          body,
				"representation": "storage",
			},
		},
	}
	if parentID != "" {
		payload["ancestors"] = []map[string]string{{"id": parentID}}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/content", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST create page: status %d: %s", resp.StatusCode, string(respBody))
	}

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &page, nil
}

// UpdatePage updates an existing page. The version number must be incremented.
func (c *Client) UpdatePage(ctx context.Context, pageID, title, body string, version int) (*PageResponse, error) {
	payload := map[string]interface{}{
		"type":  "page",
		"title": title,
		"body": map[string]interface{}{
			"storage": map[string]string{
				"value":          body,
				"representation": "storage",
			},
		},
		"version": map[string]int{"number": version},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/content/%s", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PUT update page %s: status %d: %s", pageID, resp.StatusCode, string(respBody))
	}

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &page, nil
}

// AddLabels adds labels to a page.
func (c *Client) AddLabels(ctx context.Context, pageID string, labels []string) error {
	var payload []map[string]string
	for _, l := range labels {
		payload = append(payload, map[string]string{
			"prefix": "global",
			"name":   l,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling labels: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/label", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST labels page %s: status %d: %s", pageID, resp.StatusCode, string(body))
	}
	return nil
}

// SearchByTitle searches for a page by title within a space.
func (c *Client) SearchByTitle(ctx context.Context, spaceKey, title string) (*PageResponse, error) {
	url := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&title=%s&expand=version,body.storage",
		c.baseURL, spaceKey, title)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET search: status %d: %s", resp.StatusCode, string(body))
	}

	var sr SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if sr.Size == 0 {
		return nil, nil
	}
	return &sr.Results[0], nil
}

// Validate checks that the client can reach the Confluence API.
func (c *Client) Validate(ctx context.Context) error {
	url := fmt.Sprintf("%s/rest/api/space?limit=1", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reaching Confluence API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed (status %d) — check CONFLUENCE_PAT", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.pat)
}

// ChildPageResponse represents a paginated list of child pages.
type ChildPageResponse struct {
	Results []PageResponse `json:"results"`
	Size    int            `json:"size"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// GetChildPages retrieves all child pages under the given parent page ID.
// It follows pagination links to collect all results.
func (c *Client) GetChildPages(ctx context.Context, parentID string) ([]PageResponse, error) {
	var all []PageResponse
	path := fmt.Sprintf("/rest/api/content/%s/child/page?limit=50&expand=version", parentID)

	for path != "" {
		url := c.baseURL + path
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		c.setAuth(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GET child pages of %s: status %d: %s", parentID, resp.StatusCode, string(body))
		}

		var cr ChildPageResponse
		if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		all = append(all, cr.Results...)

		path = cr.Links.Next
	}

	return all, nil
}

// SetReadOnly restricts page editing to the authenticated user only.
// Other users can still view the page but cannot edit it.
func (c *Client) SetReadOnly(ctx context.Context, pageID string) error {
	payload := []map[string]interface{}{
		{
			"operation": "update",
			"restrictions": map[string]interface{}{
				"user": []map[string]string{
					{"type": "known", "username": "svc_iddoc_github"},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling restrictions: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/restriction", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PUT restrictions page %s: status %d: %s", pageID, resp.StatusCode, string(body))
	}
	return nil
}

// AttachmentResponse represents a Confluence attachment from the REST API.
type AttachmentResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Links struct {
		Download string `json:"download"`
	} `json:"_links"`
	Extensions struct {
		MediaType string `json:"mediaType"`
		FileSize  int64  `json:"fileSize"`
		Comment   string `json:"comment"`
	} `json:"extensions"`
}

// AttachmentListResponse represents a paginated list of attachments.
type AttachmentListResponse struct {
	Results []AttachmentResponse `json:"results"`
	Size    int                  `json:"size"`
}

// GetAttachments retrieves all attachments for a page.
func (c *Client) GetAttachments(ctx context.Context, pageID string) ([]AttachmentResponse, error) {
	url := fmt.Sprintf("%s/rest/api/content/%s/child/attachment?limit=100", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET attachments for page %s: status %d: %s", pageID, resp.StatusCode, string(body))
	}

	var result AttachmentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding attachments: %w", err)
	}
	return result.Results, nil
}

// UploadAttachment uploads a file as an attachment to a Confluence page.
// If an attachment with the same filename already exists, it updates it.
// The comment field is used to store a content hash for change detection.
func (c *Client) UploadAttachment(ctx context.Context, pageID, filePath string) (*AttachmentResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filePath, err)
	}
	defer file.Close()

	// Read file content for hash
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}
	hash := sha256.Sum256(fileBytes)
	hashStr := "sha256:" + hex.EncodeToString(hash[:])

	// Reset file reader
	file.Seek(0, io.SeekStart)

	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copying file data: %w", err)
	}

	// Add comment with hash for change detection
	if err := writer.WriteField("comment", hashStr); err != nil {
		return nil, fmt.Errorf("writing comment field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/child/attachment", c.baseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "nocheck")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PUT attachment to page %s: status %d: %s", pageID, resp.StatusCode, string(respBody))
	}

	var result AttachmentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding attachment response: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no attachment returned after upload")
	}
	return &result.Results[0], nil
}

// FileHash computes the SHA-256 hash of a file and returns it as "sha256:<hex>".
func FileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}
