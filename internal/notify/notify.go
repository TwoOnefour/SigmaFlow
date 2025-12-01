package notify

import (
	"context"
	"fmt"
	"okx/internal/model"
	"time"
)

// Notifier 通知接口
type Notifier interface {
	Send(ctx context.Context, message string) error
	SendTradeAlert(ctx context.Context, decision *model.Decision) error
	SendErrorAlert(ctx context.Context, err error) error
}

// Manager 通知管理器
type Manager struct {
	notifiers []Notifier
	enabled   bool
}

// NewManager 创建通知管理器
func NewManager(enabled bool) *Manager {
	return &Manager{
		notifiers: make([]Notifier, 0),
		enabled:   enabled,
	}
}

// AddNotifier 添加通知器
func (m *Manager) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

// Send 发送消息
func (m *Manager) Send(ctx context.Context, message string) error {
	if !m.enabled {
		return nil
	}

	var lastErr error
	for _, n := range m.notifiers {
		if err := n.Send(ctx, message); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// SendTradeAlert 发送交易提醒
func (m *Manager) SendTradeAlert(ctx context.Context, decision *model.Decision) error {
	if !m.enabled || decision == nil {
		return nil
	}

	var lastErr error
	for _, n := range m.notifiers {
		if err := n.SendTradeAlert(ctx, decision); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// SendErrorAlert 发送错误提醒
func (m *Manager) SendErrorAlert(ctx context.Context, err error) error {
	if !m.enabled || err == nil {
		return nil
	}

	var lastErr error
	for _, n := range m.notifiers {
		if alertErr := n.SendErrorAlert(ctx, err); alertErr != nil {
			lastErr = alertErr
		}
	}
	return lastErr
}

// ConsoleNotifier 控制台通知器（用于开发/测试）
type ConsoleNotifier struct{}

// NewConsoleNotifier 创建控制台通知器
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (c *ConsoleNotifier) Send(ctx context.Context, message string) error {
	fmt.Printf("[NOTIFY %s] %s\n", time.Now().UTC().Format(time.RFC3339), message)
	return nil
}

func (c *ConsoleNotifier) SendTradeAlert(ctx context.Context, decision *model.Decision) error {
	emoji := "📊"
	switch decision.Action {
	case "BUY":
		emoji = "🟢"
	case "SELL":
		emoji = "🔴"
	case "HOLD":
		emoji = "🟡"
	}

	message := fmt.Sprintf(`%s Trade Alert
Action: %s
Position: %.1f%%
Stop Loss: $%.2f
Take Profit: $%.2f
Reason: %s`,
		emoji,
		decision.Action,
		decision.PositionPct*100,
		decision.StopLossPrice,
		decision.TakeProfitPrice,
		decision.Reason,
	)
	return c.Send(ctx, message)
}

func (c *ConsoleNotifier) SendErrorAlert(ctx context.Context, err error) error {
	message := fmt.Sprintf("🚨 Error Alert: %v", err)
	return c.Send(ctx, message)
}

// TelegramNotifier Telegram通知器（接口定义，实际实现需要HTTP客户端）
type TelegramNotifier struct {
	botToken string
	chatID   string
}

// NewTelegramNotifier 创建Telegram通知器
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
	}
}

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	// TODO: 实现Telegram API调用
	// 使用 https://api.telegram.org/bot<token>/sendMessage
	if t.botToken == "" || t.chatID == "" {
		return nil
	}
	// 此处添加HTTP请求发送消息
	return nil
}

func (t *TelegramNotifier) SendTradeAlert(ctx context.Context, decision *model.Decision) error {
	emoji := "📊"
	switch decision.Action {
	case "BUY":
		emoji = "🟢"
	case "SELL":
		emoji = "🔴"
	case "HOLD":
		emoji = "🟡"
	}

	message := fmt.Sprintf(`%s *Trade Alert*
*Action:* %s
*Position:* %.1f%%
*Stop Loss:* $%.2f
*Take Profit:* $%.2f
*Reason:* %s`,
		emoji,
		decision.Action,
		decision.PositionPct*100,
		decision.StopLossPrice,
		decision.TakeProfitPrice,
		decision.Reason,
	)
	return t.Send(ctx, message)
}

func (t *TelegramNotifier) SendErrorAlert(ctx context.Context, err error) error {
	message := fmt.Sprintf("🚨 *Error Alert*\n```\n%v\n```", err)
	return t.Send(ctx, message)
}
