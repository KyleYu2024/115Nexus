package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	
	r, err := utils.GetDirectClient().Post(cfg.WebhookUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		slog.Error("❌ Webhook 发送失败", "err", err, "title", title)
		return
	}
	defer r.Body.Close()
	slog.Info("📡 Webhook 已发送", "title", title)
}

// === HDHive 身份与签到 ===

func DoHdhiveCheckin(cfg models.BotConfig) (string, error) {
	if !cfg.HdhiveCheckinEnabled { return "", nil }
	var finalMsg string
	var finalErr error
	start := time.Now()

	if cfg.HdhiveUser != "" && cfg.HdhivePass != "" {
		slog.Info("📅 HDHive Web 模拟签到模式", "user", cfg.HdhiveUser)
		jar, _ := cookiejar.New(nil)
		client := utils.GetProxyClient(cfg)
		client.Jar = jar
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		
		r1, _ := http.NewRequest("GET", "https://hdhive.com/login", nil)
		r1.Header.Set("User-Agent", ua)
		if _, err := client.Do(r1); err != nil {
			finalErr = fmt.Errorf("预访问失败: %w", err)
		} else {
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
					body, _ := io.ReadAll(resp2.Body)
					txt := string(body)
					msg := "签到请求已执行"
					reDesc := regexp.MustCompile(`"description":"([^"]+)"`)
					reMsg := regexp.MustCompile(`"message":"([^"]+)"`)
					if m := reDesc.FindStringSubmatch(txt); len(m) > 1 { msg = m[1] } else if m := reMsg.FindStringSubmatch(txt); len(m) > 1 { msg = m[1] }
					finalMsg = msg
					slog.Info("✅ Web 签到成功", "msg", msg, "cost", time.Since(start).String())
				} else { finalErr = errC }
			} else { finalErr = fmt.Errorf("登录失败: Status %d", resp.StatusCode) }
		}
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
			if !w.Success { finalErr = fmt.Errorf("%s", w.Message) }
			slog.Info("✅ API 签到反馈", "msg", finalMsg, "cost", time.Since(start).String())
		} else { finalErr = errA }
	}

	if finalMsg != "" || finalErr != nil {
		content := finalMsg
		if finalErr != nil { 
			content = "失败: " + finalErr.Error()
			slog.Error("❌ 签到任务异常", "err", finalErr)
		}
		SendWebhook(cfg, "影巢签到", content)
	}
	return finalMsg, finalErr
}

// === HDHive 资源获取 ===

func FetchHdhiveResources(id string, mType string, cfg models.BotConfig) ([]models.ProcessedResource, error) {
	if cfg.HdhiveApiKey == "" { return nil, fmt.Errorf("API Key 未配置") }
	if mType != "movie" && mType != "tv" { mType = "tv" }
	
	u := fmt.Sprintf("https://hdhive.com/api/open/resources/%s/%s", mType, id)
	slog.Info("🔎 获取资源列表", "mtype", mType, "id", id)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
	
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	
	var w models.HdhiveApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if !w.Success { return nil, fmt.Errorf("HDHive 报错: %s", w.Message) }
	
	var hrs []models.HdhiveResource
	json.Unmarshal(w.Data, &hrs)

	var results []models.ProcessedResource
	for _, hr := range hrs {
		is115 := (hr.PanType == "115") || (hr.PanType == "" && strings.Contains(strings.ToLower(hr.Title), "115"))
		if !is115 { continue }

		totalMB := utils.ParseSizeToMB(hr.ShareSize)
		count := utils.EstimateEpisodeCount(hr.Title)
		avgMB := totalMB / float64(count)
		allTags := append([]string{}, hr.VideoResolution...)
		allTags = append(allTags, hr.Source...)
		allTags = append(allTags, hr.SubtitleLanguage...)
		allTags = append(allTags, hr.SubtitleType...)
		allTags = append(allTags, hr.DiskTypes...)
		allTags = append(allTags, hr.CloudTypes...)
		tags := utils.UniqueTags(allTags)
		
		officialTag := ""
		if hr.IsOfficial { officialTag = "[官组] " }

		disp := fmt.Sprintf("💽 %s%s\n └ 📦 大小: %.1fGB", officialTag, hr.Title, totalMB/1024)
		if count > 1 { 
			disp = fmt.Sprintf("💽 %s%s\n └ 📦 大小: %.1fGB (均%.1fGB)", officialTag, hr.Title, totalMB/1024, avgMB/1024) 
		}

		if hr.Remark != "" { disp = fmt.Sprintf("%s\n └ 📝 备注: %s", disp, hr.Remark) }
		if hr.User.Nickname != "" { disp = fmt.Sprintf("%s\n └ 👤 发布: @%s", disp, hr.User.Nickname) }
		if hr.UnlockPoints > 0 && !hr.IsUnlocked { disp = fmt.Sprintf("%s\n └ 💎 需 %d pt 解锁", disp, hr.UnlockPoints) }
		
		res := models.ProcessedResource{Title: hr.Title, Link: "hdhive://" + hr.Slug, TotalMB: totalMB, AvgMB: avgMB, Display: disp, Tags: tags, HdhivePoints: hr.UnlockPoints}
		if hr.IsUnlocked { res.HdhivePoints = 0 }
		if utils.IsExcluded(hr.Title, cfg.ExcludeWords) { res.IsExcluded = true }
		results = append(results, res)
	}
	
	slog.Info("🎯 资源列表就绪", "count", len(results))
	return results, nil
}

