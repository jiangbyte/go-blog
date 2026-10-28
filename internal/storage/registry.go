package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jiangbyte/go-blog/internal/config"
)

// Registry 按 provider 缓存客户端，支持默认引擎与单次覆盖。
type Registry struct {
	settings config.StorageSettings
	mu       sync.Mutex
	cache    map[string]ObjectStorage
}

func NewRegistry(settings config.StorageSettings) *Registry {
	return &Registry{
		settings: settings,
		cache:    make(map[string]ObjectStorage),
	}
}

func (r *Registry) DefaultProvider() (string, error) {
	return EngineToProvider(r.settings.DefaultEngine)
}

func (r *Registry) ResolveProvider(providerOrEngine string) (string, error) {
	if strings.TrimSpace(providerOrEngine) == "" {
		return r.DefaultProvider()
	}
	return EngineToProvider(providerOrEngine)
}

func (r *Registry) Get(providerOrEngine string) (ObjectStorage, string, error) {
	provider, err := r.ResolveProvider(providerOrEngine)
	if err != nil {
		return nil, "", err
	}
	cfg, err := engineConfig(r.settings, provider)
	if err != nil {
		return nil, "", err
	}
	if !cfg.Configured() {
		return nil, "", fmt.Errorf("存储引擎未配置完整凭据: %s", ProviderToEngine(provider))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if st, ok := r.cache[provider]; ok {
		return st, provider, nil
	}
	st, err := r.build(provider, cfg)
	if err != nil {
		return nil, "", err
	}
	r.cache[provider] = st
	return st, provider, nil
}

func (r *Registry) build(provider string, cfg config.EngineConfig) (ObjectStorage, error) {
	switch provider {
	case ProviderMinIO:
		return NewMinIO(cfg, r.settings.PresignExpireSeconds)
	case ProviderQiniu:
		return NewQiniu(cfg, r.settings.PresignExpireSeconds)
	default:
		return nil, fmt.Errorf("不支持的存储: %s", provider)
	}
}

func (r *Registry) Status() Status {
	defaultEngine := strings.ToUpper(strings.TrimSpace(r.settings.DefaultEngine))
	if defaultEngine == "" {
		defaultEngine = EngineMinIO
	}
	engines := []struct {
		engine   string
		provider string
	}{
		{EngineMinIO, ProviderMinIO},
		{EngineQiniu, ProviderQiniu},
	}
	infos := make([]EngineInfo, 0, len(engines))
	for _, e := range engines {
		cfg, _ := engineConfig(r.settings, e.provider)
		var bucket any
		var endpoint any
		if cfg.Bucket != "" {
			bucket = cfg.Bucket
		}
		if cfg.Endpoint != "" {
			endpoint = cfg.Endpoint
		}
		infos = append(infos, EngineInfo{
			Engine:       e.engine,
			Provider:     e.provider,
			DisplayName:  DisplayName(e.provider),
			Configured:   cfg.Configured(),
			IsDefault:    e.engine == defaultEngine,
			Bucket:       bucket,
			Endpoint:     endpoint,
			BucketPublic: cfg.BucketPublic,
		})
	}
	return Status{
		DefaultEngine:        defaultEngine,
		UploadMaxBytes:       r.settings.UploadMaxBytes,
		PresignExpireSeconds: r.settings.PresignExpireSeconds,
		Engines:              infos,
		Hint:                 "请在服务器 .env 中修改 STORAGE_* / DEFAULT_FILE_ENGINE 后重启服务",
	}
}

func (r *Registry) Settings() config.StorageSettings {
	return r.settings
}
