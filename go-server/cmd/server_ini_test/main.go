package main

import (
	"log"
	core2 "tsc/pkg/util/server_ini/core"
)

func main() {
	// 创建HTTP下载器 下载服务
	//httpDownloader := downloader.New(downloader.DownloadType.HTTP, core.DownloadOptions{})
	//url := "https://github.com/git-for-windows/git/archive/refs/heads/main.zip"
	//downInfo := core.NewDownloadInfo(url, "./testdata/")
	//err := httpDownloader.Download(downInfo, nil)
	//if err != nil {
	//	panic(err)
	//	return
	//}

	// 安装服务
	filePath := "./testdata/main.zip"
	pkg, err := core2.OpenPackage(filePath)
	if err != nil {
		log.Fatalf("OpenPackage failed: %v", err)
	}
	for i, file := range pkg.FileInfos {
		log.Println(i, file.FileInfo.Name())
	}

	// 提取服务
	testOutputDir := "testdata/output"
	options := core2.ExtractOptions{
		TargetDir:    testOutputDir,
		Overwrite:    true,
		PreservePerm: true,
	}
	err = pkg.Extract(options)
	if err != nil {
		log.Fatalf("Extract failed: %v", err)
		return
	}

}
