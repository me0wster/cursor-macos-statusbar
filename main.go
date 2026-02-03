package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"cursor-bar/internal/api"
	"cursor-bar/internal/config"
	"cursor-bar/internal/icon"

	"github.com/getlantern/systray"
)

var (
	client           *api.Client
	usageMutex       sync.RWMutex
	lastError        string
	eventItems       []*systray.MenuItem
	requestUsageItem *systray.MenuItem
	requestBarItem   *systray.MenuItem
	requestResetItem *systray.MenuItem
	moneyUsageItem   *systray.MenuItem
	moneyBarItem     *systray.MenuItem
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("...")
	systray.SetTooltip("Cursor Usage")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
	}

	if !config.IsValid(cfg) {
		token, userID := promptForCredentials()
		if token == "" || userID == "" {
			systray.SetTitle("!")
			setupMenu()
			return
		}

		cfg = &config.Config{Token: token, UserID: userID}
		if err := config.Save(cfg); err != nil {
			log.Printf("Error saving config: %v", err)
		}
	}

	client = api.NewClient(cfg.Token, cfg.UserID)

	if err := client.ValidateToken(); err != nil {
		systray.SetTitle("!")
		lastError = err.Error()
		setupMenu()
		return
	}

	setupMenu()
	refreshUsage()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			refreshUsage()
		}
	}()
}

