package provider

import "errors"

// ─────────────────────────────────────
// 🎯 Provider Module Errors
// ─────────────────────────────────────

var (
	// ErrProviderNotFound: Khi không tìm thấy provider theo tên
	ErrProviderNotFound = errors.New("provider not found")

	// ErrProviderUnavailable: Khi provider health check fail
	ErrProviderUnavailable = errors.New("provider is unavailable")

	// ErrMovieNotFound: Khi không tìm thấy phim theo external_id
	ErrMovieNotFound = errors.New("movie not found in provider")

	// ErrStreamingLinksNotFound: Khi không lấy được link xem
	ErrStreamingLinksNotFound = errors.New("streaming links not found")

	// ErrInvalidExternalID: Khi external_id format không đúng
	ErrInvalidExternalID = errors.New("invalid external_id format")

	// ErrAPIResponse: Khi API bên thứ 3 trả về lỗi
	ErrAPIResponse = errors.New("external API returned error")
)

// IsProviderError: Helper check lỗi có phải từ provider không
func IsProviderError(err error) bool {
	return errors.Is(err, ErrProviderNotFound) ||
		errors.Is(err, ErrProviderUnavailable) ||
		errors.Is(err, ErrMovieNotFound) ||
		errors.Is(err, ErrStreamingLinksNotFound)
}
