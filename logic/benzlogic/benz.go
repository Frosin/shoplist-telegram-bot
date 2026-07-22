package benzlogic

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/benzstorage"
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	backText = "⬅ Назад"
	timeout  = 5 * time.Second

	// minSnapshots is the threshold below which analytics are not shown.
	minSnapshots = 5

	maxNameLen = 45
)

// dayNames maps Go's time.Weekday (0=Sun … 6=Sat) to Russian abbreviations.
var dayNames = [7]string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

type benzLogic struct {
	sessionItem *session.SessionItem
	storage     *benzstorage.Storage
}

func New(storage *benzstorage.Storage) *benzLogic {
	return &benzLogic{storage: storage}
}

func (b *benzLogic) SetSession(s *session.SessionItem) {
	b.sessionItem = s
}

func (b *benzLogic) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("benz callback:", command)
	return b.getOutput()
}

func (b *benzLogic) GetMessageOutput(curData, msg string) (logic.Output, error) {
	return b.getOutput()
}

func (b *benzLogic) getOutput() (logic.Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData(backText, consts.FirstPageStart)},
		},
	}

	total, err := b.storage.GetTotalSnapshots(ctx)
	if err != nil {
		return logic.Output{Message: "⛽ Ошибка чтения данных: " + err.Error(), Keyboard: keyboard}, nil
	}

	if total < minSnapshots {
		return logic.Output{
			Message:  "⛽ АЗС аналитика\n\nДанных пока недостаточно.\nСбор начат — попробуйте через 30 минут.",
			Keyboard: keyboard,
		}, nil
	}

	analytics, err := b.storage.GetTop10(ctx)
	if err != nil {
		return logic.Output{Message: "⛽ Ошибка аналитики: " + err.Error(), Keyboard: keyboard}, nil
	}

	return logic.Output{
		Message:  formatMessage(analytics, total),
		Keyboard: keyboard,
	}, nil
}

// formatMessage builds the Telegram message text for the top-10 list.
func formatMessage(stations []benzstorage.GasAnalytics, totalSnapshots int) string {
	if len(stations) == 0 {
		return "⛽ Нет данных по заправкам."
	}

	sb := strings.Builder{}
	sb.WriteString("⛽ Топ-10 АЗС (95-й бензин)\n\n")

	for i, st := range stations {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, st.Address)
		fmt.Fprintf(&sb, "   %s\n", truncate(st.AzsName, maxNameLen))

		// Overall 95 availability
		has95Icon := "❌"
		if st.OverallHas95Pct >= 50 {
			has95Icon = "✅"
		}
		fmt.Fprintf(&sb, "   95-й: %s %.0f%% | очередь: %s\n",
			has95Icon, st.OverallHas95Pct, queueLabel(st.OverallAvgQueue))

		// Per-day optimal times with separate has_95 indicator per period
		if len(st.DayStats) > 0 {
			sb.WriteString("   Лучшее время:\n")
			for _, ds := range st.DayStats {
				parts := make([]string, 0, 2)

				if ds.Day.HasData {
					parts = append(parts, fmt.Sprintf("🌞%02d:00%s", ds.Day.Hour, has95Icon95(ds.DayHas95Pct)))
				}
				if ds.Night.HasData {
					parts = append(parts, fmt.Sprintf("🌙%02d:00%s", ds.Night.Hour, has95Icon95(ds.NightHas95Pct)))
				}

				if len(parts) > 0 {
					fmt.Fprintf(&sb, "   %s: %s\n", dayNames[ds.DayOfWeek], strings.Join(parts, "  "))
				}
			}
		}

		sb.WriteString("\n")
	}

	fmt.Fprintf(&sb, "📊 Снимков: %d", totalSnapshots)
	return sb.String()
}

// truncate cuts the string to maxLen runes (adding "…" if needed).
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// has95Icon95 returns a compact emoji indicator for has_95 availability percentage.
// It is appended directly after the time (e.g. "10:00✅").
func has95Icon95(pct float64) string {
	switch {
	case pct >= 50:
		return "✅"
	case pct > 0:
		return "⚠️"
	default:
		return "❌"
	}
}

// queueLabel converts an average queue score (1–4) to a human-readable string.
func queueLabel(score float64) string {
	switch {
	case score <= 1.5:
		return "до 30 мин"
	case score <= 2.5:
		return "30–60 мин"
	case score <= 3.5:
		return "60+ мин"
	default:
		return "нет данных"
	}
}