func GetHdhiveMe(cfg models.BotConfig) (*models.HdhiveMeData, error) {
	if cfg.HdhiveApiKey == "" { return nil, fmt.Errorf("API Key 未配置") }
	req, _ := http.NewRequest("GET", "https://hdhive.com/api/open/me", nil)
	req.Header.Set("X-API-Key", cfg.HdhiveApiKey)
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var w models.HdhiveApiResponse
	json.NewDecoder(resp.Body).Decode(&w)
	if !w.Success { return nil, fmt.Errorf("%s", w.Message) }
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
	
	resp, err := utils.GetProxyClient(cfg).Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	
	var w models.HdhiveApiResponse
	json.NewDecoder(resp.Body).Decode(&w)
	if !w.Success { return "", fmt.Errorf("解锁失败: %s", w.Message) }
	
	var d models.HdhiveUnlockData
	json.Unmarshal(w.Data, &d)
	res := d.FullUrl; if res == "" { res = d.Url }
	return res, nil
}

func DoPansouSearch(kw string, cfg models.BotConfig) ([]models.PansouItem, error) {
	if cfg.PansouUrl == "" { return nil, fmt.Errorf("Pansou 地址未配置") }
	reqBody, _ := json.Marshal(models.PansouRequest{Kw: kw, Res: "merge"})
	u := strings.TrimRight(cfg.PansouUrl, "/") + "/api/search"
	req, _ := http.NewRequest("POST", u, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	if cfg.PansouUsername != "" {
		lUrl := strings.TrimRight(cfg.PansouUrl, "/") + "/api/auth/login"
		lb, _ := json.Marshal(models.PansouLoginReq{Username: cfg.PansouUsername, Password: cfg.PansouPassword})
		lr, err := utils.GetDirectClient().Post(lUrl, "application/json", bytes.NewBuffer(lb))
		if err == nil {
			defer lr.Body.Close()
			var lResp models.PansouLoginResp; json.NewDecoder(lr.Body).Decode(&lResp)
			if lResp.Token != "" { req.Header.Set("Authorization", "Bearer "+lResp.Token) }
		}
	}
	
	resp, err := utils.GetDirectClient().Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	
	var w models.PansouApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil { return nil, err }
	
	var final []models.PansouItem
	for dt, items := range w.Data.MergedByType {
		if dt != "115" && dt != "magnet" { continue }
		for i := range items {
			items[i].DiskType = dt
			if dt == "magnet" { items[i].Note = "[磁力] " + items[i].Note } else { items[i].Note = "[115] " + items[i].Note }
			final = append(final, items[i])
		}
	}
	slog.Info("🔍 Pansou 搜索结果", "kw", kw, "count", len(final))
	return final, nil
}

// === 推送 Worker 与 异步任务 ===

func StartPushWorker(ctx context.Context) {
	slog.Info("🚀 异步推送 Worker 已启动")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-models.PushTaskQueue:
			if !ok { return }
			
			// 频率限制
			<-ticker.C
			
			go func(t models.PushTask) {
				slog.Info("📦 正在处理任务", "id", t.ID, "retry", t.Retries)
				msg := ""
				if strings.HasPrefix(t.Link, "magnet:?") {
					msg = processPushMagnet(t.Link, t.Config)
				} else {
					msg = processPushToMedia302(t.Link, t.Config)
				}
				
				if strings.Contains(msg, "✅") {
					slog.Info("✅ 任务成功", "id", t.ID, "msg", msg)
				} else {
					if t.Retries < 3 {
						t.Retries++
						delay := time.Duration(t.Retries*t.Retries) * 10 * time.Second
						slog.Warn("⚠️ 任务失败，准备重试", "id", t.ID, "err", msg, "retry", t.Retries)
						time.AfterFunc(delay, func() {
							select {
							case models.PushTaskQueue <- t:
							default:
								slog.Error("🔥 队列已满，重试丢弃", "id", t.ID)
							}
						})
					} else {
						slog.Error("🚨 任务最终失败", "id", t.ID, "err", msg)
					}
				}
			}(task)
			
		case <-ctx.Done():
			slog.Info("🛑 推送 Worker 收到停止信号")
			return
		}
	}
}

