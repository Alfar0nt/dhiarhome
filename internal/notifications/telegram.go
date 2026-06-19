package notifications

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Notifier struct {
	BotToken        string
	ChatID          string
	MessageThreadID int
	NotifyUp        bool
	NotifyDown      bool
	Cooldown        time.Duration
	SilentHours     []int
	Mock            bool

	mu           sync.Mutex
	lastNotified map[string]time.Time
}

func NewNotifier(botToken, chatID string, messageThreadID int, notifyUp, notifyDown bool, cooldownMinutes int, silentHours []int, mock bool) *Notifier {
	cd := time.Duration(cooldownMinutes) * time.Minute
	if cd <= 0 {
		cd = 5 * time.Minute
	}
	if silentHours == nil {
		silentHours = []int{}
	}
	return &Notifier{
		BotToken:        botToken,
		ChatID:          chatID,
		MessageThreadID: messageThreadID,
		NotifyUp:        notifyUp,
		NotifyDown:      notifyDown,
		Cooldown:        cd,
		SilentHours:     silentHours,
		Mock:            mock,
		lastNotified:    make(map[string]time.Time),
	}
}

func (n *Notifier) SendMessage(message string) error {
	if n.Mock {
		log.Printf("[TELEGRAM MOCK] Would send: %s", message)
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.BotToken)
	body := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"HTML"`, n.ChatID, message)
	if n.MessageThreadID > 0 {
		body += fmt.Sprintf(`,"message_thread_id":%d`, n.MessageThreadID)
	}
	body += `}`

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("telegram send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) NotifyServiceChange(name, url, oldStatus, newStatus string) {
	if !n.shouldNotify("service:"+name, newStatus == "Online") {
		return
	}

	var emoji, direction string
	if newStatus == "Online" {
		emoji = "🟢"
		direction = "Recovered"
	} else {
		emoji = "🔴"
		direction = "Down"
	}

	msg := fmt.Sprintf("%s <b>%s</b>\nService: %s\nURL: %s\nStatus: %s → %s\nTime: %s",
		emoji, direction, name, url, oldStatus, newStatus, time.Now().Format("2006-01-02 15:04:05"))

	if err := n.SendMessage(msg); err != nil {
		log.Printf("[TELEGRAM] Failed to send service notification: %v", err)
	}
}

func (n *Notifier) NotifyContainerChange(name, oldState, newState string) {
	if !n.shouldNotify("container:"+name, newState == "running") {
		return
	}

	var emoji, direction string
	if newState == "running" {
		emoji = "🟢"
		direction = "Started"
	} else {
		emoji = "🔴"
		direction = "Stopped"
	}

	msg := fmt.Sprintf("%s <b>%s</b>\nContainer: %s\nState: %s → %s\nTime: %s",
		emoji, direction, name, oldState, newState, time.Now().Format("2006-01-02 15:04:05"))

	if err := n.SendMessage(msg); err != nil {
		log.Printf("[TELEGRAM] Failed to send container notification: %v", err)
	}
}

func (n *Notifier) NotifyTest() error {
	msg := fmt.Sprintf("🔄 <b>Test Notification</b>\n\nYour dhiarhome Telegram notifications are configured correctly!\nTime: %s", time.Now().Format("2006-01-02 15:04:05"))
	return n.SendMessage(msg)
}

func (n *Notifier) shouldNotify(key string, isUp bool) bool {
	if isUp && !n.NotifyUp {
		return false
	}
	if !isUp && !n.NotifyDown {
		return false
	}

	if n.isSilentHour() {
		log.Printf("[TELEGRAM] Suppressed notification for %s (silent hour)", key)
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	last, ok := n.lastNotified[key]
	if ok && time.Since(last) < n.Cooldown {
		return false
	}
	n.lastNotified[key] = time.Now()
	return true
}

func (n *Notifier) isSilentHour() bool {
	if len(n.SilentHours) == 0 {
		return false
	}
	currentHour := time.Now().Hour()
	for _, h := range n.SilentHours {
		if currentHour == h {
			return true
		}
	}
	return false
}
