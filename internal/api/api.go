package api

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "time"
)

const (
  BaseURL = "https://cursor.com"
)

type Client struct {
  token      string
  userID     string
  httpClient *http.Client
}

func NewClient(token, userID string) *Client {
  return &Client{
    token:  token,
    userID: userID,
    httpClient: &http.Client{
      Timeout: 30 * time.Second,
    },
  }
}

type ModelUsage struct {
  NumRequests      int  `json:"numRequests"`
  NumRequestsTotal int  `json:"numRequestsTotal"`
  NumTokens        int  `json:"numTokens"`
  MaxRequestUsage  *int `json:"maxRequestUsage"`
  MaxTokenUsage    *int `json:"maxTokenUsage"`
}

type UsageResponse struct {
  GPT4         ModelUsage `json:"gpt-4"`
  StartOfMonth string     `json:"startOfMonth"`
}

type TokenUsage struct {
  InputTokens      int     `json:"inputTokens"`
  OutputTokens     int     `json:"outputTokens"`
  CacheWriteTokens int     `json:"cacheWriteTokens"`
  CacheReadTokens  int     `json:"cacheReadTokens"`
  TotalCents       float64 `json:"totalCents"`
}

type UsageEvent struct {
  Timestamp        string      `json:"timestamp"`
  Model            string      `json:"model"`
  Kind             string      `json:"kind"`
  RequestsCosts    float64     `json:"requestsCosts"`
  UsageBasedCosts  string      `json:"usageBasedCosts"`
  IsTokenBasedCall bool        `json:"isTokenBasedCall"`
  TokenUsage       *TokenUsage `json:"tokenUsage,omitempty"`
  OwningUser       string      `json:"owningUser"`
  CursorTokenFee   float64     `json:"cursorTokenFee"`
  IsChargeable     bool        `json:"isChargeable"`
}

type FilteredUsageResponse struct {
  TotalUsageEventsCount int          `json:"totalUsageEventsCount"`
  UsageEventsDisplay    []UsageEvent `json:"usageEventsDisplay"`
}

type OnDemandUsage struct {
  Enabled   bool `json:"enabled"`
  Used      int  `json:"used"`
  Limit     int  `json:"limit"`
  Remaining int  `json:"remaining"`
}

type PlanUsage struct {
  Enabled          bool    `json:"enabled"`
  Used             int     `json:"used"`
  Limit            int     `json:"limit"`
  Remaining        int     `json:"remaining"`
  AutoPercentUsed  float64 `json:"autoPercentUsed"`
  APIPercentUsed   float64 `json:"apiPercentUsed"`
  TotalPercentUsed float64 `json:"totalPercentUsed"`
}

type IndividualUsage struct {
  Plan     PlanUsage     `json:"plan"`
  OnDemand OnDemandUsage `json:"onDemand"`
}

type UsageSummaryResponse struct {
  BillingCycleStart string          `json:"billingCycleStart"`
  BillingCycleEnd   string          `json:"billingCycleEnd"`
  MembershipType    string          `json:"membershipType"`
  IndividualUsage   IndividualUsage `json:"individualUsage"`
}

func (c *Client) getBrowserHeaders() map[string]string {
  return map[string]string{
    "Content-Type":    "application/json",
    "Cookie":          fmt.Sprintf("WorkosCursorSessionToken=%s", c.token),
    "Origin":          "https://cursor.com",
    "Referer":         "https://cursor.com/dashboard",
    "Sec-Fetch-Site":  "same-origin",
    "Sec-Fetch-Mode":  "cors",
    "Sec-Fetch-Dest":  "empty",
    "User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Accept":          "*/*",
    "Accept-Language": "en",
  }
}

func (c *Client) GetUsage() (*UsageResponse, error) {
  reqURL := fmt.Sprintf("%s/api/usage?user=%s", BaseURL, c.userID)
  req, err := http.NewRequest("GET", reqURL, nil)
  if err != nil {
    return nil, fmt.Errorf("failed to create request: %w", err)
  }

  for k, v := range c.getBrowserHeaders() {
    req.Header.Set(k, v)
  }

  resp, err := c.httpClient.Do(req)
  if err != nil {
    return nil, fmt.Errorf("request failed: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
  }

  var usage UsageResponse
  if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
    return nil, fmt.Errorf("failed to decode response: %w", err)
  }

  return &usage, nil
}

func (c *Client) GetFilteredUsageEvents(startDate, endDate time.Time, page, pageSize int) (*FilteredUsageResponse, error) {
  payload := map[string]interface{}{
    "startDate": fmt.Sprintf("%d", startDate.UnixMilli()),
    "endDate":   fmt.Sprintf("%d", endDate.UnixMilli()),
    "page":      page,
    "pageSize":  pageSize,
  }

  jsonPayload, err := json.Marshal(payload)
  if err != nil {
    return nil, fmt.Errorf("failed to marshal payload: %w", err)
  }

  req, err := http.NewRequest("POST", BaseURL+"/api/dashboard/get-filtered-usage-events", bytes.NewBuffer(jsonPayload))
  if err != nil {
    return nil, fmt.Errorf("failed to create request: %w", err)
  }

  for k, v := range c.getBrowserHeaders() {
    req.Header.Set(k, v)
  }

  resp, err := c.httpClient.Do(req)
  if err != nil {
    return nil, fmt.Errorf("request failed: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
  }

  var eventsResp FilteredUsageResponse
  if err := json.NewDecoder(resp.Body).Decode(&eventsResp); err != nil {
    return nil, fmt.Errorf("failed to decode response: %w", err)
  }

  return &eventsResp, nil
}

func (c *Client) GetUsageSummary() (*UsageSummaryResponse, error) {
  reqURL := fmt.Sprintf("%s/api/usage-summary?user=%s", BaseURL, c.userID)
  req, err := http.NewRequest("GET", reqURL, nil)
  if err != nil {
    return nil, fmt.Errorf("failed to create request: %w", err)
  }

  for k, v := range c.getBrowserHeaders() {
    req.Header.Set(k, v)
  }

  resp, err := c.httpClient.Do(req)
  if err != nil {
    return nil, fmt.Errorf("request failed: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
  }

  var summary UsageSummaryResponse
  if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
    return nil, fmt.Errorf("failed to decode response: %w", err)
  }

  return &summary, nil
}

func (c *Client) ValidateToken() error {
  _, err := c.GetUsage()
  return err
}
