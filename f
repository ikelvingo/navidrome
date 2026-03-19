package netease

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/core/agents"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/utils/cache"
)

const (
	neteaseAgentName       = "netease"
	neteaseArtistSearchLimit = 50
	neteaseAlbumSearchLimit  = 50
)

// Image size constants for NetEase
const (
	neteaseImageSizeXL     = 1024
	neteaseImageSizeLarge  = 512
	neteaseImageSizeMedium = 256
	neteaseImageSizeSmall  = 128
)

type neteaseAgent struct {
	dataStore model.DataStore
	client    *client
}

func neteaseConstructor(dataStore model.DataStore) agents.Interface {
	if !conf.Server.Netease.Enabled {
		return nil
	}

	agent := &neteaseAgent{
		dataStore: dataStore,
	}

	httpClient := &http.Client{
		Timeout: consts.DefaultHttpClientTimeOut,
	}
	cachedHttpClient := cache.NewHTTPClient(httpClient, consts.DefaultHttpClientTimeOut)

	// Parse API URLs from config
	var apiURLs []string
	if conf.Server.Netease.APIUrls != "" {
		for _, u := range strings.Split(conf.Server.Netease.APIUrls, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				apiURLs = append(apiURLs, u)
			}
		}
	}

	// Parse load balance mode from config
	mode := LoadBalanceModeRandom
	if strings.ToLower(conf.Server.Netease.LoadBalanceMode) == "roundrobin" {
		mode = LoadBalanceModeRoundRobin
	}

	agent.client = newClient(cachedHttpClient, apiURLs, mode)
	return agent
}

func (n *neteaseAgent) AgentName() string {
	return neteaseAgentName
}

// GetArtistImages returns artist images from NetEase
func (n *neteaseAgent) GetArtistImages(ctx context.Context, _, name, _ string) ([]agents.ExternalImage, error) {
	artist, err := n.searchArtist(ctx, name)
	if err != nil {
		if errors.Is(err, agents.ErrNotFound) {
			log.Warn(ctx, "Artist not found in NetEase", "artist", name)
		} else {
			log.Error(ctx, "Error calling NetEase", "artist", name, err)
		}
		return nil, err
	}

	var res []agents.ExternalImage

	// NetEase provides different image sizes through URL parameters
	// Format: picUrl?param=WxH
	if artist.PicURL != "" {
		// Add different sizes
		sizes := []struct {
			size   int
			suffix string
		}{
			{neteaseImageSizeXL, fmt.Sprintf("?param=%dx%d", neteaseImageSizeXL, neteaseImageSizeXL)},
			{neteaseImageSizeLarge, fmt.Sprintf("?param=%dx%d", neteaseImageSizeLarge, neteaseImageSizeLarge)},
			{neteaseImageSizeMedium, fmt.Sprintf("?param=%dx%d", neteaseImageSizeMedium, neteaseImageSizeMedium)},
			{neteaseImageSizeSmall, fmt.Sprintf("?param=%dx%d", neteaseImageSizeSmall, neteaseImageSizeSmall)},
		}

		for _, s := range sizes {
			res = append(res, agents.ExternalImage{
				URL:  artist.PicURL + s.suffix,
				Size: s.size,
			})
		}
	}

	// Also try img1v1Url if available
	if artist.Img1v1URL != "" && artist.Img1v1URL != artist.PicURL {
		res = append(res, agents.ExternalImage{
			URL:  artist.Img1v1URL + fmt.Sprintf("?param=%dx%d", neteaseImageSizeXL, neteaseImageSizeXL),
			Size: neteaseImageSizeXL,
		})
	}

	if len(res) == 0 {
		return nil, agents.ErrNotFound
	}

	return res, nil
}

// GetArtistBiography returns artist biography from NetEase
func (n *neteaseAgent) GetArtistBiography(ctx context.Context, _, name, _ string) (string, error) {
	artist, err := n.searchArtist(ctx, name)
	if err != nil {
		return "", err
	}

	// Try to get detailed artist description
	desc, err := n.client.getArtistDesc(ctx, artist.ID)
	if err != nil {
		log.Debug(ctx, "Failed to get artist description from NetEase", "artist", name, err)
		// Fall back to brief description if available
		if artist.BriefDesc != "" {
			return artist.BriefDesc, nil
		}
		return "", agents.ErrNotFound
	}

	// Build biography from introduction sections
	var bio strings.Builder
	if desc.BriefDesc != "" {
		bio.WriteString(desc.BriefDesc)
	}

	for _, intro := range desc.Introduction {
		if intro.Ti != "" && intro.Txt != "" {
			if bio.Len() > 0 {
				bio.WriteString("\n\n")
			}
			bio.WriteString(fmt.Sprintf("【%s】\n%s", intro.Ti, intro.Txt))
		}
	}

	if bio.Len() == 0 {
		return "", agents.ErrNotFound
	}

	return bio.String(), nil
}

// GetArtistURL returns the NetEase artist page URL
func (n *neteaseAgent) GetArtistURL(ctx context.Context, _, name, _ string) (string, error) {
	artist, err := n.searchArtist(ctx, name)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://music.163.com/#/artist?id=%d", artist.ID), nil
}

// GetArtistTopSongs returns top songs for an artist from NetEase
func (n *neteaseAgent) GetArtistTopSongs(ctx context.Context, _, artistName, _ string, count int) ([]agents.Song, error) {
	artist, err := n.searchArtist(ctx, artistName)
	if err != nil {
		return nil, err
	}

	topSongs, err := n.client.getArtistTopSongs(ctx, artist.ID)
	if err != nil {
		log.Error(ctx, "Error getting top songs from NetEase", "artist", artistName, err)
		return nil, err
	}

	var res []agents.Song
	for _, s := range topSongs.Songs {
		if len(res) >= count {
			break
		}

		var albumName string
		if s.Al.Name != "" {
			albumName = s.Al.Name
		}

		res = append(res, agents.Song{
			Name:     s.Name,
			Album:    albumName,
			Duration: uint32(s.Dt), // Already in milliseconds
		})
	}

	if len(res) == 0 {
		return nil, agents.ErrNotFound
	}

	return res, nil
}

