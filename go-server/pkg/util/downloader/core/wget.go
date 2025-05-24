package core

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"tsc/pkg/util/downloader"
)

type wgetDownloader struct {
	httpDownloader downloader.Downloader
	options        downloader.DownloadOptions
}

func NewWgetDownloader(options downloader.DownloadOptions) downloader.Downloader {
	return &wgetDownloader{
		httpDownloader: NewHTTPDownloader(options),
		options:        options,
	}
}

func (w *wgetDownloader) Download(info downloader.DownloadInfo, writer io.Writer) error {
	// 自动从URL提取文件名
	if fi, err := os.Stat(info.Dest); err == nil && fi.IsDir() {
		parsedURL, err := url.Parse(info.URL)
		if err != nil {
			return fmt.Errorf("parse URL failed: %w", err)
		}

		filename := filepath.Base(parsedURL.Path)
		if filename == "" {
			filename = "index.html"
		}
		info.Dest = filepath.Join(info.Dest, filename)
	}

	return w.httpDownloader.Download(info, writer)
}

func (w *wgetDownloader) SetDefaultOptions(options downloader.DownloadOptions) {
	w.options = options
	w.httpDownloader.SetDefaultOptions(options)
}
