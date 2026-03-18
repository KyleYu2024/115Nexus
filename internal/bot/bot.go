package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"115nexus/internal/models"
	"115nexus/internal/services"
	"115nexus/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(token string) {
	if token == "" { return }
	// 提升超时时间至 60s 以应对代理延迟
	client := utils.GetProxyClient(models.GlobalConfig)
	client.Timeout = 60 * time.Second
	
	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil { slog.Error("❌ TG Bot 启动失败", "err", err); return }
	
	models.CurrentBot = bot
	slog.Info("🤖 Bot Online", "user", bot.Self.UserName)
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60 // 长轮询超时设为 60s
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { handleMsg(bot, update.Message) } else if update.CallbackQuery != nil { handleCB(bot, update.CallbackQuery) }
	}
}

func handleMsg(bot *tgbotapi.BotAPI, m *tgbotapi.Message) {
	txt := strings.TrimSpace(m.Text); cfg := models.GlobalConfig
	if strings.HasPrefix(txt, "/start") || strings.HasPrefix(txt, "/help") {
		msg := tgbotapi.NewMessage(m.Chat.ID, "👋 *欢迎使用 115Media\\-Bot v0\\.2\\.3*\n\n• `/s 关键词` \\- 搜索影视\n• `/ps 关键词` \\- Pansou 搜索\n\n💡 直接发送 115 链接或磁力链可一键推送。")
		msg.ParseMode = "MarkdownV2"; bot.Send(msg); return
	}
	if strings.HasPrefix(txt, "/ps") {
		kw := strings.TrimSpace(strings.TrimPrefix(txt, "/ps"))
		if kw != "" { doPansouSearch(bot, m.Chat.ID, kw) }
	} else if strings.HasPrefix(txt, "/s") {
		kw := strings.TrimSpace(strings.TrimPrefix(txt, "/s"))
		if kw != "" { doSearch(bot, m.Chat.ID, kw) }
	} else if l := utils.Extract115Link(txt); l != "" {
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, services.PushToMedia302(l, cfg)))
	}
}

func doSearch(bot *tgbotapi.BotAPI, cid int64, kw string) {
	cfg := models.GlobalConfig
	var res []models.MovieMetadata
	tUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=%s&language=zh-CN&query=%s", cfg.TmdbApiKey, url.QueryEscape(kw))
	slog.Info("🔍 TMDB Request URL", "url", tUrl)
	resp, err := utils.GetProxyClient(cfg).Get(tUrl)
	if err == nil {
		defer resp.Body.Close()
		var w models.TmdbSearchResponse; json.NewDecoder(resp.Body).Decode(&w); res = w.Results
	} else {
		slog.Error("❌ TMDB Search Error", "err", err)
		bot.Send(tgbotapi.NewMessage(cid, "❌ 搜索请求失败，请检查网络或代理"))
		return
	}
	if len(res) == 0 { bot.Send(tgbotapi.NewMessage(cid, "❌ No results")); return }
	sid := fmt.Sprintf("s_%d", time.Now().UnixNano())
	models.SearchCache.Store(sid, models.SearchSession{Keyword: kw, Items: res, Time: time.Now()})
	sendSearchPage(bot, cid, sid, 1, false, 0)
}

func doPansouSearch(bot *tgbotapi.BotAPI, cid int64, kw string) {
	cfg := models.GlobalConfig
	items, err := services.DoPansouSearch(kw, cfg)
	if err != nil {
		slog.Error("❌ Pansou Search Error", "err", err)
		bot.Send(tgbotapi.NewMessage(cid, "❌ Pansou 搜索失败，请检查配置"))
		return
	}
	if len(items) == 0 {
		bot.Send(tgbotapi.NewMessage(cid, "📭 Pansou 搜索无结果"))
		return
	}

	sid := fmt.Sprintf("ps_%d", time.Now().UnixNano())
	models.PansouCache.Store(sid, models.PansouSession{Keyword: kw, Items: items, Time: time.Now()})
	sendPansouPage(bot, cid, sid, 1, false, 0)
}

func sendPansouPage(bot *tgbotapi.BotAPI, cid int64, sid string, page int, edit bool, mid int) {
	v, ok := models.PansouCache.Load(sid); if !ok { return }
	sess := v.(models.PansouSession); ps := 5; start := (page-1)*ps; end := start+ps
	if end > len(sess.Items) { end = len(sess.Items) }
	
	var txtBuilder strings.Builder
	txtBuilder.WriteString(fmt.Sprintf("🔍 *Pansou: %s* \\(P%d\\)\n\n", utils.TgEscape(sess.Keyword), page))

	var kb [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i := start; i < end; i++ {
		item := sess.Items[i]
		idx := i - start + 1
		txtBuilder.WriteString(fmt.Sprintf("*%d\\.* %s\n\n", idx, utils.TgEscape(item.Note)))
		
		btnLabel := fmt.Sprintf("📥 存 %d", idx)
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(btnLabel, fmt.Sprintf("p|%s", item.Url)))
		if len(currentRow) == 2 {
			kb = append(kb, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 { kb = append(kb, currentRow) }

	var nav []tgbotapi.InlineKeyboardButton
	if page > 1 { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("psp|%s|%d", sid, page-1))) }
	if end < len(sess.Items) { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("psp|%s|%d", sid, page+1))) }
	if len(nav) > 0 { kb = append(kb, nav) }

	if edit {
		m := tgbotapi.NewEditMessageText(cid, mid, txtBuilder.String()); m.ParseMode="MarkdownV2"; m.ReplyMarkup=&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}; bot.Send(m)
	} else {
		m := tgbotapi.NewMessage(cid, txtBuilder.String()); m.ParseMode="MarkdownV2"; m.ReplyMarkup=tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}; bot.Send(m)
	}
}

