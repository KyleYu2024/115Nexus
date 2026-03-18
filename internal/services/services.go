package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"115nexus/internal/models"
	"115nexus/internal/utils"
)

// === 基础服务 ===

func SendWebhook(cfg models.BotConfig, title, content string) {
	if cfg.WebhookUrl == "" { return }
	p := map[string]string{"title": title, "content": content}
	b, _ := json.Marshal(p)
	slog.Info("📡 发送 Webhook", "title", title, "content", content)
	r, err := utils.GetDirectClient().Post(cfg.WebhookUrl, "application/json", bytes.NewBuffer(b))
	if err == nil { r.Body.Close() }
}

// === HDHive 身份与签到 ===

func DoHdhiveCheckin(cfg models.BotConfig) (string, error) {
	if !cfg.HdhiveCheckinEnabled { return "", nil }
	var finalMsg string
	var finalErr error

	if cfg.HdhiveUser != "" && cfg.HdhivePass != "" {
		slog.Info("📅 HDHive Web 模拟签到模式", "user", cfg.HdhiveUser)
		jar, _ := cookiejar.New(nil)
		client := utils.GetProxyClient(cfg)
		client.Jar = jar
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		
		r1, _ := http.NewRequest("GET", "https://hdhive.com/login", nil)
		r1.Header.Set("User-Agent", ua)
		client.Do(r1)

		lPayload := []interface{}{map[string]string{"username": cfg.HdhiveUser, "password": cfg.HdhivePass}, "/"}
		lb, _ := json.Marshal(lPayload)
		r2, _ := http.NewRequest("POST", "https://hdhive.com/login", bytes.NewBuffer(lb))
		r2.Header.Set("User-Agent", ua)
		r2.Header.Set("Accept", "text/x-component")
		r2.Header.Set("Content-Type", "text/plain;charset=UTF-8")
		r2.Header.Set("Next-Action", "601d84138632d39f16adfce544ceb527a6f6243670")
		
		resp, errL := client.Do(r2)
		if errL == nil && (resp.StatusCode == 200 || resp.StatusCode == 303) {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			cb, _ := json.Marshal([]bool{cfg.HdhiveGamblerMode})
			r3, _ := http.NewRequest("POST", "https://hdhive.com/", bytes.NewBuffer(cb))
			r3.Header.Set("User-Agent", ua)
			r3.Header.Set("Accept", "text/x-component")
			r3.Header.Set("Content-Type", "text/plain;charset=UTF-8")
			r3.Header.Set("Next-Action", "409539c7faa0ad25d3e3e8c21465c10661896ca5a2")
			
			resp2, errC := client.Do(r3)
			if errC == nil {
				defer resp2.Body.Close()
				body, _ := ioutil.ReadAll(resp2.Body)
				txt := string(body)
				msg := "签到请求已执行"
				reDesc := regexp.MustCompile(`"description":"([^"]+)"`)
				reMsg := regexp.MustCompile(`"message":"([^"]+)"`)
				if m := reDesc.FindStringSubmatch(txt); len(m) > 1 { msg = m[1] } else if m := reMsg.FindStringSubmatch(txt); len(m) > 1 { msg = m[1] }
				finalMsg = msg
				slog.Info("✅ Web 签到结果详情", "feedback", msg)
			} else { finalErr = errC }
		} else { finalErr = fmt.Errorf("登录失败: Status %d", resp.StatusCode) }
	} else {
		slog.Info("📅 HDHive API 签到模式")
		req, _ := http.NewRequest("POST", "https://hdhive.com/api/open/checkin", nil)
		req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
		resp, errA := utils.GetProxyClient(cfg).Do(req)
		if errA == nil {
			defer resp.Body.Close()
			var w models.HdhiveApiResponse
			json.NewDecoder(resp.Body).Decode(&w)
			finalMsg = w.Message
			if !w.Success { finalErr = fmt.Errorf(w.Message) }
		} else { finalErr = errA }
	}

	if finalMsg != "" || finalErr != nil {
		content := finalMsg
		if finalErr != nil { content = "失败: " + finalErr.Error() }
		SendWebhook(cfg, "影巢签到", content)
	}
	return finalMsg, finalErr
}

