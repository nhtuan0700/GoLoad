package logic

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
)

const (
	HTTPResponseHeaderContentType = "Content-Type"
	HTTPMetadataKeyContentType    = "content-type"
)

type Downloader interface {
	Download(ctx context.Context, writer io.Writer) (map[string]any, error)
}

type downloader struct {
	url    string
	logger *zap.Logger
}

func NewDownloader(
	url string,
	logger *zap.Logger,
) Downloader {
	return &downloader{
		url:    url,
		logger: logger,
	}
}

func (d downloader) Download(ctx context.Context, writer io.Writer) (map[string]any, error) {
	logger := utils.LoggerWithContext(ctx, d.logger)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, http.NoBody)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create http get request")
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to make http get request")
		return nil, err
	}
	defer response.Body.Close()

	copied, err := io.Copy(writer, response.Body)
	log.Println("copied len: ", copied)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to read response and write to writer")
		return nil, err
	}

	metadata := map[string]any{
		HTTPMetadataKeyContentType: response.Header.Get(HTTPResponseHeaderContentType),
	}

	return metadata, nil
}
