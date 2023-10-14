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

	anilist, err := processor.NewAniListProcessor(httpClient, "cache/anilist")
	if err != nil {
		slog.ErrorContext(ctx, "failed to create AniList processor", slog.Any("err", err))
		panic(err)
	}

	mal, err := processor.NewMALProcessor(httpClient, "cache/mal")
	if err != nil {
		slog.ErrorContext(ctx, "failed to create MAL processor", slog.Any("err", err))
		panic(err)
	}

	anilistTitles, err := anilist.FetchTitles(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch AniList titles", slog.Any("err", err))
		panic(err)
	}

	malTitles, err := mal.FetchTitles(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch MAL titles", slog.Any("err", err))
		panic(err)
	}

	titles, err := processor.MergeTitles(append(anilistTitles, malTitles...)...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to merge titles", slog.Any("err", err))
		panic(err)
	}

	content, err := json.Marshal(titles)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal titles", slog.Any("err", err))
		panic(err)
	}

	outputPath := "dist/japanese.json"
	if err = os.WriteFile(outputPath, content, 0644); err != nil {
		slog.ErrorContext(ctx, "failed to write japanese.json", slog.Any("err", err))
		panic(err)
	}
}
