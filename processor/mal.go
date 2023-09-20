package processor

import (
	"context"
	"net/http"

	"anime-titles-map/external"
)

type MALProcessor struct {
	client *external.CachingMALClient
}

func NewMALProcessor(httpClient *http.Client, cacheDir string) *MALProcessor {
	return &MALProcessor{
		client: external.NewCachingMALClient(external.NewMALClient(httpClient), cacheDir),
	}
}

func (p *MALProcessor) FetchTitles(ctx context.Context) ([]*MediaTitle, error) {
	results, err := p.client.FetchMediaAll(ctx)
	if err != nil {
		return nil, err
	}

	var titles []*MediaTitle
	for _, result := range results {
		for _, media := range result.Data {
			title := NewMediaTitle()

			for _, t := range media.Titles {
				if t.Type == "Japanese" {
					title.Japanese.Add(t.Title)
				} else {
					title.English.Add(t.Title)
				}
			}

			if title.Japanese.Cardinality() == 0 {
				continue
			}

			titles = append(titles, title)
		}
	}

	return titles, nil
}

var _ Service = &MALProcessor{}
