package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegoutil"
)

func main() {
	// --- Load .env ---
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	// --- Slog setup ---
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// --- Bot setup ---
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	voiceBaseURL := os.Getenv("VOICE_BASE_URL")

	if token == "" || voiceBaseURL == "" {
		slog.Error("Required environment variables not set", "TELEGRAM_BOT_TOKEN", token, "VOICE_BASE_URL", voiceBaseURL)
		os.Exit(1)
	}

	// Optional: Use your own HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	bot, err := telego.NewBot(token,
		telego.WithHTTPClient(client),
		telego.WithDefaultDebugLogger(),
	)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	// Start polling updates
	err = bot.StartPolling(nil)
	if err != nil {
		slog.Error("Failed to start polling", "error", err)
		os.Exit(1)
	}
	defer bot.StopPolling()

	slog.Info("Bot started!")

	// --- Inline query handler ---
	go func() {
		for update := range bot.Updates {
			if update.InlineQuery != nil {
				go handleInlineQuery(bot, update.InlineQuery, voiceBaseURL)
			}
		}
	}()

	// Keep main running
	select {}
}

func handleInlineQuery(bot *telego.Bot, q *telego.InlineQuery, baseURL string) {
	slog.Info("Inline query received", "query", q.Query, "user", q.From.Username)

	// Example voice message (must be a public URL to .ogg Opus file)
	voice := telegoutil.InlineQueryResultVoice("voice-1", baseURL+"/greeting.ogg").
		WithTitle("Example Voice")

	err := bot.AnswerInlineQuery(&telego.AnswerInlineQueryParams{
		InlineQueryID: q.ID,
		Results:       []telego.InlineQueryResult{voice},
		CacheTime:     0,
	})
	if err != nil {
		slog.Error("Failed to answer inline query", "error", err)
	}
}
