package external

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

type AniListClient struct {
	graphqlClient *graphql.Client
}

func NewAniListClient(httpClient *http.Client) *AniListClient {
	return &AniListClient{
		graphqlClient: graphql.
			NewClient("https://graphql.anilist.co", httpClient).
			WithRequestModifier(func(req *http.Request) {
				req.Header.Set("User-Agent", "anime-titles-map (+https://github.com/SlashNephy/anime-titles-map)")
			}),
	}
}

type AniListMediaQuery struct {
	Page struct {
		Media []struct {
			Title struct {
				English *string `graphql:"english" json:"english"`
				Romaji  *string `graphql:"romaji" json:"romaji"`
				Native  *string `graphql:"native" json:"native"`
			} `graphql:"title" json:"title"`
		} `graphql:"media(sort: ID, type: ANIME, countryOfOrigin: JP)" json:"media"`
		PageInfo struct {
			HasNextPage bool `graphql:"hasNextPage" json:"hasNextPage"`
		} `graphql:"pageInfo" json:"pageInfo"`
	} `graphql:"Page(page: $page, perPage: 50)" json:"page"`
}

func (c *AniListClient) fetchMediaWithPage(ctx context.Context, page int) (*AniListMediaQuery, error) {
	var query AniListMediaQuery
	variables := map[string]any{
		"page": page,
	}

	if err := c.graphqlClient.Query(ctx, &query, variables); err != nil {
		var errs graphql.Errors
		if errors.As(err, &errs) && len(errs) > 0 {
			for _, e := range errs {
				if strings.Contains(e.Message, "429 Too Many Requests") {
					return nil, ErrRateLimited
				}
				if strings.Contains(e.Message, "500 Internal Server Error") {
					return nil, ErrServerError
				}
			}
		}

		return nil, err
	}

	slog.Info("[AniList] fetched media", slog.Int("page", page))
	return &query, nil
}

func (c *AniListClient) FetchMediaAll(ctx context.Context) ([]*AniListMediaQuery, error) {
	var results []*AniListMediaQuery
	page := 1

	defer func() {
		slog.Info("[AniList] last page", slog.Int("page", page))
	}()

	for {
		result, err := c.fetchMediaWithPage(ctx, page)
		if err != nil {
			if errors.Is(err, ErrRateLimited) {
				time.Sleep(time.Second)
				continue
			}

			return nil, err
		}

		results = append(results, result)

		if !result.Page.PageInfo.HasNextPage || len(result.Page.Media) == 0 {
			return results, nil
		}

		page++
	}
}

type CachingAniListClient struct {
	*AniListClient
	cacheDir string
}

func NewCachingAniListClient(client *AniListClient, cacheDir string) (*CachingAniListClient, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &CachingAniListClient{
		AniListClient: client,
		cacheDir:      cacheDir,
	}, nil
}

func (c *CachingAniListClient) FetchMediaAll(ctx context.Context) ([]*AniListMediaQuery, error) {
	var results []*AniListMediaQuery
	page := 1

	defer func() {
		slog.Info("[AniList] last page", slog.Int("page", page))
	}()

	for {
		var result AniListMediaQuery
		cachePath := fmt.Sprintf("%s/anilist-%d.json", c.cacheDir, page)

		found, err := loadCache(cachePath, &result)
		if err != nil {
			return nil, err
		}

		if found {
			slog.Debug("[AniList] use cached response", slog.Int("page", page), slog.String("path", cachePath))
		} else {
			r, err := c.fetchMediaWithPage(ctx, page)
			if err != nil {
				if errors.Is(err, ErrRateLimited) {
					time.Sleep(time.Second)
					continue
				}

				return nil, err
			}

			result = *r
			if err = saveCache(cachePath, &result); err != nil {
				return nil, err
			}
		}

		results = append(results, &result)

		if !result.Page.PageInfo.HasNextPage || len(result.Page.Media) == 0 {
			return results, nil
		}

		page++
	}
}
