package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/sunmi-OS/gocore/v2/utils"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/urfave/cli/v2"
)

func main() {
	// 打印banner
	utils.PrintBanner("oss")
	// 配置cli参数
	app := cli.NewApp()
	app.Name = "oss"
	app.Usage = "oss"
	app.Version = "v1.0.0"
	// 指定命令运行的函数
	app.Commands = []*cli.Command{
		{
			Name: "push",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "common/test.txt",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "common/test.txt",
				},
			},
			Usage:  "oss push --name [common/test.txt] --path [common/test.txt]",
			Action: push,
		},
		{
			Name: "pull",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "common/test.txt",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "common/test.txt",
				},
			},
			Usage:  "oss pull --name [common/test.txt] --path [common/test.txt]",
			Action: pull,
		},
		{
			Name: "delete",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "common/test.txt",
				},
			},
			Usage:  "oss delete --name [common/test.txt]",
			Action: delete,
		},
		{
			Name: "list",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "prefix",
					Usage: "conf",
				},
			},
			Usage:  "oss list --prefix [conf]",
			Action: list,
		},
		{
			Name:   "scale",
			Usage:  "oss scale",
			Action: scale,
		},
	}
	// 启动cli
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
func NewClient() *cos.Client {
	u, _ := url.Parse(os.Getenv("OSS_HOST"))
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  os.Getenv("OSS_SECRET_ID"),  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
			SecretKey: os.Getenv("OSS_SECRET_KEY"), // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
		},
	})
	return client
}

func push(c *cli.Context) error {
	client := NewClient()
	name := c.String("name")
	path := c.String("path")
	_, err := client.Object.PutFromFile(context.Background(), name, path, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("oss.xiaoyuantongbbs.cn/%s\n", name)
	return nil
}

func pull(c *cli.Context) error {
	client := NewClient()
	name := c.String("name")
	path := c.String("path")
	_, err := client.Object.GetToFile(context.Background(), name, path, nil)
	if err != nil {
		panic(err)
	}
	return nil
}

func delete(c *cli.Context) error {
	client := NewClient()
	name := c.String("name")
	_, err := client.Object.Delete(context.Background(), name)
	if err != nil {
		panic(err)
	}
	return nil
}

func list(c *cli.Context) error {
	client := NewClient()
	prefix := c.String("prefix")
	opt := &cos.BucketGetOptions{
		Prefix:  prefix,
		MaxKeys: 10,
	}
	v, _, err := client.Bucket.Get(context.Background(), opt)
	if err != nil {
		panic(err)
	}

	for _, c := range v.Contents {
		fmt.Printf("%s, %d KB\n", c.Key, c.Size/1024)
	}
	return nil
}

func scale(c *cli.Context) error {
	a1 := []string{
		"!<Scale>p",
		"!<Scale>px",
		"!x<Scale>p",
		"<Width>x",
		"x<Height>",
		"<Width>x<Height>!",
	}
	a2 := []string{
		"指定图片的宽高为原图的 Scale%",
		"指定图片的宽为原图的 Scale%，高度不变",
		"指定图片的宽为原图的 Scale%，高度不变",
		"指定目标图片宽度为 Width，高度等比缩放",
		"指定目标图片高度为 Height，宽度等比缩放",
		"忽略原图宽高比例，指定图片宽度为 Width，高度为 Height，强行缩放图片，可能导致目标图片变形",
	}
	for k1 := range a1 {
		fmt.Printf("?imageMogr2/thumbnail/%s :%s kb\n", a1[k1], a2[k1])
	}
	return nil
}