func sendSearchPage(bot *tgbotapi.BotAPI, cid int64, sid string, page int, edit bool, mid int) {
	v, ok := models.SearchCache.Load(sid); if !ok { return }
	sess := v.(models.SearchSession); ps := 5; start := (page-1)*ps; end := start+ps
	if end > len(sess.Items) { end = len(sess.Items) }
	txt := fmt.Sprintf("🎬 *Search: %s* \\(P%d\\)\n", utils.TgEscape(sess.Keyword), page)
	var kb [][]tgbotapi.InlineKeyboardButton
	for i := start; i < end; i++ {
		item := sess.Items[i]; t := item.Title; if t == "" { t = item.Name }
		yr := ""; if len(item.ReleaseDate) >= 4 { yr = item.ReleaseDate[:4] } else if len(item.FirstAirDate) >= 4 { yr = item.FirstAirDate[:4] }
		label := fmt.Sprintf("📂 %s (%s)", t, yr)
		kb = append(kb, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("f|%v|%s|%s", item.Id, item.MediaType, sid))})
	}
	var nav []tgbotapi.InlineKeyboardButton
	if page > 1 { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("sp|%s|%d", sid, page-1))) }
	if end < len(sess.Items) { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("sp|%s|%d", sid, page+1))) }
	if len(nav) > 0 { kb = append(kb, nav) }
	if edit {
		m := tgbotapi.NewEditMessageText(cid, mid, txt); m.ParseMode="MarkdownV2"; m.ReplyMarkup=&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}; bot.Send(m)
	} else {
		m := tgbotapi.NewMessage(cid, txt); m.ParseMode="MarkdownV2"; m.ReplyMarkup=tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}; bot.Send(m)
	}
}

func handleCB(bot *tgbotapi.BotAPI, q *tgbotapi.CallbackQuery) {
	cid := q.Message.Chat.ID; data := q.Data; cfg := models.GlobalConfig
	if strings.HasPrefix(data, "f|") {
		p := strings.Split(data, "|"); bot.Request(tgbotapi.NewCallback(q.ID, "Fetching..."))
		res, _ := services.FetchHdhiveResources(p[1], p[2], cfg)
		if len(res) == 0 { bot.Send(tgbotapi.NewMessage(cid, "❌ No resources")); return }
		rsid := fmt.Sprintf("r_%d", time.Now().UnixNano())
		models.ResourceCache.Store(rsid, models.ResourceSession{Title: res[0].Title, Items: res, Time: time.Now()})
		sendResPage(bot, cid, rsid, 1, true, q.Message.MessageID, p[3])
	} else if strings.HasPrefix(data, "p|") {
		p := strings.Split(data, "|"); bot.Send(tgbotapi.NewMessage(cid, services.PushToMedia302(p[1], cfg)))
	} else if strings.HasPrefix(data, "sp|") {
		p := strings.Split(data, "|"); pg, _ := strconv.Atoi(p[2]); sendSearchPage(bot, cid, p[1], pg, true, q.Message.MessageID)
	} else if strings.HasPrefix(data, "rp|") {
		p := strings.Split(data, "|"); pg, _ := strconv.Atoi(p[2]); sendResPage(bot, cid, p[1], pg, true, q.Message.MessageID, p[3])
	} else if strings.HasPrefix(data, "psp|") {
		p := strings.Split(data, "|"); pg, _ := strconv.Atoi(p[2]); sendPansouPage(bot, cid, p[1], pg, true, q.Message.MessageID)
	} else if strings.HasPrefix(data, "back|") {
		p := strings.Split(data, "|"); sendSearchPage(bot, cid, p[1], 1, true, q.Message.MessageID)
	}
}

func sendResPage(bot *tgbotapi.BotAPI, cid int64, rsid string, page int, edit bool, mid int, ssid string) {
	v, ok := models.ResourceCache.Load(rsid); if !ok { return }
	sess := v.(models.ResourceSession)
	ps := 5 // 每页显示5条，因为文本变长了
	start := (page-1)*ps
	end := start+ps
	if end > len(sess.Items) { end = len(sess.Items) }
	
	var txtBuilder strings.Builder
	txtBuilder.WriteString(fmt.Sprintf("📦 *资源列表* \\(第 %d 页\\)\n\n", page))

	var kb [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i := start; i < end; i++ {
		item := sess.Items[i]
		idx := i - start + 1
		
		// 拼接正文
		txtBuilder.WriteString(fmt.Sprintf("*%d\\.* %s\n\n", idx, utils.TgEscape(item.Display)))
		
		// 拼接按钮
		btnLabel := fmt.Sprintf("📥 存 %d", idx)
		if item.HdhivePoints > 0 { 
			btnLabel = fmt.Sprintf("💎 %dpt | 存 %d", item.HdhivePoints, idx) 
		}
		
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(btnLabel, fmt.Sprintf("p|%s", item.Link)))
		if len(currentRow) == 2 {
			kb = append(kb, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 { kb = append(kb, currentRow) }

	var nav []tgbotapi.InlineKeyboardButton
	if page > 1 { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("rp|%s|%d|%s", rsid, page-1, ssid))) }
	if ssid != "" { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("back|%s", ssid))) }
	if end < len(sess.Items) { nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("rp|%s|%d|%s", rsid, page+1, ssid))) }
	if len(nav) > 0 { kb = append(kb, nav) }
	
	m := tgbotapi.NewEditMessageText(cid, mid, txtBuilder.String())
	if !edit {
		mNew := tgbotapi.NewMessage(cid, txtBuilder.String())
		mNew.ParseMode = "MarkdownV2"
		mNew.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}
		bot.Send(mNew)
		return
	}
	
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kb}
	bot.Send(m)
}
