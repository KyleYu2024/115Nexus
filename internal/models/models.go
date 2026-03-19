package models

import (
	"encoding/json"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotConfig 全局配置
type BotConfig struct {
	TgToken string `json:"tg_token"`

	TmdbApiKey string `json:"tmdb_api_key"`

	PansouUrl      string `json:"pansou_url"`
	PansouUsername string `json:"pansou_username"`
	PansouPassword string `json:"pansou_password"`

	Media302BaseUrl string `json:"media302_base_url"`
	Media302Token   string `json:"media302_token"`
	Media302Folder  string `json:"media302_folder"`
	MagnetFolder    string `json:"magnet_folder"`

	HdhiveApiKey         string `json:"hdhive_api_key"`
	HdhiveUser           string `json:"hdhive_user"`
	HdhivePass           string `json:"hdhive_pass"`
	HdhiveGamblerMode    bool   `json:"hdhive_gambler_mode"`
	HdhiveCheckinEnabled bool   `json:"hdhive_checkin_enabled"`
	HdhiveCheckinCron    string `json:"hdhive_checkin_cron"`

	WebhookUrl   string `json:"webhook_url"`
	ProxyUrl     string `json:"proxy_url"`
	ExcludeWords string `json:"exclude_words"`

	MovieMinSize int `json:"movie_min_size"`
	MovieMaxSize int `json:"movie_max_size"`
	TvMinSize    int `json:"tv_min_size"`
	TvMaxSize    int `json:"tv_max_size"`
}

type MovieMetadata struct {
	Id           interface{} `json:"id"`
	Title        string      `json:"title"`
	Name         string      `json:"name"`
	ReleaseDate  string      `json:"release_date"`
	FirstAirDate string      `json:"first_air_date"`
	MediaType    string      `json:"media_type"`
}

type TmdbSearchResponse struct {
	Results []MovieMetadata `json:"results"`
}

type ProcessedResource struct {
	Title        string   `json:"title"`
	Link         string   `json:"link"`
	TotalMB      float64  `json:"total_mb"`
	AvgMB        float64  `json:"avg_mb"`
	Display      string   `json:"display"`
	IsExcluded   bool     `json:"is_excluded"`
	Tags         []string `json:"tags"`
	HdhivePoints int      `json:"hdhive_points"`
}

type HdhiveApiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type HdhiveResource struct {
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	PanType          string   `json:"pan_type"`
	ShareSize        string   `json:"share_size"`
	Remark           string   `json:"remark"`
	IsOfficial       bool     `json:"is_official"` // 🌟 新增：官组标识
	User             struct {
		Nickname string `json:"nickname"`
	} `json:"user"` // 🌟 新增：发布者信息
	DiskTypes        []string `json:"disk_types"`
	CloudTypes       []string `json:"cloud_types"`
	VideoResolution  []string `json:"video_resolution"`
	Source           []string `json:"source"`
	SubtitleLanguage []string `json:"subtitle_language"`
	SubtitleType     []string `json:"subtitle_type"`
	UnlockPoints     int      `json:"unlock_points"`
	IsUnlocked       bool     `json:"is_unlocked"`
}

type HdhiveUnlockData struct {
	Url     string `json:"url"`
	FullUrl string `json:"full_url"`
}

type HdhiveMeData struct {
	Nickname string `json:"nickname"`
	IsVip    bool   `json:"is_vip"`
	UserMeta struct {
		Points int `json:"points"`
	} `json:"user_meta"`
	VipExpirationDate string `json:"vip_expiration_date"`
}

type PansouRequest struct {
	Kw  string `json:"kw"`
	Res string `json:"res,omitempty"`
}

type PansouApiResponse struct {
	Data struct {
		MergedByType map[string][]PansouItem `json:"merged_by_type"`
	} `json:"data"`
}

type PansouItem struct {
	Url      string   `json:"url"`
	Note     string   `json:"note"`
	Source   string   `json:"source"`
	Datetime string   `json:"datetime"`
	DiskType string   `json:"disk_type"`
}

type PansouLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PansouLoginResp struct {
	Token string `json:"token"`
}

type SearchSession struct {
	Keyword string
	Items   []MovieMetadata
	Time    time.Time
}

type ResourceSession struct {
	Title string
	Items []ProcessedResource
	Time  time.Time
}

type PansouSession struct {
	Keyword string
	Items   []PansouItem
	Time    time.Time
}

type BatchOfflineRequest struct {
	Urls   string `json:"urls"`
	Folder string `json:"folder"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PushTask struct {
	ID      string
	Link    string
	Config  BotConfig
	Retries int
}

var (
	CurrentBot    *tgbotapi.BotAPI
	globalConfig  BotConfig
	configMutex   sync.RWMutex
	SearchCache   sync.Map
	ResourceCache sync.Map
	PansouCache   sync.Map
	WebUser       string
	WebPassword   string
	PushTaskQueue = make(chan PushTask, 100)
)

func GetConfig() BotConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

func UpdateConfig(cfg BotConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}
