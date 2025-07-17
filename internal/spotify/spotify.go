package spotify

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

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

// GetRecommendationsWithAttributes gets recommendations with specific audio attributes
func (c *Client) GetRecommendationsWithAttributes(seedArtists []string, attributes *spotify.TrackAttributes, limit int) ([]Track, error) {
	ctx := context.Background()

	// Convert string IDs to spotify.ID and shuffle for variety
	var artistIDs []spotify.ID
	for _, id := range seedArtists {
		artistIDs = append(artistIDs, spotify.ID(id))
	}

	// Shuffle seed artists for variety
	rand.Shuffle(len(artistIDs), func(i, j int) {
		artistIDs[i], artistIDs[j] = artistIDs[j], artistIDs[i]
	})

	// Limit seed artists to maximum of 5 (Spotify API limit)
	if len(artistIDs) > 5 {
		artistIDs = artistIDs[:5]
	}

	seeds := spotify.Seeds{
		Artists: artistIDs,
	}

	recommendations, err := c.client.GetRecommendations(ctx, seeds, attributes, spotify.Limit(limit))
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

// GeneratePlaylistForArtists generates a smart, varied playlist based on user's favorite artists
func (c *Client) GeneratePlaylistForArtists(favoriteArtists []string, playlistSize int) ([]Track, error) {
	// Initialize random seed for variety
	rand.Seed(time.Now().UnixNano())
	
	log.Printf("🎵 Generating smart playlist for %d artists, target size: %d tracks", len(favoriteArtists), playlistSize)

	if len(favoriteArtists) == 0 {
		return nil, fmt.Errorf("no favorite artists provided")
	}

	var allTracks []Track
	
	// Strategy 1: Get some popular tracks from favorite artists (30% of playlist)
	popularTracksCount := max(playlistSize*3/10, 2)
	log.Printf("📈 Getting %d popular tracks from favorite artists", popularTracksCount)
	
	popularTracks, err := c.getPopularTracksFromArtists(favoriteArtists, popularTracksCount)
	if err != nil {
		log.Printf("⚠️ Failed to get popular tracks: %v", err)
	} else {
		allTracks = append(allTracks, popularTracks...)
		log.Printf("✅ Added %d popular tracks", len(popularTracks))
	}

	// Strategy 2: Smart recommendations with varied attributes (50% of playlist)
	recommendationsCount := max(playlistSize*5/10, 5)
	log.Printf("🤖 Getting %d smart recommendations", recommendationsCount)
	
	smartRecommendations, err := c.getSmartRecommendations(favoriteArtists, recommendationsCount)
	if err != nil {
		log.Printf("⚠️ Failed to get smart recommendations: %v", err)
	} else {
		allTracks = append(allTracks, smartRecommendations...)
		log.Printf("✅ Added %d smart recommendations", len(smartRecommendations))
	}

	// Strategy 3: Deep discovery through related artists (20% of playlist)
	discoveryCount := max(playlistSize*2/10, 2)
	log.Printf("🔍 Getting %d discovery tracks from related artists", discoveryCount)
	
	discoveryTracks, err := c.getDiscoveryTracks(favoriteArtists, discoveryCount)
	if err != nil {
		log.Printf("⚠️ Failed to get discovery tracks: %v", err)
	} else {
		allTracks = append(allTracks, discoveryTracks...)
		log.Printf("✅ Added %d discovery tracks", len(discoveryTracks))
	}

	// Remove duplicates and shuffle
	uniqueTracks := removeDuplicateTracks(allTracks)
	
	// Shuffle for variety
	rand.Shuffle(len(uniqueTracks), func(i, j int) {
		uniqueTracks[i], uniqueTracks[j] = uniqueTracks[j], uniqueTracks[i]
	})

	// Limit to requested size
	if len(uniqueTracks) > playlistSize {
		uniqueTracks = uniqueTracks[:playlistSize]
	}

	log.Printf("🎉 Generated playlist with %d unique tracks", len(uniqueTracks))
	return uniqueTracks, nil
}

// getPopularTracksFromArtists gets popular tracks from favorite artists with randomization
func (c *Client) getPopularTracksFromArtists(artistIDs []string, count int) ([]Track, error) {
	var tracks []Track
	tracksPerArtist := max(count/len(artistIDs), 1)
	
	// Shuffle artists for variety
	shuffledArtists := make([]string, len(artistIDs))
	copy(shuffledArtists, artistIDs)
	rand.Shuffle(len(shuffledArtists), func(i, j int) {
		shuffledArtists[i], shuffledArtists[j] = shuffledArtists[j], shuffledArtists[i]
	})

	for _, artistID := range shuffledArtists {
		topTracks, err := c.GetArtistTopTracks(artistID, "US")
		if err != nil {
			continue
		}

		// Shuffle top tracks to avoid always picking the same ones
		rand.Shuffle(len(topTracks), func(i, j int) {
			topTracks[i], topTracks[j] = topTracks[j], topTracks[i]
		})

		// Take random selection from top tracks
		limit := min(len(topTracks), tracksPerArtist)
		tracks = append(tracks, topTracks[:limit]...)

		if len(tracks) >= count {
			break
		}
	}

	return tracks, nil
}

// getSmartRecommendations uses Spotify's recommendation engine with varied attributes
func (c *Client) getSmartRecommendations(artistIDs []string, count int) ([]Track, error) {
	var tracks []Track
	
	// Create different recommendation "moods" for variety
	attributeVariants := []*spotify.TrackAttributes{
		// Energetic recommendations
		spotify.NewTrackAttributes().
			MinEnergy(0.6).
			MinDanceability(0.5).
			MinValence(0.4),
		// Chill recommendations  
		spotify.NewTrackAttributes().
			MaxEnergy(0.7).
			MinValence(0.3).
			TargetAcousticness(0.3),
		// Balanced recommendations
		spotify.NewTrackAttributes().
			TargetEnergy(0.5).
			TargetValence(0.5).
			TargetDanceability(0.6),
		// Discovery-focused (no specific attributes)
		nil,
	}

	// Get recommendations with different attributes
	tracksPerVariant := max(count/len(attributeVariants), 3)
	
	for _, attributes := range attributeVariants {
		variantTracks, err := c.GetRecommendationsWithAttributes(artistIDs, attributes, tracksPerVariant)
		if err != nil {
			log.Printf("Failed to get recommendations with specific attributes: %v", err)
			continue
		}
		
		tracks = append(tracks, variantTracks...)
		
		if len(tracks) >= count {
			break
		}
	}

	return tracks, nil
}

// getDiscoveryTracks finds new music through related artists
func (c *Client) getDiscoveryTracks(artistIDs []string, count int) ([]Track, error) {
	var tracks []Track
	var allRelatedArtists []Artist

	// Collect related artists from favorites
	for _, artistID := range artistIDs {
		related, err := c.GetRelatedArtists(artistID)
		if err != nil {
			continue
		}
		allRelatedArtists = append(allRelatedArtists, related...)
	}

	if len(allRelatedArtists) == 0 {
		return tracks, nil
	}

	// Shuffle related artists for discovery variety
	rand.Shuffle(len(allRelatedArtists), func(i, j int) {
		allRelatedArtists[i], allRelatedArtists[j] = allRelatedArtists[j], allRelatedArtists[i]
	})

	// Get tracks from related artists
	artistsToCheck := min(len(allRelatedArtists), 10) // Limit to avoid too many API calls
	tracksPerArtist := max(count/artistsToCheck, 1)

	for i := 0; i < artistsToCheck && len(tracks) < count; i++ {
		artistTracks, err := c.GetArtistTopTracks(allRelatedArtists[i].SpotifyID, "US")
		if err != nil {
			continue
		}

		// Shuffle and pick random tracks
		rand.Shuffle(len(artistTracks), func(i, j int) {
			artistTracks[i], artistTracks[j] = artistTracks[j], artistTracks[i]
		})

		limit := min(len(artistTracks), tracksPerArtist)
		tracks = append(tracks, artistTracks[:limit]...)
	}

	return tracks, nil
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