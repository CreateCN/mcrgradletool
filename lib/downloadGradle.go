package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// 镜像源配置
var mirrors = []struct {
	name string
	url  string
}{
	{"腾讯镜像-bin", "https://mirrors.cloud.tencent.com/gradle/gradle-{{version}}-bin.zip"},
	{"腾讯镜像-all", "https://mirrors.cloud.tencent.com/gradle/gradle-{{version}}-all.zip"},
	{"华为云镜像-bin", "https://mirrors.huaweicloud.com/gradle/gradle-{{version}}-bin.zip"},
	{"华为云镜像-all", "https://mirrors.huaweicloud.com/gradle/gradle-{{version}}-all.zip"},
	{"清华镜像-bin", "https://mirrors.tuna.tsinghua.edu.cn/gradle/gradle-{{version}}-bin.zip"},
	{"清华镜像-all", "https://mirrors.tuna.tsinghua.edu.cn/gradle/gradle-{{version}}-all.zip"},
}

// 检查镜像是否可用
func checkMirrorAvailability(url string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 使用简单的GET请求检查镜像可用性
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	// 添加常见的HTTP请求头，模拟浏览器行为
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 接受200 OK状态码
	return resp.StatusCode == http.StatusOK
}

// 下载文件
func downloadFile(url, filepath string) error {
	client := &http.Client{
		Timeout: 300 * time.Second, // 下载需要更长的超时时间（5分钟）
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %v", err)
	}

	// 添加常见的HTTP请求头，模拟浏览器行为
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 获取文件大小用于进度条
	contentLength := resp.ContentLength
	
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	// 创建进度条
	bar := progressbar.NewOptions64(
		contentLength,
		progressbar.OptionSetDescription("下载进度"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// 使用带进度条的下载
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}

	return nil
}

// 检查所有镜像源的可用性
func CheckAllMirrors() ([]string, []string) {
	var availableMirrors []string
	var unavailableMirrors []string
	
	// 使用一个测试版本号来检查镜像可用性
	testVersion := "8.7"
	
	for _, mirror := range mirrors {
		// 替换版本号占位符
		url := strings.Replace(mirror.url, "{{version}}", testVersion, -1)
		
		if checkMirrorAvailability(url) {
			availableMirrors = append(availableMirrors, mirror.name)
		} else {
			unavailableMirrors = append(unavailableMirrors, mirror.name)
		}
	}
	
	return availableMirrors, unavailableMirrors
}

// 获取缓存目录路径
func GetCacheDir() string {
	return filepath.Join(".", "cache")
}

// 删除缓存目录中的所有文件
func ClearCache() error {
	cacheDir := GetCacheDir()
	
	// 检查缓存目录是否存在
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return fmt.Errorf("缓存目录不存在: %s", cacheDir)
	}
	
	// 遍历缓存目录并删除所有文件
	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 跳过目录本身，只删除文件
		if !info.IsDir() {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("删除文件失败: %s, 错误: %v", path, err)
			}
			fmt.Printf("已删除: %s\n", filepath.Base(path))
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("清理缓存失败: %v", err)
	}
	
	return nil
}

// 获取缓存目录中的文件列表
func ListCacheFiles() ([]string, error) {
	cacheDir := GetCacheDir()
	var files []string
	
	// 检查缓存目录是否存在
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return files, nil // 目录不存在，返回空列表
	}
	
	// 读取缓存目录
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("读取缓存目录失败: %v", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// 下载并安装Gradle
// edition参数指定下载版本："bin" 或 "all"
func DownloadGradle(version, edition string) error {

	// 创建缓存目录
	cacheDir := filepath.Join(".", "cache")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建缓存目录失败: %v", err)
	}

	// 检查是否已存在（检查ZIP文件，区分edition）
	gradleZipFile := filepath.Join(cacheDir, version+"-"+edition+".zip")
	if _, err := os.Stat(gradleZipFile); err == nil {
		fmt.Printf("Gradle %s %s版 已存在于缓存目录中\n", version, edition)
		return nil
	}

	// 检查edition参数有效性
	if edition != "bin" && edition != "all" {
		return fmt.Errorf("edition参数必须为 'bin' 或 'all'，当前为: %s", edition)
	}

	// 尝试不同的镜像源（根据edition过滤）
	var availableMirror string
	for _, mirror := range mirrors {
		// 只检查与指定edition匹配的镜像源
		if strings.HasSuffix(mirror.name, "-"+edition) {
			url := strings.Replace(mirror.url, "{{version}}", version, -1)
			fmt.Printf("正在检查 %s 可用性...\n", mirror.name)

			if checkMirrorAvailability(url) {
				fmt.Printf("%s 可用\n", mirror.name)
				availableMirror = url
				break
			}
			fmt.Printf("%s 不可用\n", mirror.name)
		}
	}

	if availableMirror == "" {
		return fmt.Errorf("所有%s版镜像源都不可用", edition)
	}

	// 下载文件
	tempFile := filepath.Join(cacheDir, version+"-"+edition+".zip")
	fmt.Printf("正在从镜像下载 %s %s版...\n", version, edition)

	if err := downloadFile(availableMirror, tempFile); err != nil {
		return err
	}

	fmt.Printf("Gradle %s %s版 下载完成\n", version, edition)
	return nil
}