// === HDHive 资源获取 ===

func FetchHdhiveResources(id string, mType string, cfg models.BotConfig) ([]models.ProcessedResource, error) {
	if cfg.HdhiveApiKey == "" { return nil, nil }
	if mType != "movie" && mType != "tv" { mType = "tv" }
	
	u := fmt.Sprintf("https://hdhive.com/api/open/resources/%s/%s", mType, id)
	slog.Info("🔎 [HDHive] 请求 API 列表", "url", u)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	
	var w models.HdhiveApiResponse
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &w)
	if !w.Success { return nil, fmt.Errorf("HDHive: %s", w.Message) }
	
	var hrs []models.HdhiveResource
	json.Unmarshal(w.Data, &hrs)
	slog.Info("📥 [API] 收到原始资源", "count", len(hrs))

	var results []models.ProcessedResource
	for _, hr := range hrs {
		is115 := (hr.PanType == "115") || (hr.PanType == "" && strings.Contains(strings.ToLower(hr.Title), "115"))
		if !is115 { continue }

		totalMB := utils.ParseSizeToMB(hr.ShareSize)
		count := utils.EstimateEpisodeCount(hr.Title)
		avgMB := totalMB / float64(count)
		tags := utils.UniqueTags(append(hr.VideoResolution, hr.Source...))
		
		disp := fmt.Sprintf("[HDHive] [%.1fG] %s", totalMB/1024, hr.Title)
		if count > 1 { disp = fmt.Sprintf("[HDHive][均%.1fG | 总%.1fG] %s", avgMB/1024, totalMB/1024, hr.Title) }
		
		// 🌟 新增：Remark 智能集成
		if hr.Remark != "" {
			remarkRunes := []rune(hr.Remark)
			if len(remarkRunes) > 30 {
				disp = fmt.Sprintf("%s | 📝 %s...", disp, string(remarkRunes[:28]))
			} else {
				disp = fmt.Sprintf("%s | 📝 %s", disp, hr.Remark)
			}
		}

		if hr.UnlockPoints > 0 && !hr.IsUnlocked { disp = fmt.Sprintf("💎 %dpt | %s", hr.UnlockPoints, disp) }
		
		res := models.ProcessedResource{Title: hr.Title, Link: "hdhive://" + hr.Slug, TotalMB: totalMB, AvgMB: avgMB, Display: disp, Tags: tags, HdhivePoints: hr.UnlockPoints}
		if hr.IsUnlocked { res.HdhivePoints = 0 }
		if utils.IsExcluded(hr.Title, cfg.ExcludeWords) { res.IsExcluded = true }
		results = append(results, res)
	}
	
	slog.Info("🎯 精准展示 115 资源数量", "count", len(results))
	return results, nil
}

func GetHdhiveMe(cfg models.BotConfig) (*models.HdhiveMeData, error) {
	if cfg.HdhiveApiKey == "" { return nil, fmt.Errorf("API Key missing") }
	req, _ := http.NewRequest("GET", "https://hdhive.com/api/open/me", nil)
	req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var w models.HdhiveApiResponse
	json.NewDecoder(resp.Body).Decode(&w)
	if !w.Success { return nil, fmt.Errorf(w.Message) }
	var d models.HdhiveMeData
	json.Unmarshal(w.Data, &d)
	return &d, nil
}

func UnlockHdhive(slug string, cfg models.BotConfig) (string, error) {
	payload := map[string]string{"slug": slug}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://hdhive.com/api/open/resources/unlock", bytes.NewBuffer(b))
	req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
	req.Header.Set("Content-Type", "application/json")
	slog.Info("🔓 解锁 Slug", "slug", slug)
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	var w models.HdhiveApiResponse
	json.NewDecoder(resp.Body).Decode(&w)
	if !w.Success { return "", fmt.Errorf(w.Message) }
	var d models.HdhiveUnlockData
	json.Unmarshal(w.Data, &d)
	res := d.FullUrl; if res == "" { res = d.Url }
	return res, nil
}

// === Media302 & Pansou ===

