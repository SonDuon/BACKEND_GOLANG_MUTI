// internal/provider/ophim1/types.go

package ophim1

// searchResponse: Cấu trúc response của API /v1/api/tim-kiem
type searchResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    searchData `json:"data"`
}

type searchData struct {
	SeoOnPage       seoOnPage    `json:"seoOnPage"`
	BreadCrumb      []breadCrumb `json:"breadCrumb"`
	TitlePage       string       `json:"titlePage"`
	Items           []searchItem `json:"items"`
	Params          searchParams `json:"params"`
	TypeList        string       `json:"type_list"`
	AppDomainFront  string       `json:"APP_DOMAIN_FRONTEND"`
	AppDomainCDNImg string       `json:"APP_DOMAIN_CDN_IMAGE"`
}

type seoOnPage struct {
	OgType          string   `json:"og_type"`
	TitleHead       string   `json:"titleHead"`
	DescriptionHead string   `json:"descriptionHead"`
	OgImage         []string `json:"og_image"`
	OgURL           string   `json:"og_url"`
}

type breadCrumb struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	Position  int    `json:"position"`
}

type searchParams struct {
	TypeSlug       string   `json:"type_slug"`
	Keyword        string   `json:"keyword"`
	FilterCategory []string `json:"filterCategory"`
	FilterCountry  []string `json:"filterCountry"`
	FilterYear     string   `json:"filterYear"`
	FilterType     string   `json:"filterType"`
	SortField      string   `json:"sortField"`
	SortType       string   `json:"sortType"`
	Pagination     struct {
		TotalItems        int64 `json:"totalItems"`
		TotalItemsPerPage int   `json:"totalItemsPerPage"`
		CurrentPage       int   `json:"currentPage"`
		PageRanges        int   `json:"pageRanges"`
	} `json:"pagination"`
}

type searchItem struct {
	Name             string         `json:"name"`
	OriginName       string         `json:"origin_name"`
	AlternativeNames []string       `json:"alternative_names"`
	Slug             string         `json:"slug"`
	Type             string         `json:"type"` // "single", "series", "hoathinh"
	ThumbURL         string         `json:"thumb_url"`
	PosterURL        string         `json:"poster_url"`
	SubDocquyen      bool           `json:"sub_docquyen"`
	Chieurap         bool           `json:"chieurap"`
	Time             string         `json:"time"`
	EpisodeCurrent   string         `json:"episode_current"`
	Quality          string         `json:"quality"`
	Lang             string         `json:"lang"`
	LangKey          []string       `json:"lang_key"`
	Year             int            `json:"year"`
	Category         []categoryItem `json:"category"`
	Country          []countryItem  `json:"country"`
	TMDB             tmdbInfo       `json:"tmdb"`
	IMDB             imdbInfo       `json:"imdb"`
	LastEpisodes     []lastEpisode  `json:"last_episodes"`
	Modified         modifiedInfo   `json:"modified"`
	ID               string         `json:"_id"`
}

type categoryItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type countryItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tmdbInfo struct {
	Type        string  `json:"type"`
	ID          string  `json:"id"`
	Season      *int    `json:"season"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

type imdbInfo struct {
	ID          string  `json:"id"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

type lastEpisode struct {
	ServerName string `json:"server_name"`
	IsAI       bool   `json:"is_ai"`
	Name       string `json:"name"`
}

type modifiedInfo struct {
	Time string `json:"time"`
}

type detailResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    detailData `json:"data"`
}

type detailData struct {
	SeoOnPage       detailSeoOnPage `json:"seoOnPage"`
	BreadCrumb      []breadCrumb    `json:"breadCrumb"`
	Params          detailParams    `json:"params"`
	Item            detailItem      `json:"item"` // ✅ Singular: 1 phim
	AppDomainCDNImg string          `json:"APP_DOMAIN_CDN_IMAGE"`
}

type detailSeoOnPage struct {
	OgType          string   `json:"og_type"`
	TitleHead       string   `json:"titleHead"`
	DescriptionHead string   `json:"descriptionHead"`
	OgImage         []string `json:"og_image"`
	OgURL           string   `json:"og_url"`
}

type detailParams struct {
	Slug string `json:"slug"`
}

// detailItem: Chứa tất cả thông tin chi tiết của 1 phim
type detailItem struct {
	// 🔑 ID & Source
	ID         string `json:"_id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	OriginName string `json:"origin_name"`

	// 🎬 Metadata
	Type        string `json:"type"`    // "single", "series"
	Status      string `json:"status"`  // "completed", "ongoing"
	Content     string `json:"content"` // Overview (có HTML)
	ThumbURL    string `json:"thumb_url"`
	PosterURL   string `json:"poster_url"`
	BackdropURL string `json:"backdrop_url,omitempty"`
	TrailerURL  string `json:"trailer_url"`
	Time        string `json:"time"` // "138 phút"
	Year        int    `json:"year"`
	View        int    `json:"view"`
	Quality     string `json:"quality"`
	Lang        string `json:"lang"`

	// 📺 Episodes
	EpisodeCurrent string        `json:"episode_current"`
	EpisodeTotal   string        `json:"episode_total"`
	Episodes       []episodeItem `json:"episodes"` // ✅ Array với cấu trúc mới

	// 🏷️ Categories
	Category []categoryItem `json:"category"`
	Country  []countryItem  `json:"country"`
	Actor    []string       `json:"actor"`
	Director []string       `json:"director"`

	// 🌐 External IDs
	TMDB tmdbInfo `json:"tmdb"`
	IMDB imdbInfo `json:"imdb"`

	// 📅 Timestamps
	Created  timestampInfo `json:"created"`
	Modified timestampInfo `json:"modified"`

	// 🔐 Flags
	IsCopyright bool `json:"is_copyright"`
	SubDocquyen bool `json:"sub_docquyen"`
	Chieurap    bool `json:"chieurap"`

	// 🔄 Other
	AlternativeNames []string `json:"alternative_names"`
	LangKey          []string `json:"lang_key"`
	Notify           string   `json:"notify"`
	Showtimes        string   `json:"showtimes"`
}

// episodeItem: Cấu trúc MỚI - server_data là array of object
type episodeItem struct {
	ServerName string           `json:"server_name"`
	IsAI       bool             `json:"is_ai"`
	ServerData []serverDataItem `json:"server_data"` // ✅ Array, không phải map
}

// serverDataItem: Cấu trúc của mỗi server trong server_data array
type serverDataItem struct {
	Name      string `json:"name"` // "Full", "Tập 1", ...
	Slug      string `json:"slug"` // "full", "tap-1", ...
	Filename  string `json:"filename"`
	LinkEmbed string `json:"link_embed"` // Embed URL
	LinkM3u8  string `json:"link_m3u8"`  // HLS URL (ưu tiên dùng)
}

type timestampInfo struct {
	Time string `json:"time"`
}

