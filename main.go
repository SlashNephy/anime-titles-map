package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/goccy/go-json"

	_ "anime-titles-map/logger"
	"anime-titles-map/processor"
)

func main() {
	ctx := context.Background()
	httpClient := &http.Client{}

	anilist := processor.NewAniListProcessor(httpClient, "cache/anilist")
	mal := processor.NewMALProcessor(httpClient, "cache/mal")

	anilistTitles, err := anilist.FetchTitles(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch AniList titles", slog.Any("err", err))
		return
	}

	malTitles, err := mal.FetchTitles(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch MAL titles", slog.Any("err", err))
		return
	}

	titles, err := processor.MergeTitles(append(anilistTitles, malTitles...)...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to merge titles", slog.Any("err", err))
		return
	}

	content, err := json.Marshal(titles)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal titles", slog.Any("err", err))
		return
	}

	outputPath := "dist/japanese.json"
	if err = os.WriteFile(outputPath, content, 0644); err != nil {
		slog.ErrorContext(ctx, "failed to write japanese.json", slog.Any("err", err))
		return
	}
}
