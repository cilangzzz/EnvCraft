package main

import "tsc/pkg/util/downloader"

func main() {
	httpDownloader := downloader.New(downloader.DownloadType.HTTP, downloader.DownloadOptions{})
	httpDownloader.Download(downloader.DownloadInfo{}, nil)
}
