package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"gopkg.in/yaml.v3"
)

type flags struct {
	concurrency   int
	files         int
	iterations    int
	fileTimeout   time.Duration
	globalTimeout time.Duration
	outputDir     string
}

func main() {
	fl := parseFlags()

	if _, err := os.Stat(fl.outputDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Error("не удалось получить информацию о выходной директории", slog.Any("ошибка", err))
			return
		}

		if err = os.MkdirAll(fl.outputDir, 0755); err != nil {
			slog.Error("не удалось создать выходные директории", slog.Any("ошибка", err))
			return
		}
	}

	wp := newWorkerPool(fl.concurrency)
	defer wp.stop()

	globalCtx := context.Background()

	if fl.globalTimeout > 0 {
		var cancel context.CancelFunc
		globalCtx, cancel = context.WithTimeout(globalCtx, fl.globalTimeout)
		defer cancel()
	}

	for i := range fl.files {
		wp.submit(func() {
			fileCtx := globalCtx

			if fl.fileTimeout > 0 {
				var cancel context.CancelFunc
				fileCtx, cancel = context.WithTimeout(globalCtx, fl.fileTimeout)
				defer cancel()
			}

			filePath := path.Join(fl.outputDir, fmt.Sprintf("file-%d.yaml", i+1))

			for range fl.iterations {
				writeFileRandomPayload(fileCtx, filePath)
			}
		})
	}

	slog.Info("Файлы успешно записаны")
}

func writeFileRandomPayload(ctx context.Context, filePath string) {
	select {
	case <-ctx.Done():
		slog.Warn("таймаут записи файла", slog.String("файл", filePath))
		return
	default:
	}

	flePayload := generateRandomPayload()

	data, err := yaml.Marshal(flePayload)
	if err != nil {
		slog.Error(
			"не удалось сериализовать данные файла",
			slog.Any("ошибка", err),
			slog.String("файл", filePath),
		)
		return
	}

	if err = os.WriteFile(filePath, data, 0644); err != nil {
		slog.Error(
			"не удалось записать данные в файл",
			slog.Any("ошибка", err),
			slog.String("файл", filePath),
		)
	}
}

func parseFlags() flags {
	var fl flags

	flag.IntVar(&fl.concurrency, "concurrency", 10, "Количество потоков для конкурентной записи файлов")
	flag.IntVar(&fl.files, "files", 10, "Количество файлов для записи")
	flag.IntVar(&fl.iterations, "iterations", 1, "Количество итераций перезаписи каждого файла")
	flag.DurationVar(&fl.fileTimeout, "file-timeout", -1, "Таймаут на перезапись отдельного файла")
	flag.DurationVar(&fl.globalTimeout, "global-timeout", -1, "Таймаут на перезапись всех файлов")
	flag.StringVar(&fl.outputDir, "output-dir", ".output", "Директория, в которой где должны храниться выходные файлы")
	flag.Parse()

	return fl
}
