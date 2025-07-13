package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"

	"github.com/aberyotaro/drink-tracker/internal/services"
)

type SlackHandler struct {
	client        *slack.Client
	signingSecret string
	userService   *services.UserService
	drinkService  *services.DrinkService
}

type DrinkConfig struct {
	Name              string
	DefaultAmount     int
	AlcoholPercentage float64
	DisplayName       string
}

func NewSlackHandler(client *slack.Client, signingSecret string, userService *services.UserService, drinkService *services.DrinkService) *SlackHandler {
	return &SlackHandler{
		client:        client,
		signingSecret: signingSecret,
		userService:   userService,
		drinkService:  drinkService,
	}
}

func (h *SlackHandler) getDrinkConfigs() map[string]DrinkConfig {
	return map[string]DrinkConfig{
		"beer":     {"beer", 350, 0.05, "ビール"},
		"b":        {"beer", 350, 0.05, "ビール"},
		"wine":     {"wine", 150, 0.12, "ワイン"},
		"w":        {"wine", 150, 0.12, "ワイン"},
		"sake":     {"sake", 180, 0.15, "日本酒"},
		"sk":       {"sake", 180, 0.15, "日本酒"},
		"whiskey":  {"whiskey", 30, 0.40, "ウイスキー"},
		"wh":       {"whiskey", 30, 0.40, "ウイスキー"},
		"shochu":   {"shochu", 60, 0.25, "焼酎"},
		"sh":       {"shochu", 60, 0.25, "焼酎"},
		"highball": {"highball", 350, 0.09, "ハイボール"},
		"hi":       {"highball", 350, 0.09, "ハイボール"},
	}
}

func (h *SlackHandler) HandleSlashCommand(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot read request body"})
	}

	// デバッグ用ログ
	timestamp := c.Request().Header.Get("X-Slack-Request-Timestamp")
	signature := c.Request().Header.Get("X-Slack-Signature")
	fmt.Printf("DEBUG: timestamp=%s, signature=%s, signingSecret_len=%d\n", timestamp, signature, len(h.signingSecret))

	if !h.verifySlackRequest(c.Request(), body) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid signature"})
	}

	// Body を再設定（SlashCommandParse 用）
	c.Request().Body = io.NopCloser(strings.NewReader(string(body)))

	command, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		fmt.Printf("DEBUG: SlashCommandParse error: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot parse slash command"})
	}

	if command.Command != "/drink" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown command"})
	}

	response := h.processDrinkCommand(command)
	return c.JSON(http.StatusOK, response)
}

func (h *SlackHandler) processDrinkCommand(command slack.SlashCommand) *slack.Msg {
	args := strings.Fields(command.Text)

	if len(args) == 0 {
		return &slack.Msg{
			Text: "使用方法: `/drink [種類] [量]`\n例: `/drink beer`, `/drink wine 150ml`",
		}
	}

	drinkType := args[0]
	amount := ""

	if len(args) > 1 {
		amount = args[1]
	}

	switch drinkType {
	case "stats":
		return h.handleStatsRequest(amount)
	case "help":
		return h.handleHelpRequest()
	default:
		drinkConfigs := h.getDrinkConfigs()
		if config, exists := drinkConfigs[drinkType]; exists {
			return h.handleDrinkRecord(command, config, amount)
		}
		return &slack.Msg{
			Text: fmt.Sprintf("未対応の飲み物です: %s\n対応している種類: b/beer, w/wine, sk/sake, wh/whiskey, sh/shochu, hi/highball", drinkType),
		}
	}
}

func (h *SlackHandler) handleDrinkRecord(command slack.SlashCommand, config DrinkConfig, amount string) *slack.Msg {
	ctx := context.Background()

	// ユーザーを取得または作成
	user, err := h.userService.GetOrCreateUser(ctx, command.UserID, command.TeamID)
	if err != nil {
		return &slack.Msg{
			Text: "ユーザー情報の取得に失敗しました。",
		}
	}

	// 量をパース
	amountMl := config.DefaultAmount
	if amount != "" {
		// 日本酒の特別処理
		if config.Name == "sake" && (amount == "1go" || amount == "一合") {
			amountMl = 180
		} else if parsed, err := parseAmount(amount); err == nil {
			amountMl = parsed
		}
	}

	// データベースに記録
	record, err := h.drinkService.RecordDrink(ctx, int64(user.ID), config.Name, int64(amountMl), config.AlcoholPercentage)
	if err != nil {
		fmt.Printf("DEBUG: RecordDrink error: %v\n", err)
		return &slack.Msg{
			Text: "記録の保存に失敗しました。",
		}
	}
	fmt.Printf("DEBUG: RecordDrink success: ID=%d, UserID=%d, DrinkType=%s, AmountML=%d\n", 
		record.ID, record.UserID, record.DrinkType, record.AmountML)

	// 今日の合計を取得
	totalAlcohol, totalMl, err := h.drinkService.GetTodayTotalAlcohol(ctx, int64(user.ID))
	if err != nil {
		fmt.Printf("DEBUG: GetTodayTotalAlcohol error: %v\n", err)
		return &slack.Msg{
			Text: "統計情報の取得に失敗しました。",
		}
	}
	
	fmt.Printf("DEBUG: totalAlcohol=%.1f, totalMl=%d\n", totalAlcohol, totalMl)

	// 適量チェック
	var emoji string
	var warning string
	if totalAlcohol <= 20 {
		emoji = "😊"
		warning = "適量内です"
	} else {
		emoji = "🚨"
		warning = "適量を超えています。お気をつけください。"
	}

	return &slack.Msg{
		Text: fmt.Sprintf("✅ %s%dmlを記録しました\n📊 今日の飲酒量: %dml (純アルコール%.1fg)\n%s %s",
			config.DisplayName, amountMl, totalMl, totalAlcohol, emoji, warning),
	}
}

func (h *SlackHandler) handleStatsRequest(_ string) *slack.Msg {
	return &slack.Msg{
		Text: "📊 統計機能は現在開発中です",
	}
}

func (h *SlackHandler) handleHelpRequest() *slack.Msg {
	return &slack.Msg{
		Text: "🍺 飲酒管理アプリ\n\n" +
			"**使用方法:**\n" +
			"`/drink b` または `/drink beer` - ビール350mlを記録\n" +
			"`/drink b 350` - ビール350mlを記録\n" +
			"`/drink b 500` - ビール500mlを記録\n" +
			"`/drink w` または `/drink wine` - ワイン150mlを記録\n" +
			"`/drink sk` または `/drink sake` - 日本酒1合を記録\n" +
			"`/drink wh` または `/drink whiskey` - ウイスキー30mlを記録\n" +
			"`/drink sh` または `/drink shochu` - 焼酎60mlを記録\n" +
			"`/drink hi` または `/drink highball` - ハイボール350mlを記録\n" +
			"`/drink hi 500` - ハイボール500mlを記録\n" +
			"`/drink stats` - 統計表示（開発中）\n" +
			"`/drink help` - このヘルプを表示",
	}
}

func parseAmount(amount string) (int, error) {
	amount = strings.ToLower(amount)
	amount = strings.ReplaceAll(amount, "ml", "")
	amount = strings.TrimSpace(amount)

	return strconv.Atoi(amount)
}

func (h *SlackHandler) verifySlackRequest(r *http.Request, body []byte) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-t > 300 {
		return false
	}

	baseString := "v0:" + timestamp + ":" + string(body)

	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write([]byte(baseString))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
