package jackett

import (
	"context"
	"os"

	"github.com/Muxcore-Media/core/pkg/contracts"
)

type Module struct {
	client *Client
	addr   string
	apiKey string
}

func NewModule() *Module {
	return &Module{
		addr:   envOrDefault("MUXCORE_JACKETT_ADDR", "http://localhost:9117"),
		apiKey: os.Getenv("MUXCORE_JACKETT_APIKEY"),
	}
}

func (m *Module) Info() contracts.ModuleInfo {
	return contracts.ModuleInfo{
		ID:           "indexer-jackett",
		Name:         "Jackett",
		Version:      "1.0.0",
		Kinds:        []contracts.ModuleKind{contracts.ModuleKindIndexer},
		Description:  "Jackett/Prowlarr indexer connector via Torznab API",
		Author:       "MuxCore",
		Capabilities: []string{"indexer.torznab", "indexer.newznab", "indexer.search"},
	}
}

func (m *Module) Init(ctx context.Context) error {
	m.client = NewClient(m.addr, m.apiKey)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }
func (m *Module) Stop(ctx context.Context) error  { return nil }

func (m *Module) Health(ctx context.Context) error {
	_, err := m.client.Capabilities(ctx)
	return err
}

func (m *Module) Name() string {
	return m.client.Name()
}

func (m *Module) Search(ctx context.Context, query contracts.SearchQuery) ([]contracts.IndexerResult, error) {
	return m.client.Search(ctx, query)
}

func (m *Module) Capabilities(ctx context.Context) ([]string, error) {
	return m.client.Capabilities(ctx)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
