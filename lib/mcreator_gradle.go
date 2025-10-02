package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// GradleFileInfo 存储Gradle文件信息
type GradleFileInfo struct {
	Version   string // Gradle版本号
	Edition   string // 版本类型 (bin/all)
	LockFile  string // .lck文件完整路径
	PartFile  string // .part文件完整路径
	TargetDir string // 目标目录
}

// 从文件名中提取Gradle版本信息
func extractGradleVersion(filename string) (version, edition string, err error) {
	// 正则表达式匹配 gradle-版本号-版本类型.zip 格式
	// 例如: gradle-8.14.2-bin.zip
	re := regexp.MustCompile(`gradle-(\d+\.\d+\.\d+)-(bin|all)\.zip`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 3 {
		err = fmt.Errorf("无法从文件名中提取Gradle版本信息: %s", filename)
		return
	}

	version = matches[1]
	edition = matches[2]
	return
}

// 扫描MCreator Gradle目录，查找.lck和.part文件
func ScanMCreatorGradleFiles(gradlePath string) ([]GradleFileInfo, error) {
	var results []GradleFileInfo

	// 检查目录是否存在
	if _, err := os.Stat(gradlePath); os.IsNotExist(err) {
		return results, fmt.Errorf("gradle目录不存在: %s", gradlePath)
	}

	// 遍历目录结构
	err := filepath.Walk(gradlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件
		if !info.IsDir() {
			filename := info.Name()

			// 检查是否为.lck或.part文件
			if strings.HasSuffix(filename, ".lck") || strings.HasSuffix(filename, ".part") {
				// 从文件名中提取版本信息（去掉.lck或.part后缀）
				baseName := strings.TrimSuffix(strings.TrimSuffix(filename, ".lck"), ".part")
				version, edition, err := extractGradleVersion(baseName)
				if err != nil {
					// 如果无法提取版本信息，跳过此文件
					return nil
				}

				// 查找是否已经存在相同版本的信息
				found := false
				for i, result := range results {
					if result.Version == version && result.Edition == edition {
						// 更新现有记录
						if strings.HasSuffix(filename, ".lck") {
							results[i].LockFile = path
						} else {
							results[i].PartFile = path
						}
						found = true
						break
					}
				}

				// 如果没找到，创建新记录
				if !found {
					info := GradleFileInfo{
						Version:   version,
						Edition:   edition,
						TargetDir: filepath.Dir(path),
					}

					if strings.HasSuffix(filename, ".lck") {
						info.LockFile = path
					} else {
						info.PartFile = path
					}

					results = append(results, info)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描Gradle目录失败: %v", err)
	}

	return results, nil
}

// 删除.lck和.part文件
func DeleteGradleTempFiles(info GradleFileInfo) error {
	if info.LockFile != "" {
		if err := os.Remove(info.LockFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除.lck文件失败: %v", err)
		}
		fmt.Printf("已删除: %s\n", filepath.Base(info.LockFile))
	}

	if info.PartFile != "" {
		if err := os.Remove(info.PartFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除.part文件失败: %v", err)
		}
		fmt.Printf("已删除: %s\n", filepath.Base(info.PartFile))
	}

	return nil
}

// 复制Gradle文件到目标目录
func CopyGradleToTarget(version, edition, targetDir string) error {
	// 下载Gradle
	if err := DownloadGradle(version, edition); err != nil {
		return fmt.Errorf("下载Gradle失败: %v", err)
	}

	// 源文件路径（缓存目录）
	cacheDir := filepath.Join(".", "cache")
	sourceFile := filepath.Join(cacheDir, version+"-"+edition+".zip")

	// 目标文件路径
	targetFile := filepath.Join(targetDir, fmt.Sprintf("gradle-%s-%s.zip", version, edition))

	// 获取源文件大小用于进度条
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %v", err)
	}
	fileSize := sourceInfo.Size()

	// 打开源文件
	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer source.Close()

	// 创建目标文件
	target, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer target.Close()

	// 创建复制进度条
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription("📋 复制进度"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(60),
		progressbar.OptionThrottle(50*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "✅ 复制完成\n")
		}),
		progressbar.OptionSpinnerType(9),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "🟢",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	// 使用带进度条的文件复制
	_, err = io.Copy(io.MultiWriter(target, bar), source)
	if err != nil {
		return fmt.Errorf("复制文件失败: %v", err)
	}

	fmt.Printf("✅ 已复制: %s -> %s\n",
		filepath.Base(sourceFile),
		filepath.Join(filepath.Base(targetDir), filepath.Base(targetFile)))

	return nil
}

// 处理MCreator Gradle下载问题
func ProcessMCreatorGradle(gradlePath string) error {
	fmt.Println("正在扫描MCreator Gradle目录...")

	// 扫描.lck和.part文件
	files, err := ScanMCreatorGradleFiles(gradlePath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("未找到需要处理的Gradle文件")
		return nil
	}

	fmt.Printf("找到 %d 个需要处理的Gradle版本:\n", len(files))

	// 处理每个Gradle版本
	for i, fileInfo := range files {
		fmt.Printf("\n[%d/%d] 处理Gradle %s %s版:\n",
			i+1, len(files), fileInfo.Version, fileInfo.Edition)

		// 1. 删除临时文件
		fmt.Println("1. 删除临时文件...")
		if err := DeleteGradleTempFiles(fileInfo); err != nil {
			return err
		}

		// 2. 下载并复制Gradle
		fmt.Println("2. 下载并复制Gradle...")
		if err := CopyGradleToTarget(fileInfo.Version, fileInfo.Edition, fileInfo.TargetDir); err != nil {
			return err
		}

		fmt.Printf("✅ Gradle %s %s版处理完成\n", fileInfo.Version, fileInfo.Edition)
	}

	fmt.Printf("\n✅ 所有Gradle版本处理完成！共处理了 %d 个版本\n", len(files))
	return nil
}