func processPushToMedia302(link string, cfg models.BotConfig) string {
	if cfg.Media302BaseUrl == "" { return "⚠️ 未配置Media302" }
	final := link
	if strings.HasPrefix(link, "hdhive://") {
		u, err := UnlockHdhive(strings.TrimPrefix(link, "hdhive://"), cfg)
		if err != nil { return "❌ 解锁失败: " + err.Error() }
		final = u
	}
	
	vals := url.Values{}
	vals.Add("folder", cfg.Media302Folder); vals.Add("token", cfg.Media302Token); vals.Add("url", final)
	u := fmt.Sprintf("%s/strm/api/task/save-share?%s", strings.TrimRight(cfg.Media302BaseUrl, "/"), vals.Encode())
	
	resp, err := utils.GetDirectClient().Get(u)
	if err != nil { return "❌ 请求异常: " + err.Error() }
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if resp.StatusCode == 200 { 
		if strings.Contains(bodyStr, "false") || strings.Contains(bodyStr, "error") {
			return "❌ Media302 业务报错: " + bodyStr
		}
		return "✅ 115 推送成功" 
	}
	return "❌ HTTP 状态错误: " + resp.Status
}

func processPushMagnet(magnet string, cfg models.BotConfig) string {
	if cfg.Media302BaseUrl == "" { return "⚠️ 未配置Media302" }
	u := fmt.Sprintf("%s/strm/api/task/batch-offline?token=%s", strings.TrimRight(cfg.Media302BaseUrl, "/"), cfg.Media302Token)
	folder := cfg.MagnetFolder; if folder == "" { folder = "离线下载" }
	p := map[string]string{"urls": magnet, "folder": folder}
	b, _ := json.Marshal(p)
	
	resp, err := utils.GetDirectClient().Post(u, "application/json", bytes.NewBuffer(b))
	if err != nil { return "❌ 请求异常: " + err.Error() }
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 { 
		return "✅ 磁力已添加" 
	}
	return "❌ 离线失败: " + bodyStr
}

func PushToMedia302(link string, cfg models.BotConfig) string {
	taskID := fmt.Sprintf("T%d", time.Now().UnixNano()%10000)
	task := models.PushTask{
		ID:      taskID,
		Link:    link,
		Config:  cfg,
		Retries: 0,
	}
	select {
	case models.PushTaskQueue <- task:
		slog.Info("🕒 任务已入队", "id", taskID)
		return "✅ 任务已提交"
	default:
		return "❌ 队列已满"
	}
}

func PushMagnetToOffline(magnet string, cfg models.BotConfig) string {
	return PushToMedia302(magnet, cfg)
}
