package main

import (
	"fmt"
	"github.com/antmoveh/micro-version-management/pkg/models"
	"github.com/antmoveh/micro-version-management/pkg/repository"
	"github.com/urfave/cli"
	"log"
	"strings"
)

var searchCommand = cli.Command{
	Name:  "search",
	Usage: "镜像仓库中搜索镜像Tag列表: app search -t nexus -url http://username:password/xxx -name imageName",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t",
			Usage: "镜像仓库类型：nexus/harbor/dockerHub",
		},
		cli.StringFlag{
			Name:  "url",
			Usage: "镜像仓库地址",
		},
		cli.StringFlag{
			Name:     "name",
			Usage:    "镜像名称",
			Required: true,
		},
	},

	Action: func(context *cli.Context) error {

		searchRequest := &models.Search{
			Type: context.String("t"),
			Url:  context.String("url"),
			Name: context.String("name"),
		}
		if searchRequest.Type != "" && strings.ToLower(searchRequest.Type) != models.DockerHub && searchRequest.Url == "" {
			log.Fatal("镜像仓库类型不为dockerHub的必须指定镜像仓库地址，例：-url http://username:password@repository.service.cloud.com:8444")
		}
		SearchTags(searchRequest)
		return nil
	},
}

var releaseCommand = cli.Command{
	Name:  "release",
	Usage: "生成release版本yaml文件: app release -v v1.9 -t nexus -url http://username:password/xxx -f /tmp/template -o /tmp/release -apply",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "v",
			Usage: "指定过滤版本: v1.9",
		},
		cli.StringFlag{
			Name:  "t",
			Usage: "镜像仓库类型：nexus/harbor/dockerHub",
		},
		cli.StringFlag{
			Name:  "url",
			Usage: "镜像仓库地址",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "模板文件路径：默认值/tmp/template",
		},
		cli.StringFlag{
			Name:  "o",
			Usage: "生成模板路路径:默认值/tmp/release",
		},
		cli.BoolFlag{
			Name:  "apply",
			Usage: "是否自动执行: kubectl delete -f /tmp/release && kubectl apply -f /tmp/release",
		},
	},

	Action: func(context *cli.Context) error {

		return nil
	},
}

func SearchTags(searchRequest *models.Search) {
	if searchRequest.Type == "" || strings.ToLower(searchRequest.Type) == models.DockerHub {
		it, err := repository.DockerHubTags(searchRequest)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range it {
			fmt.Println(t.ImageName + ":" + t.ImageTag)
		}
		return
	}
	if strings.ToLower(searchRequest.Type) == models.Nexus {
		it, err := repository.NexusSearchTags(searchRequest)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range it {
			fmt.Println(t.ImageName + ":" + t.ImageTag)
		}
		return
	}
	if strings.ToLower(searchRequest.Type) == models.Harbor {
		it, err := repository.HarborTags(searchRequest)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range it {
			fmt.Println(t.ImageName + ":" + t.ImageTag)
		}
		return
	}
	log.Println("镜像仓库类型不正确，支持nexus/harbor/dockerhub")

}
