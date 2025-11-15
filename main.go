package main

import (
	"fmt"
	"log"
	"mcr_gradletools/lib"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// 版本信息常量
const (
	AppName    = "MCr_gradletools"
	Version    = "0.4.3"
	BuildDate  = "2025-10-03"
	GoVersion  = "go1.25"
	Repository = "https://gitee.com/CreateCN/mcrgradletool"
)

var currentUser, _ = user.Current()
var GradlePath = filepath.Join(currentUser.HomeDir, ".mcreator", "gradle", "wrapper", "dists")

func main() {
	app := &cli.App{
		Name:  "MCr_gradletools",
		Usage: "一款Go语言编写的MCreator Gradle工具",
		Commands: []*cli.Command{
			{
				Name:  "check-mirrors",
				Usage: "检查所有镜像源的可用性",
				Action: func(c *cli.Context) error {
					fmt.Println("正在检查镜像源可用性...")

					// 调用CheckAllMirrors函数
					availableMirrors, unavailableMirrors := lib.CheckAllMirrors()

					fmt.Printf("\n✅ 可用镜像源 (%d个):\n", len(availableMirrors))
					for _, mirror := range availableMirrors {
						fmt.Printf("  ✓ %s\n", mirror)
					}

					fmt.Printf("\n❌ 不可用镜像源 (%d个):\n", len(unavailableMirrors))
					for _, mirror := range unavailableMirrors {
						fmt.Printf("  ✗ %s\n", mirror)
					}

					fmt.Printf("\n总计: %d个镜像源，%d个可用，%d个不可用\n",
						len(availableMirrors)+len(unavailableMirrors),
						len(availableMirrors),
						len(unavailableMirrors))

					return nil
				},
			},
			{
				Name:  "clear-cache",
				Usage: "清理Gradle下载缓存",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "仅列出缓存文件，不删除",
					},
				},
				Action: func(c *cli.Context) error {
					listOnly := c.Bool("list")

					if listOnly {
						// 仅列出缓存文件
						files, err := lib.ListCacheFiles()
						if err != nil {
							return fmt.Errorf("获取缓存文件列表失败: %v", err)
						}

						if len(files) == 0 {
							fmt.Println("缓存目录为空")
							return nil
						}

						fmt.Printf("缓存目录 (%s) 中的文件:\n", lib.GetCacheDir())
						for i, file := range files {
							fmt.Printf("  %d. %s\n", i+1, file)
						}
						fmt.Printf("总计: %d 个文件\n", len(files))
						return nil
					}

					// 清理缓存
					fmt.Println("正在清理Gradle下载缓存...")

					// 先列出缓存文件
					files, err := lib.ListCacheFiles()
					if err != nil {
						return fmt.Errorf("获取缓存文件列表失败: %v", err)
					}

					if len(files) == 0 {
						fmt.Println("缓存目录为空，无需清理")
						return nil
					}

					fmt.Printf("即将删除 %d 个缓存文件:\n", len(files))
					for i, file := range files {
						fmt.Printf("  %d. %s\n", i+1, file)
					}

					// 确认删除
					fmt.Print("\n确认删除这些文件吗？(y/N): ")
					var confirm string
					fmt.Scanln(&confirm)

					if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
						fmt.Println("操作已取消")
						return nil
					}

					// 执行清理
					if err := lib.ClearCache(); err != nil {
						return fmt.Errorf("清理缓存失败: %v", err)
					}

					fmt.Println("✅ 缓存清理完成")
					return nil
				},
			},
			{
				Name:  "download",
				Usage: "下载指定版本的Gradle",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "version",
						Aliases:  []string{"v"},
						Usage:    "Gradle版本号 (例如: 8.7)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "edition",
						Aliases: []string{"e"},
						Usage:   "Gradle版本类型: bin (二进制版) 或 all (完整版)",
						Value:   "bin",
					},
				},
				Action: func(c *cli.Context) error {
					version := c.String("version")
					edition := c.String("edition")

					// 调用DownloadGradle函数
					err := lib.DownloadGradle(version, edition)
					if err != nil {
						return fmt.Errorf("下载Gradle失败: %v", err)
					}

					fmt.Printf("Gradle %s %s版下载安装成功！\n", version, edition)
					return nil
				},
			},
			{
				Name:  "gradle",
				Usage: "自动处理MCreator的Gradle下载问题",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "MCreator Gradle目录路径",
						Value:   GradlePath,
					},
				},
				Action: func(c *cli.Context) error {
					gradlePath := c.String("path")

					// 调用ProcessMCreatorGradle函数
					err := lib.ProcessMCreatorGradle(gradlePath)
					if err != nil {
						return fmt.Errorf("处理MCreator Gradle失败: %v", err)
					}

					return nil
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v", "ver"},
				Usage:   "显示程序版本信息",
				Action: func(c *cli.Context) error {
					fmt.Printf("%s 版本信息\n", AppName)
					fmt.Printf("版本: %s\n", Version)
					fmt.Printf("构建日期: %s\n", BuildDate)
					fmt.Printf("Go版本: %s\n", GoVersion)
					fmt.Printf("项目仓库: %s\n", Repository)
					fmt.Println("\n一款专为MCreator设计的Gradle管理工具")
					fmt.Println("作者: CreateCN")
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("MCr_gradletools - MCreator Gradle管理工具")
			fmt.Println("使用 '--help' 查看可用命令")
			fmt.Println("可用命令:")
			fmt.Println("  check-mirrors - 检查镜像源可用性")
			fmt.Println("  clear-cache   - 清理Gradle下载缓存")
			fmt.Println("  download      - 下载指定版本的Gradle")
			fmt.Println("  gradle        - 自动处理MCreator的Gradle下载问题")
			fmt.Println("  version       - 显示程序版本信息")
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
