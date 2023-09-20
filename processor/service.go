package processor

import (
	"context"

	mapSet "github.com/deckarep/golang-set/v2"
)

type Service interface {
	FetchTitles(ctx context.Context) ([]*MediaTitle, error)
}

type MediaTitle struct {
	English  mapSet.Set[string]
	Japanese mapSet.Set[string]
}

func NewMediaTitle() *MediaTitle {
	return &MediaTitle{
		English:  mapSet.NewSet[string](),
		Japanese: mapSet.NewSet[string](),
	}
}