func DoPansouSearch(kw string, cfg models.BotConfig) ([]models.PansouItem, error) {
	if cfg.PansouUrl == "" { return nil, nil }
	reqBody, _ := json.Marshal(models.PansouRequest{Kw: kw, Res: "merge"})
	u := strings.TrimRight(cfg.PansouUrl, "/") + "/api/search"
	req, _ := http.NewRequest("POST", u, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if cfg.PansouUsername != "" {
		lUrl := strings.TrimRight(cfg.PansouUrl, "/") + "/api/auth/login"
		lb, _ := json.Marshal(models.PansouLoginReq{Username: cfg.PansouUsername, Password: cfg.PansouPassword})
		lr, err := utils.GetDirectClient().Post(lUrl, "application/json", bytes.NewBuffer(lb))
		if err == nil {
			var lResp models.PansouLoginResp; json.NewDecoder(lr.Body).Decode(&lResp)
			lr.Body.Close()
			if lResp.Token != "" { req.Header.Set("Authorization", "Bearer "+lResp.Token) }
		}
	}
	resp, err := utils.GetDirectClient().Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var w models.PansouApiResponse
	json.NewDecoder(resp.Body).Decode(&w)
	var final []models.PansouItem
	for dt, items := range w.Data.MergedByType {
		if dt != "115" && dt != "magnet" { continue }
		for i := range items {
			items[i].DiskType = dt
			if dt == "magnet" { items[i].Note = "[磁力] " + items[i].Note } else { items[i].Note = "[115] " + items[i].Note }
			final = append(final, items[i])
		}
	}
	return final, nil
}

func PushToMedia302(link string, cfg models.BotConfig) string {
	if cfg.Media302BaseUrl == "" { return "⚠️ 未配置Media302" }
	final := link
	if strings.HasPrefix(link, "hdhive://") {
		u, err := UnlockHdhive(strings.TrimPrefix(link, "hdhive://"), cfg)
		if err != nil { return "❌ 解锁失败: " + err.Error() }
		if !strings.Contains(u, "115") { return "❌ 拦截: 非 115 链接" }
		final = u
	}
	vals := url.Values{}
	vals.Add("folder", cfg.Media302Folder); vals.Add("token", cfg.Media302Token); vals.Add("url", final)
	u := fmt.Sprintf("%s/strm/api/task/save-share?%s", strings.TrimRight(cfg.Media302BaseUrl, "/"), vals.Encode())
	
	slog.Debug("📤 发送转存请求", "url", u)
	resp, err := utils.GetDirectClient().Get(u)
	if err != nil { return "❌ 请求失败" }
	defer resp.Body.Close()
	
	// 🌟 深度诊断：读取 Media302 的反馈
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	slog.Info("📥 Media302 响应反馈", "status", resp.StatusCode, "body", bodyStr)

	if resp.StatusCode == 200 { 
		if strings.Contains(bodyStr, "false") || strings.Contains(bodyStr, "error") {
			return "❌ Media302报错: " + bodyStr
		}
		return "✅ 115 推送成功" 
	}
	return "❌ 失败: " + bodyStr
}

func PushMagnetToOffline(magnet string, cfg models.BotConfig) string {
	if cfg.Media302BaseUrl == "" { return "⚠️ 未配置Media302" }
	u := fmt.Sprintf("%s/strm/api/task/batch-offline?token=%s", strings.TrimRight(cfg.Media302BaseUrl, "/"), cfg.Media302Token)
	folder := cfg.MagnetFolder; if folder == "" { folder = "离线下载" }
	p := map[string]string{"urls": magnet, "folder": folder}
	b, _ := json.Marshal(p)
	
	slog.Debug("📤 发送离线下载请求", "api", u)
	resp, err := utils.GetDirectClient().Post(u, "application/json", bytes.NewBuffer(b))
	if err != nil { return "❌ 请求失败" }
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	slog.Info("📥 Media302 离线反馈", "status", resp.StatusCode, "body", bodyStr)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 { 
		return "✅ 磁力已添加" 
	}
	return "❌ 离线失败: " + bodyStr
}
