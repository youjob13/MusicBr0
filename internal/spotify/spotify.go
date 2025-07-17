package spotify

import (
	"context"
	"fmt"
	"log"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// Client wraps Spotify API client
type Client struct {
	client *spotify.Client
}

// Track represents a simplified track for playlists
type Track struct {
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	SpotifyURL  string `json:"spotify_url"`
	Duration    int    `json:"duration_ms"`
}

// Artist represents a simplified artist
type Artist struct {
	Name      string `json:"name"`
	SpotifyID string `json:"spotify_id"`
	Genres    []string `json:"genres"`
}

// New creates a new Spotify client using client credentials flow
func New(clientID, clientSecret string) (*Client, error) {
	ctx := context.Background()

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get token: %w", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	return &Client{client: client}, nil
}

// SearchArtists searches for artists by name
func (c *Client) SearchArtists(query string, limit int) ([]Artist, error) {
	ctx := context.Background()

	results, err := c.client.Search(ctx, query, spotify.SearchTypeArtist, spotify.Limit(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to search artists: %w", err)
	}

	var artists []Artist
	if results.Artists != nil {
		for _, artist := range results.Artists.Artists {
			var genres []string
			for _, genre := range artist.Genres {
				genres = append(genres, genre)
			}

			artists = append(artists, Artist{
				Name:      artist.Name,
				SpotifyID: string(artist.ID),
				Genres:    genres,
			})
		}
	}

	return artists, nil
}

// GetArtistTopTracks gets top tracks for an artist
func (c *Client) GetArtistTopTracks(artistID string, country string) ([]Track, error) {
	ctx := context.Background()

	tracks, err := c.client.GetArtistsTopTracks(ctx, spotify.ID(artistID), "US")
	if err != nil {
		return nil, fmt.Errorf("failed to get artist top tracks: %w", err)
	}

	var result []Track
	for _, track := range tracks {
		var artistNames []string
		for _, artist := range track.Artists {
			artistNames = append(artistNames, artist.Name)
		}

		result = append(result, Track{
			Name:       track.Name,
			Artist:     artistNames[0], // Use primary artist
			Album:      track.Album.Name,
			SpotifyURL: track.ExternalURLs["spotify"],
			Duration:   int(track.Duration),
		})
	}

	return result, nil
}

// GetRelatedArtists gets artists related to the given artist
func (c *Client) GetRelatedArtists(artistID string) ([]Artist, error) {
	ctx := context.Background()

	relatedArtists, err := c.client.GetRelatedArtists(ctx, spotify.ID(artistID))
	if err != nil {
		return nil, fmt.Errorf("failed to get related artists: %w", err)
	}

	var artists []Artist
	for _, artist := range relatedArtists {
		var genres []string
		for _, genre := range artist.Genres {
			genres = append(genres, genre)
		}

		artists = append(artists, Artist{
			Name:      artist.Name,
			SpotifyID: string(artist.ID),
			Genres:    genres,
		})
	}

	return artists, nil
}

// GetRecommendations gets track recommendations based on seed artists
func (c *Client) GetRecommendations(seedArtists []string, limit int) ([]Track, error) {
	ctx := context.Background()

	// Convert string IDs to spotify.ID
	var artistIDs []spotify.ID
	for _, id := range seedArtists {
		artistIDs = append(artistIDs, spotify.ID(id))
	}

	// Limit seed artists to maximum of 5 (Spotify API limit)
	if len(artistIDs) > 5 {
		artistIDs = artistIDs[:5]
	}

	seeds := spotify.Seeds{
		Artists: artistIDs,
	}

	recommendations, err := c.client.GetRecommendations(ctx, seeds, &spotify.TrackAttributes{}, spotify.Limit(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	var tracks []Track
	for _, track := range recommendations.Tracks {
		var artistNames []string
		for _, artist := range track.Artists {
			artistNames = append(artistNames, artist.Name)
		}

		if len(artistNames) > 0 {
			tracks = append(tracks, Track{
				Name:       track.Name,
				Artist:     artistNames[0],
				Album:      track.Album.Name,
				SpotifyURL: track.ExternalURLs["spotify"],
				Duration:   int(track.Duration),
			})
		}
	}

	return tracks, nil
}

// GetArtistByID gets artist information by Spotify ID
func (c *Client) GetArtistByID(artistID string) (*Artist, error) {
	ctx := context.Background()

	artist, err := c.client.GetArtist(ctx, spotify.ID(artistID))
	if err != nil {
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	var genres []string
	for _, genre := range artist.Genres {
		genres = append(genres, genre)
	}

	return &Artist{
		Name:      artist.Name,
		SpotifyID: string(artist.ID),
		Genres:    genres,
	}, nil
}

// GeneratePlaylistForArtists generates a playlist based on user's favorite artists
func (c *Client) GeneratePlaylistForArtists(favoriteArtists []string, playlistSize int) ([]Track, error) {
	var allTracks []Track
	tracksPerArtist := max(1, playlistSize/(len(favoriteArtists)*2)) // Half from favorites, half from related

	log.Printf("Generating playlist for %d artists, %d tracks per artist", len(favoriteArtists), tracksPerArtist)

	// Get tracks from favorite artists
	for _, artistID := range favoriteArtists {
		tracks, err := c.GetArtistTopTracks(artistID, "US")
		if err != nil {
			log.Printf("Failed to get tracks for artist %s: %v", artistID, err)
			continue
		}

		// Take limited number of tracks from each artist
		limit := min(len(tracks), tracksPerArtist)
		allTracks = append(allTracks, tracks[:limit]...)
	}

	// Get tracks from related artists for discovery
	for _, artistID := range favoriteArtists {
		relatedArtists, err := c.GetRelatedArtists(artistID)
		if err != nil {
			log.Printf("Failed to get related artists for %s: %v", artistID, err)
			continue
		}

		// Take tracks from first few related artists
		for i, related := range relatedArtists {
			if i >= 2 { // Limit to 2 related artists per favorite
				break
			}

			tracks, err := c.GetArtistTopTracks(related.SpotifyID, "US")
			if err != nil {
				continue
			}

			// Take limited tracks from related artists
			limit := min(len(tracks), tracksPerArtist/2)
			allTracks = append(allTracks, tracks[:limit]...)
		}
	}

	// Remove duplicates and limit to requested size
	uniqueTracks := removeDuplicateTracks(allTracks)
	if len(uniqueTracks) > playlistSize {
		uniqueTracks = uniqueTracks[:playlistSize]
	}

	log.Printf("Generated playlist with %d unique tracks", len(uniqueTracks))
	return uniqueTracks, nil
}

// removeDuplicateTracks removes duplicate tracks based on name and artist
func removeDuplicateTracks(tracks []Track) []Track {
	seen := make(map[string]bool)
	var unique []Track

	for _, track := range tracks {
		key := fmt.Sprintf("%s-%s", track.Name, track.Artist)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, track)
		}
	}

	return unique
}

// Helper functions for Go versions that don't have min/max built-in
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
} 