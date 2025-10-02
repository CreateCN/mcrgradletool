package main

import (
	"fmt"
	"log"
	"mcr_gradletools/lib"
	"os"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var currentUser, _ = user.Current()
var GradlePath = filepath.Join(currentUser.HomeDir, ".mcreator", "gradle", "wrapper", "dists")

func main() {
	app := &cli.App{
		Name:  "MCr_gradletools",
		Usage: "MCr_gradletools is a tool to help you manage your MCr mod",
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
		},
		Action: func(c *cli.Context) error {
			fmt.Println("MCr_gradletools - MCreator Gradle管理工具")
			fmt.Println("使用 '--help' 查看可用命令")
			fmt.Println("可用命令:")
			fmt.Println("  check-mirrors - 检查镜像源可用性")
			fmt.Println("  download     - 下载指定版本的Gradle")
			fmt.Println("  gradle       - 自动处理MCreator的Gradle下载问题")
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
