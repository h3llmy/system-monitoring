package response

// hold data of itemcount obtained from jellyfin service.
type LibrariesItemCountResponse struct {
	MovieCount  int64 `json:"MovieCount"`
	SeriesCount int64 `json:"SeriesCount"`
	SongCount   int64 `json:"SongCount"`
}
