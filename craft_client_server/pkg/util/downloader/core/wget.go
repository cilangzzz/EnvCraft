package core

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

type wgetDownloader struct {
	httpDownloader Downloader
	options        DownloadOptions
}

func NewWgetDownloader(options DownloadOptions) Downloader {
	return &wgetDownloader{
		httpDownloader: NewHTTPDownloader(options),
		options:        options,
	}
}

func (w *wgetDownloader) Download(info DownloadInfo, writer io.Writer) error {
	// 自动从URL提取文件名
	if fi, err := os.Stat(info.Dest); err == nil && fi.IsDir() {
		parsedURL, err := url.Parse(info.URL)
		if err != nil {
			return fmt.Errorf("parse URL failed: %w", err)
		}

		// 处理目标路径
		dest := info.Dest
		if fi, err := os.Stat(dest); err == nil && fi.IsDir() {
		} else if os.IsNotExist(err) {
			// 如果目标路径不存在，创建所有父目录
			dir := filepath.Dir(dest)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
		filename := filepath.Base(parsedURL.Path)
		if filename == "" {
			filename = "index.html"
		}
		info.Dest = filepath.Join(info.Dest, filename)
	}

	return w.httpDownloader.Download(info, writer)
}

func (w *wgetDownloader) SetDefaultOptions(options DownloadOptions) {
	w.options = options
	w.httpDownloader.SetDefaultOptions(options)
}