func setupMenu() {
	// Request Usage Section
	mRequestHeader := systray.AddMenuItem("Request Usage", "")
	mRequestHeader.Disable()

	requestUsageItem = systray.AddMenuItem("...", "")
	requestUsageItem.Disable()

	requestBarItem = systray.AddMenuItem("...", "")
	requestBarItem.Disable()

	requestResetItem = systray.AddMenuItem("...", "")
	requestResetItem.Disable()

	systray.AddSeparator()

	// Money Usage Section
	mMoneyHeader := systray.AddMenuItem("Request Usage (Money)", "")
	mMoneyHeader.Disable()

	moneyUsageItem = systray.AddMenuItem("...", "")
	moneyUsageItem.Disable()

	moneyBarItem = systray.AddMenuItem("...", "")
	moneyBarItem.Disable()

	systray.AddSeparator()

	// Recent Usage Section
	mHeader := systray.AddMenuItem("Recent Usage", "")
	mHeader.Disable()

	systray.AddSeparator()

	eventItems = make([]*systray.MenuItem, 10)
	for i := 0; i < 10; i++ {
		eventItems[i] = systray.AddMenuItem("...", "")
		eventItems[i].Disable()
	}

	systray.AddSeparator()

	mRefresh := systray.AddMenuItem("Refresh", "")
	mDashboard := systray.AddMenuItem("Open Dashboard", "")
	mEditConfig := systray.AddMenuItem("Edit Config", "")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "")

	go func() {
		for {
			select {
			case <-mRefresh.ClickedCh:
				systray.SetTitle("...")
				refreshUsage()
			case <-mDashboard.ClickedCh:
				openURL("https://cursor.com/settings")
			case <-mEditConfig.ClickedCh:
				openConfig()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func refreshUsage() {
	if client == nil {
		return
	}

	usage, err := client.GetUsage()
	if err != nil {
		usageMutex.Lock()
		lastError = err.Error()
		usageMutex.Unlock()
		systray.SetTitle("!")
		return
	}

	maxReq := 500
	if usage.GPT4.MaxRequestUsage != nil {
		maxReq = *usage.GPT4.MaxRequestUsage
	}
	percent := 0
	if maxReq > 0 {
		percent = (usage.GPT4.NumRequests * 100) / maxReq
	}

	if percent >= 90 {
		systray.SetTitle(fmt.Sprintf("⚠️%d%%", percent))
	} else {
		systray.SetTitle(fmt.Sprintf("%d%%", percent))
	}

	now := time.Now()
	startOfMonth, _ := time.Parse(time.RFC3339, usage.StartOfMonth)
	if startOfMonth.IsZero() {
		startOfMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	// Update request usage section
	updateRequestUsage(usage, startOfMonth)

	// Fetch and update money usage section
	summary, err := client.GetUsageSummary()
	if err != nil {
		log.Printf("Error fetching usage summary: %v", err)
	} else {
		updateMoneyUsage(summary)
	}

	events, err := client.GetFilteredUsageEvents(startOfMonth, now, 1, 10)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		return
	}

	updateEventItems(events.UsageEventsDisplay)
}

func updateEventItems(events []api.UsageEvent) {
	for i, item := range eventItems {
		if i < len(events) {
			e := events[i]
			item.SetTitle(formatEvent(e))
		} else {
			item.SetTitle("---")
		}
	}
}

func updateRequestUsage(usage *api.UsageResponse, startOfMonth time.Time) {
	maxReq := 500
	if usage.GPT4.MaxRequestUsage != nil {
		maxReq = *usage.GPT4.MaxRequestUsage
	}

	numReq := usage.GPT4.NumRequestsTotal
	percent := 0
	if maxReq > 0 {
		percent = (numReq * 100) / maxReq
	}

	requestUsageItem.SetTitle(fmt.Sprintf("%d/%d = %d%%", numReq, maxReq, percent))
	requestBarItem.SetTitle(makeProgressBar(numReq, maxReq, 20))
	requestResetItem.SetTitle(fmt.Sprintf("Resets on %s", startOfMonth.Format("Jan 2, 2006")))
}

func updateMoneyUsage(summary *api.UsageSummaryResponse) {
	used := summary.IndividualUsage.OnDemand.Used
	limit := summary.IndividualUsage.OnDemand.Limit

	dollars := float64(used) / 100.0
	maxDollars := float64(limit) / 100.0

	moneyUsageItem.SetTitle(fmt.Sprintf("$%.2f/$%.0f", dollars, maxDollars))
	moneyBarItem.SetTitle(makeProgressBar(used, limit, 20))
}

func makeProgressBar(current, max, width int) string {
	if max == 0 {
		return strings.Repeat("○", width)
	}

	filled := (current * width) / max
	if filled > width {
		filled = width
	}

	return strings.Repeat("●", filled) + strings.Repeat("○", width-filled)
}

func formatEvent(e api.UsageEvent) string {
	ts, _ := strconv.ParseInt(e.Timestamp, 10, 64)
	t := time.UnixMilli(ts)
	dateStr := t.Format("01/02 15:04")

	kindShort := parseKind(e.Kind)
	modelShort := shortenModel(e.Model)

	inputTokens := 0
	outputTokens := 0
	totalCents := 0.0
	if e.TokenUsage != nil {
		inputTokens = e.TokenUsage.InputTokens
		outputTokens = e.TokenUsage.OutputTokens
		totalCents = e.TokenUsage.TotalCents
	}

	// Use generous spacing for proportional fonts in macOS menu bars
	return fmt.Sprintf("%-13s   %-7s   %-18s   %s   %5.0f   $%.2f",
		dateStr,
		kindShort,
		modelShort,
		formatTokens(inputTokens, outputTokens),
		e.RequestsCosts,
		totalCents/100.0,
	)
}

func parseKind(kind string) string {
	switch kind {
	case "USAGE_EVENT_KIND_INCLUDED_IN_PRO":
		return "Pro"
	case "USAGE_EVENT_KIND_USER_API_KEY":
		return "API"
	case "USAGE_EVENT_KIND_ABORTED_NOT_CHARGED":
		return "Abort"
	default:
		if strings.Contains(kind, "PRO") {
			return "Pro"
		}
		if strings.Contains(kind, "API") {
			return "API"
		}
		return "Other"
	}
}

func shortenModel(model string) string {
	replacements := map[string]string{
		"claude-4.5-sonnet-thinking":    "claude-sonnet",
		"claude-4.5-sonnet":             "claude-sonnet",
		"claude-4.5-opus-high-thinking": "claude-opus",
		"claude-4.5-haiku-thinking":     "claude-haiku",
		"gemini-3-pro-preview":          "gemini-pro",
		"gemini-3-flash-preview":        "gemini-flash",
		"gpt-5.2":                       "gpt-5.2",
		"gpt-5.1-codex-max":             "gpt-codex",
		"composer-1":                    "composer",
		"agent_review":                  "agent",
	}

	if short, ok := replacements[model]; ok {
		return short
	}
	if len(model) > 15 {
		return model[:15]
	}
	return model
}

func formatTokens(input, output int) string {
	return fmt.Sprintf("%s (IN) - (OUT) %s", formatTokenCount(input), formatTokenCount(output))
}

func formatTokenCount(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func promptForCredentials() (string, string) {
	if runtime.GOOS != "darwin" {
		log.Println("Prompt only supported on macOS")
		return "", ""
	}

	tokenScript := `
    set tokenValue to text returned of (display dialog "Enter your WorkosCursorSessionToken cookie:" & return & return & "Find it in DevTools > Application > Cookies > cursor.com" default answer "" with title "Cursor Bar Setup" with icon note)
    return tokenValue
  `
	cmd := exec.Command("osascript", "-e", tokenScript)
	tokenOut, err := cmd.Output()
	if err != nil {
		log.Printf("Error prompting for token: %v", err)
		return "", ""
	}
	token := strings.TrimSpace(string(tokenOut))

	userIDScript := `
    set userValue to text returned of (display dialog "Enter your Cursor User ID:" & return & return & "Find it in the cookie value or API responses" default answer "" with title "Cursor Bar Setup" with icon note)
    return userValue
  `
	cmd = exec.Command("osascript", "-e", userIDScript)
	userOut, err := cmd.Output()
	if err != nil {
		log.Printf("Error prompting for user ID: %v", err)
		return "", ""
	}
	userID := strings.TrimSpace(string(userOut))

	return token, userID
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	cmd.Start()
}

func openConfig() {
	path, err := config.GetConfigPath()
	if err != nil {
		log.Printf("Error getting config path: %v", err)
		return
	}
	exec.Command("open", "-t", path).Start()
}

func onExit() {}
