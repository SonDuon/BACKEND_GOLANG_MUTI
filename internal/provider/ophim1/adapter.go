package ophim1

import (
	"context"
	"fmt"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/pkg/logger"
)

type Adapter struct {
	cfg    Config
	client *client
	mapper *mapper
	cache  *metaCache
	log    *logger.Logger
}

func New(cfg Config, log *logger.Logger) *Adapter {
	return &Adapter{
		cfg:    cfg,
		client: newClient(cfg.BaseURL, cfg.Timeout),
		mapper: newMapper(),
		cache:  newMetaCache(30 * time.Minute),
		log:    log,
	}
}

// ✅ Implement Interface
func (a *Adapter) Name() string                         { return "ophim1" }
func (a *Adapter) Priority() int                        { return a.cfg.Priority }
func (a *Adapter) IsAvailable(ctx context.Context) bool { return a.client.ping(ctx) }

func (a *Adapter) Search(ctx context.Context, params *provider.SearchParams) (*provider.SearchResult, error) {
	raw, err := a.client.search(ctx, params.Keyword, params.Page, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("%w: search ophim1: %v", provider.ErrAPIResponse, err)
	}

	// ✅ Lấy items từ raw.Data.Items
	items := a.mapper.toMovieDTOs(raw.Data.Items, a.Name())
	total := raw.Data.Params.Pagination.TotalItems

	return &provider.SearchResult{
		Items:   items,
		Total:   total,
		Page:    params.Page,
		Limit:   params.Limit,
		HasMore: len(items) == params.Limit,
	}, nil
}

func (a *Adapter) GetByExternalID(ctx context.Context, externalID string) (*provider.MovieDTO, error) {
	if cached := a.cache.get(externalID); cached != nil {
		return cached, nil
	}

	raw, err := a.client.detail(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("%w: detail ophim1: %v", provider.ErrAPIResponse, err)
	}

	dto := a.mapper.toMovieDTO(raw, a.Name())
	a.cache.set(externalID, dto)
	return dto, nil
}

func (a *Adapter) GetList(ctx context.Context, page, limit int, sortBy string) (*provider.SearchResult, error) {
	endpoint := sortToEndpoint(sortBy)
	raw, err := a.client.list(ctx, endpoint, page, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: list ophim1: %v", provider.ErrAPIResponse, err)
	}

	items := a.mapper.toMovieDTOs(raw.Data.Items, a.Name())
	total := raw.Data.Params.Pagination.TotalItems

	return &provider.SearchResult{
		Items:   items,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: len(items) == limit,
	}, nil
}

func (a *Adapter) GetStreamingLinks(ctx context.Context, movieExternalID, episodeExternalID string) (*provider.StreamingDTO, error) {
	raw, err := a.client.detail(ctx, movieExternalID)
	if err != nil {
		return nil, fmt.Errorf("%w: fetch detail for streaming: %v", provider.ErrAPIResponse, err)
	}

	sources := a.mapper.toVideoSources(raw, episodeExternalID, a.Name())
	if len(sources) == 0 {
		return nil, provider.ErrStreamingLinksNotFound
	}

	// ✅ Lấy dữ liệu từ Item
	item := raw.Data.Item
	durationSeconds := parseDuration(item.Time)

	return &provider.StreamingDTO{
		MovieID:   movieExternalID,
		EpisodeID: episodeExternalID,
		Title:     item.Name,
		Sources:   sources,
		Thumbnail: joinImageURL(item.ThumbURL),
		Duration:  durationSeconds,
		ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
		Metadata: map[string]any{
			"source":     a.Name(),
			"fetched_at": time.Now().Unix(),
		},
	}, nil
}

func (a *Adapter) RefreshStreamingLinks(ctx context.Context, movieExternalID, episodeExternalID string) (*provider.StreamingDTO, error) {
	a.cache.invalidate(movieExternalID)
	return a.GetStreamingLinks(ctx, movieExternalID, episodeExternalID)
}