// GetAlbumInfo returns album information from NetEase
func (n *neteaseAgent) GetAlbumInfo(ctx context.Context, name, artist, _ string) (*agents.AlbumInfo, error) {
	album, err := n.searchAlbum(ctx, name, artist)
	if err != nil {
		return nil, err
	}

	info := &agents.AlbumInfo{
		Name:        album.Name,
		Description: album.Description,
		URL:         fmt.Sprintf("https://music.163.com/#/album?id=%d", album.ID),
	}

	return info, nil
}

// GetAlbumImages returns album images from NetEase
func (n *neteaseAgent) GetAlbumImages(ctx context.Context, name, artist, _ string) ([]agents.ExternalImage, error) {
	album, err := n.searchAlbum(ctx, name, artist)
	if err != nil {
		return nil, err
	}

	if album.PicURL == "" {
		return nil, agents.ErrNotFound
	}

	var res []agents.ExternalImage
	sizes := []struct {
		size   int
		suffix string
	}{
		{neteaseImageSizeXL, fmt.Sprintf("?param=%dx%d", neteaseImageSizeXL, neteaseImageSizeXL)},
		{neteaseImageSizeLarge, fmt.Sprintf("?param=%dx%d", neteaseImageSizeLarge, neteaseImageSizeLarge)},
		{neteaseImageSizeMedium, fmt.Sprintf("?param=%dx%d", neteaseImageSizeMedium, neteaseImageSizeMedium)},
		{neteaseImageSizeSmall, fmt.Sprintf("?param=%dx%d", neteaseImageSizeSmall, neteaseImageSizeSmall)},
	}

	for _, s := range sizes {
		res = append(res, agents.ExternalImage{
			URL:  album.PicURL + s.suffix,
			Size: s.size,
		})
	}

	return res, nil
}

// Note: GetSimilarArtists is NOT implemented for NetEase as per requirements
// NetEase Cloud Music API does not provide a reliable similar artists endpoint

// searchArtist searches for an artist by name and returns the best match
func (n *neteaseAgent) searchArtist(ctx context.Context, name string) (*Artist, error) {
	artists, err := n.client.searchArtists(ctx, name, neteaseArtistSearchLimit)
	if errors.Is(err, ErrNotFound) || len(artists) == 0 {
		return nil, agents.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	log.Trace(ctx, "NetEase artists found", "count", len(artists), "searched_name", name)

	// Find best match - exact match first
	for _, a := range artists {
		if strings.EqualFold(a.Name, name) {
			log.Trace(ctx, "Found exact artist match", "name", a.Name, "id", a.ID)
			return &a, nil
		}
		// Also check aliases
		for _, alias := range a.Alias {
			if strings.EqualFold(alias, name) {
				log.Trace(ctx, "Found artist match via alias", "name", a.Name, "alias", alias, "id", a.ID)
				return &a, nil
			}
		}
	}

	// If no exact match, use the first result if it's a close match
	if len(artists) > 0 {
		log.Trace(ctx, "Using first artist result", "searched", name, "found", artists[0].Name)
		return &artists[0], nil
	}

	return nil, agents.ErrNotFound
}

// searchAlbum searches for an album by name and artist
func (n *neteaseAgent) searchAlbum(ctx context.Context, name, artist string) (*Album, error) {
	// Search with both album name and artist name for better accuracy
	searchQuery := name
	if artist != "" {
		searchQuery = fmt.Sprintf("%s %s", name, artist)
	}

	albums, err := n.client.searchAlbums(ctx, searchQuery, neteaseAlbumSearchLimit)
	if errors.Is(err, ErrNotFound) || len(albums) == 0 {
		return nil, agents.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	log.Trace(ctx, "NetEase albums found", "count", len(albums), "searched_name", name, "artist", artist)

	// Find best match
	for _, a := range albums {
		// Check if album name matches
		if strings.EqualFold(a.Name, name) {
			// If artist is provided, also check artist name
			if artist == "" || strings.EqualFold(a.Artist.Name, artist) {
				log.Trace(ctx, "Found exact album match", "name", a.Name, "artist", a.Artist.Name, "id", a.ID)
				return &a, nil
			}
		}
	}

	// If no exact match, use the first result
	if len(albums) > 0 {
		log.Trace(ctx, "Using first album result", "searched", name, "found", albums[0].Name)
		return &albums[0], nil
	}

	return nil, agents.ErrNotFound
}

func init() {
	conf.AddHook(func() {
		if conf.Server.Netease.Enabled {
			agents.Register(neteaseAgentName, func(ds model.DataStore) agents.Interface {
				a := neteaseConstructor(ds)
				if a != nil {
					return a
				}
				return nil
			})
		}
	})
}

// Ensure neteaseAgent implements required interfaces
var _ agents.Interface = (*neteaseAgent)(nil)
var _ agents.ArtistImageRetriever = (*neteaseAgent)(nil)
var _ agents.ArtistBiographyRetriever = (*neteaseAgent)(nil)
var _ agents.ArtistURLRetriever = (*neteaseAgent)(nil)
var _ agents.ArtistTopSongsRetriever = (*neteaseAgent)(nil)
var _ agents.AlbumInfoRetriever = (*neteaseAgent)(nil)
var _ agents.AlbumImageRetriever = (*neteaseAgent)(nil)
