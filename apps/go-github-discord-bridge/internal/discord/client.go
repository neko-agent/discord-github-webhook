package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DiscordAPIBase = "https://discord.com/api/v10"
)

type Client struct {
	token          string
	forumChannelID string
	httpClient     *http.Client
}

// NewClient 建立 Discord API client
func NewClient(token, forumChannelID string) *Client {
	return &Client{
		token:          token,
		forumChannelID: forumChannelID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateThreadRequest 建立 thread 的請求結構
type CreateThreadRequest struct {
	Name    string        `json:"name"`    // Thread 標題
	Message ThreadMessage `json:"message"` // 第一則訊息
	// AppliedTags []string `json:"applied_tags,omitempty"` // Forum tags (可選)
}

type ThreadMessage struct {
	Content string  `json:"content,omitempty"` // 純文字內容
	Embeds  []Embed `json:"embeds,omitempty"`  // Rich embed
}

// Embed Discord 的 rich embed 結構
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Color       int          `json:"color,omitempty"` // 顏色（整數）
	Fields      []EmbedField `json:"fields,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"` // ISO 8601 format
	Footer      *EmbedFooter `json:"footer,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type EmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// CreateThreadResponse Discord API 的回應
type CreateThreadResponse struct {
	ID   string `json:"id"`   // Thread ID
	Name string `json:"name"` // Thread 名稱
}

// CreateThread 在 forum channel 建立新的 thread
func (c *Client) CreateThread(title string, message ThreadMessage) (string, error) {
	url := fmt.Sprintf("%s/channels/%s/threads", DiscordAPIBase, c.forumChannelID)

	reqBody := CreateThreadRequest{
		Name:    title,
		Message: message,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("discord API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result CreateThreadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ID, nil
}

// PostMessage 在已存在的 thread 中發送訊息
func (c *Client) PostMessage(threadID string, message ThreadMessage) error {
	url := fmt.Sprintf("%s/channels/%s/messages", DiscordAPIBase, threadID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ArchiveThreadRequest archive thread 的請求
type ArchiveThreadRequest struct {
	Archived bool `json:"archived"`
}

// ArchiveThread 關閉並 archive 一個 thread
func (c *Client) ArchiveThread(threadID string) error {
	url := fmt.Sprintf("%s/channels/%s", DiscordAPIBase, threadID)

	reqBody := ArchiveThreadRequest{
		Archived: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
