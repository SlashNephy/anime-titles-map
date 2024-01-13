package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/goccy/go-json"

	"github.com/SlashNephy/anime-titles-map/external"
	_ "github.com/SlashNephy/anime-titles-map/logger"
	"github.com/SlashNephy/anime-titles-map/processor"
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

	var (
		anilistTitles []*processor.MediaTitle
		malTitles     []*processor.MediaTitle
		wg            sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()

		titles, err := anilist.FetchTitles(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch AniList titles", slog.Any("err", err))
			panic(err)
		}
		anilistTitles = append(anilistTitles, titles...)
	}()
	go func() {
		defer wg.Done()

		titles, err := mal.FetchTitles(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch MAL titles", slog.Any("err", err))
			panic(err)
		}
		malTitles = append(malTitles, titles...)
	}()
	wg.Wait()

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
