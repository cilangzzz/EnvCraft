package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
	"tsc/pkg/util/downloader/util"
)

type httpDownloader struct {
	options DownloadOptions
}

func NewHTTPDownloader(options DownloadOptions) Downloader {
	return &httpDownloader{
		options: options,
	}
}

func (h *httpDownloader) Download(info DownloadInfo, writer io.Writer) error {
	// 设置默认值
	if info.Timeout == 0 {
		info.Timeout = h.options.DefaultTimeout
	}
	if info.MaxRetries == 0 {
		info.MaxRetries = h.options.DefaultMaxRetries
	}
	if info.UserAgent == "" {
		info.UserAgent = h.options.DefaultUserAgent
	}

	client := h.options.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: info.Timeout,
		}
	}

	// 处理代理
	if info.ProxyURL != "" {
		proxyURL, err := url.Parse(info.ProxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	var lastError error
	for i := 0; i < info.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(info.RetryDelay)
		}

		err := h.doDownload(client, info, writer)
		if err == nil {
			return nil
		}
		lastError = err
	}

	return fmt.Errorf("after %d retries, last error: %w", info.MaxRetries, lastError)
}

func (h *httpDownloader) doDownload(client *http.Client, info DownloadInfo, writer io.Writer) error {
	req, err := http.NewRequest("GET", info.URL, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range info.Headers {
		req.Header.Set(k, v)
	}
	if info.UserAgent != "" {
		req.Header.Set("User-Agent", info.UserAgent)
	}

	// 支持断点续传
	if info.ResumeDownload {
		if fi, err := os.Stat(info.Dest); err == nil {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fi.Size()))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 处理目标路径
	dest := info.Dest
	if fi, err := os.Stat(dest); err == nil && fi.IsDir() {
		filename := filepath.Base(info.URL)
		if filename == "" {
			filename = "downloaded_file"
		}
		dest = filepath.Join(dest, filename)
	}

	// 创建目标文件
	flags := os.O_CREATE | os.O_WRONLY
	if info.ResumeDownload {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(dest, flags, info.FileMode)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	// 设置进度写入器
	var destWriter io.Writer = file
	if writer != nil {
		if pw, ok := writer.(ProgressWriter); ok {
			pw.SetTotal(resp.ContentLength)
		}
		destWriter = io.MultiWriter(file, writer)
	}

	// 执行下载
	if _, err := io.Copy(destWriter, resp.Body); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	// 校验文件
	if info.Checksum != "" {
		if err := util.VerifyChecksum(dest, info.Checksum, info.ChecksumType); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	return nil
}

func (h *httpDownloader) SetDefaultOptions(options DownloadOptions) {
	h.options = options
}
