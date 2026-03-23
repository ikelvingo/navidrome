package netease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"

	"github.com/navidrome/navidrome/log"
)

// Default API base URLs - these are third-party NetEase Cloud Music API endpoints
var defaultAPIBaseURLs = []string{
	"https://apis.netstart.cn/music",
	// "https://apic.netstart.cn/music", 延时过高，暂时移除
	"https://ncm.zhenxin.me",
	"https://ncmapi.btwoa.com",
	"https://zm.wwoyun.cn",
}

var (
	ErrNotFound    = errors.New("netease: not found")
	ErrAPIError    = errors.New("netease: api error")
	ErrInvalidCode = errors.New("netease: invalid response code")
)

// LoadBalanceMode defines how to select API endpoints
type LoadBalanceMode int

const (
	// LoadBalanceModeRandom selects a random endpoint for each request
	LoadBalanceModeRandom LoadBalanceMode = iota
	// LoadBalanceModeRoundRobin cycles through endpoints in order
	LoadBalanceModeRoundRobin
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	httpDoer        httpDoer
	apiBaseURLs     []string
	loadBalanceMode LoadBalanceMode
	currentIndex    uint64 // Used for round-robin
}

func newClient(hc httpDoer, apiBaseURLs []string, mode LoadBalanceMode) *client {
	urls := apiBaseURLs
	if len(urls) == 0 {
		urls = defaultAPIBaseURLs
	}
	return &client{
		httpDoer:        hc,
		apiBaseURLs:     urls,
		loadBalanceMode: mode,
		currentIndex:    0,
	}
}

// getBaseURL returns the next API base URL based on load balance mode
func (c *client) getBaseURL() string {
	if len(c.apiBaseURLs) == 0 {
		return defaultAPIBaseURLs[0]
	}
	if len(c.apiBaseURLs) == 1 {
		return c.apiBaseURLs[0]
	}

	switch c.loadBalanceMode {
	case LoadBalanceModeRoundRobin:
		idx := atomic.AddUint64(&c.currentIndex, 1)
		return c.apiBaseURLs[int(idx)%len(c.apiBaseURLs)]
	case LoadBalanceModeRandom:
		fallthrough
	default:
		return c.apiBaseURLs[rand.Intn(len(c.apiBaseURLs))]
	}
}

// searchArtists searches for artists by name
func (c *client) searchArtists(ctx context.Context, name string, limit int) ([]Artist, error) {
	params := url.Values{}
	params.Add("keywords", name)
	params.Add("type", "100") // 100 = artist search
	params.Add("limit", strconv.Itoa(limit))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/cloudsearch", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result SearchResult
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	if len(result.Result.Artists) == 0 {
		return nil, ErrNotFound
	}

	return result.Result.Artists, nil
}

// searchSongs searches for songs by name
func (c *client) searchSongs(ctx context.Context, name string, limit int) ([]Song, error) {
	params := url.Values{}
	params.Add("keywords", name)
	params.Add("type", "1") // 1 = song search
	params.Add("limit", strconv.Itoa(limit))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/cloudsearch", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result SearchResult
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	if len(result.Result.Songs) == 0 {
		return nil, ErrNotFound
	}

	return result.Result.Songs, nil
}

// searchAlbums searches for albums by name
func (c *client) searchAlbums(ctx context.Context, name string, limit int) ([]Album, error) {
	params := url.Values{}
	params.Add("keywords", name)
	params.Add("type", "10") // 10 = album search
	params.Add("limit", strconv.Itoa(limit))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/cloudsearch", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result SearchResult
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	if len(result.Result.Albums) == 0 {
		return nil, ErrNotFound
	}

	return result.Result.Albums, nil
}

// getArtistDetail gets detailed information about an artist
func (c *client) getArtistDetail(ctx context.Context, artistID int) (*ArtistDetail, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(artistID))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/artist/detail", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result ArtistDetail
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	return &result, nil
}

// getArtistDesc gets artist description/biography
func (c *client) getArtistDesc(ctx context.Context, artistID int) (*ArtistDesc, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(artistID))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/artist/desc", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result ArtistDesc
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	return &result, nil
}

// getArtistTopSongs gets top songs for an artist
func (c *client) getArtistTopSongs(ctx context.Context, artistID int) (*ArtistTopSongsV2, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(artistID))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/artist/top/song", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result ArtistTopSongsV2
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	return &result, nil
}

// getAlbumDetail gets detailed information about an album
func (c *client) getAlbumDetail(ctx context.Context, albumID int) (*AlbumDetail, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(albumID))

	baseURL := c.getBaseURL()
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/album", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	var result AlbumDetail
	if err := c.makeRequest(req, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("%w: code %d", ErrInvalidCode, result.Code)
	}

	return &result, nil
}

func (c *client) makeRequest(req *http.Request, response any) error {
	// Add necessary headers for NetEase API
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0")
	req.Header.Set("Origin", "https://music.163.com")
	req.Header.Set("Referer", "https://music.163.com")

	log.Trace(req.Context(), fmt.Sprintf("Sending NetEase %s request", req.Method), "url", req.URL)
	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%w: http status %d", ErrAPIError, resp.StatusCode)
	}

	return json.Unmarshal(data, response)
}
