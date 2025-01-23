package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	"golang.org/x/sync/errgroup"

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
	)
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		titles, err := anilist.FetchTitles(egCtx)
		if err != nil {
			return err
		}

		anilistTitles = append(anilistTitles, titles...)
		return nil
	})
	eg.Go(func() error {
		titles, err := mal.FetchTitles(egCtx)
		if err != nil {
			return err
		}

		malTitles = append(malTitles, titles...)
		return nil
	})
	if err = eg.Wait(); err != nil {
		if errors.Is(err, external.ErrServerError) {
			slog.WarnContext(ctx, "failed to fetch due to server error. Exiting...")
			return
		}

		slog.ErrorContext(ctx, "failed to fetch titles", slog.Any("err", err))
		panic(err)
	}

	titles, err := processor.MergeTitles(append(anilistTitles, malTitles...)...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to merge titles", slog.Any("err", err))
		panic(err)
	}

	content, err := json.MarshalIndent(titles, "", "  ")
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
