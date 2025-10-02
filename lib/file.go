package lib

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// FileInfo 存储文件信息的结构体
type FileInfo struct {
	FilePath string // 文件完整路径
	FileName string // 文件名
}

// FileFind 从指定文件夹搜寻符合正则模式的文件并返回文件信息列表
func FileFind(root string, pattern string) ([]FileInfo, error) {
	var result []FileInfo

	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// 遍历目录
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是文件且文件名匹配正则表达式
		if !info.IsDir() && regex.MatchString(info.Name()) {
			// 添加到结果列表
			result = append(result, FileInfo{
				FilePath: path,
				FileName: info.Name(),
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func FileCopy(filePath string, newPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return err
	}
	//复制文件
	srcFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	//删除文件
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}
