package netease

// SearchResult represents the response from search API
type SearchResult struct {
	Result struct {
		Artists    []Artist `json:"artists"`
		ArtistCount int     `json:"artistCount"`
		Songs      []Song   `json:"songs"`
		SongCount  int      `json:"songCount"`
		Albums     []Album  `json:"albums"`
		AlbumCount int      `json:"albumCount"`
	} `json:"result"`
	Code int `json:"code"`
}

// Artist represents a NetEase artist
type Artist struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	PicURL     string   `json:"picUrl"`
	Img1v1URL  string   `json:"img1v1Url"`
	Alias      []string `json:"alias"`
	AlbumSize  int      `json:"albumSize"`
	Trans      string   `json:"trans"`
	MvSize     int      `json:"mvSize"`
	BriefDesc  string   `json:"briefDesc"`
}

// ArtistDetail represents the response from artist detail API
type ArtistDetail struct {
	Data struct {
		Artist    Artist `json:"artist"`
		VideoCount int   `json:"videoCount"`
		Identify  struct {
			ImageDesc string `json:"imageDesc"`
		} `json:"identify"`
		User interface{} `json:"user"`
	} `json:"data"`
	Code int `json:"code"`
}

// ArtistDesc represents the response from artist description API
type ArtistDesc struct {
	Introduction []struct {
		Ti  string `json:"ti"` // title
		Txt string `json:"txt"` // text content
	} `json:"introduction"`
	BriefDesc string `json:"briefDesc"`
	Code      int    `json:"code"`
}

// Song represents a NetEase song
type Song struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	Duration   int      `json:"duration"` // Duration in milliseconds
	Alias      []string `json:"alias"`
	CopyrightID int     `json:"copyrightId"`
	Status     int      `json:"status"`
	Fee        int      `json:"fee"`
	Mvid       int      `json:"mvid"`
}

// SongDetail represents the response from song detail API
type SongDetail struct {
	Songs []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Al    Album  `json:"al"` // album
		Ar    []struct { // artists
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"ar"`
		Dt int `json:"dt"` // duration in milliseconds
	} `json:"songs"`
	Code int `json:"code"`
}

// Album represents a NetEase album
type Album struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	PicURL      string  `json:"picUrl"`
	Artist      Artist  `json:"artist"`
	PublishTime int64   `json:"publishTime"`
	Size        int     `json:"size"`
	CopyrightID int     `json:"copyrightId"`
	Status      int     `json:"status"`
	BlurPicURL  string  `json:"blurPicUrl"`
	Description string  `json:"description"`
}

// AlbumDetail represents the response from album detail API
type AlbumDetail struct {
	Album Album  `json:"album"`
	Songs []Song `json:"songs"`
	Code  int    `json:"code"`
}

// TopSongs represents the response from artist top songs API
type TopSongs struct {
	Songs []Song `json:"songs"`
	More  bool   `json:"more"`
	Code  int    `json:"code"`
}

// ArtistTopSongsV2 represents the response from new artist top songs API
type ArtistTopSongsV2 struct {
	Songs []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Al   struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			PicURL string `json:"picUrl"`
		} `json:"al"`
		Ar []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"ar"`
		Dt int `json:"dt"` // duration in milliseconds
	} `json:"songs"`
	More bool `json:"more"`
	Code int  `json:"code"`
}

// Error represents a NetEase API error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}
