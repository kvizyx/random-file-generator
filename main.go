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

func parseFlags() flags {
	var fl flags

	flag.IntVar(&fl.concurrency, "concurrency", 10, "Количество потоков для конкурентной записи файлов")
	flag.IntVar(&fl.files, "files", 10, "Количество файлов для записи")
	flag.IntVar(&fl.iterations, "iterations", 1, "Количество итераций перезаписи каждого файла")
	flag.DurationVar(&fl.fileTimeout, "file-timeout", -1, "Таймаут на перезапись отдельного файла")
	flag.DurationVar(&fl.globalTimeout, "global-timeout", -1, "Таймаут на перезапись всех файлов")
	flag.StringVar(&fl.outputDir, "output-dir", ".output", "Директория, в которой должны храниться выходные файлы")
	flag.Parse()

	return fl
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

			var (
				filePath   = path.Join(fl.outputDir, fmt.Sprintf("file-%d.yaml", i+1))
				fileExists = true
			)

			filePayload, err := os.ReadFile(filePath)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					slog.Error("не удалось получить информацию о файле", slog.Any("ошибка", err))
					return
				}

				fileExists = false
			}

			for range fl.iterations {
				select {
				case <-fileCtx.Done():
					slog.Warn("таймаут записи файла", slog.String("файл", filePath))
					return
				default:
				}

				if fileExists {
					if err = updateFileRandomPayload(filePath, filePayload); err != nil {
						slog.Error(
							"не удалось обновить данные файла",
							slog.Any("ошибка", err),
							slog.String("файл", filePath),
						)
					}
					continue
				}

				if err = createFileRandomPayload(filePath); err != nil {
					slog.Error(
						"не удалось создать файл с данными",
						slog.Any("ошибка", err),
						slog.String("файл", filePath),
					)
				}
			}
		})
	}

	slog.Info("Файлы успешно записаны")
}

func updateFileRandomPayload(filePath string, oldFilePayloadBytes []byte) error {
	var oldFilePayload payload

	if err := yaml.Unmarshal(oldFilePayloadBytes, &oldFilePayload); err != nil {
		return fmt.Errorf("не удалось десериализовать данные файла: %w", err)
	}

	freshFilePayload := generateRandomPayload()
	freshFilePayload.Id = oldFilePayload.Id
	freshFilePayload.CreatedAt = oldFilePayload.CreatedAt

	freshFilePayloadBytes, err := yaml.Marshal(freshFilePayload)
	if err != nil {
		return fmt.Errorf("не удалось сериализовать данные файла: %w", err)
	}

	if err = os.WriteFile(filePath, freshFilePayloadBytes, 0644); err != nil {
		return fmt.Errorf("не удалось записать данные в файл: %w", err)
	}

	return nil
}

func createFileRandomPayload(filePath string) error {
	filePayload := generateRandomPayload()

	data, err := yaml.Marshal(filePayload)
	if err != nil {
		return fmt.Errorf("не удалось сериализовать данные файла: %w", err)
	}

	if err = os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать данные в файл: %w", err)
	}

	return nil
}
