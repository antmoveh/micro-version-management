package main

import (
	"errors"
	"fmt"
	"github.com/antmoveh/micro-version-management/pkg/models"
	"github.com/antmoveh/micro-version-management/pkg/repository"
	"github.com/urfave/cli"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var searchCommand = cli.Command{
	Name:  "search",
	Usage: "镜像仓库中搜索镜像Tag列表: app search -t nexus -url http://username:password/xxx -name imageName -v v1.3",
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
		cli.StringFlag{
			Name:  "v",
			Usage: "指定过滤版本: v1.9",
		},
	},

	Action: func(context *cli.Context) error {

		searchRequest := &models.Search{
			Type:    context.String("t"),
			Url:     context.String("url"),
			Name:    context.String("name"),
			Version: context.String("v"),
		}
		if searchRequest.Type != "" && strings.ToLower(searchRequest.Type) != models.DockerHub && searchRequest.Url == "" {
			log.Fatal("镜像仓库类型不为dockerHub的必须指定镜像仓库地址，例：-url http://username:password@repository.service.cloud.com:8444")
		}
		printRemoteImage(searchRequest)
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
		cli.StringFlag{
			Name:  "prefix",
			Usage: "镜像名称前缀,默认moebius/release/",
		},
		cli.StringFlag{
			Name:  "domain",
			Usage: "指定生成镜像的域名，默认为空",
		},
	},

	Action: func(context *cli.Context) error {

		autoApply := false
		if context.Args().Get(0) == "apply" {
			autoApply = true
		}

		releaseRequest := &models.Release{
			Type:         context.String("t"),
			Url:          context.String("url"),
			TemplatePath: context.String("f"),
			ReleasePath:  context.String("o"),
			Apply:        autoApply,
			Version:      context.String("v"),
			Prefix:       context.String("prefix"),
			Domain:       context.String("domain"),
		}
		releaseYaml(releaseRequest)
		return nil
	},
}

func printRemoteImage(searchRequest *models.Search) {
	it, err := SearchImage(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range it {
		if searchRequest.Version == "" {
			fmt.Println(t.ImageName + ":" + t.ImageTag)
		} else if strings.Contains(t.ImageTag, searchRequest.Version) {
			fmt.Println(t.ImageName + ":" + t.ImageTag)
		}
	}

}

func releaseYaml(releaseRequest *models.Release) {
	if releaseRequest.TemplatePath == "" {
		releaseRequest.TemplatePath = "/tmp/template"
	}
	if _, err := os.Stat(releaseRequest.TemplatePath); err != nil {
		log.Fatal("模板文件目录不存在：/tmp/template")
	}
	if releaseRequest.ReleasePath == "" {
		releaseRequest.ReleasePath = "/tmp/release"
	}

	_ = os.RemoveAll(releaseRequest.ReleasePath)

	err := os.MkdirAll(releaseRequest.ReleasePath, os.ModePerm)
	if err != nil {
		log.Fatal("创建release目录失败：" + err.Error())
	}

	if releaseRequest.Prefix == "" {
		releaseRequest.Prefix = "moebius/release/"
	}

	imageNameList := []string{}
	imagePathMap := map[string]string{}
	err = filepath.Walk(releaseRequest.TemplatePath, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yaml") {
			path1 := strings.Split(path[:len(path)-5], "/")
			imageName := fmt.Sprintf("%s%s", releaseRequest.Prefix, path1[len(path1)-1])
			imageNameList = append(imageNameList, imageName)
			imagePathMap[imageName] = path
		}

	})
	if err != nil {
		log.Fatal("获取镜像名称失败")
	}

	for _, name := range imageNameList {
		searchRequest := &models.Search{
			Type:    releaseRequest.Type,
			Url:     releaseRequest.Url,
			Name:    name,
			Version: releaseRequest.Version,
		}
		it, err := SearchImage(searchRequest)
		if err != nil {
			log.Fatal("镜像查询失败")
		}
		latest := QueryReleaseLatestVersion(it, releaseRequest.Version)
	}

}

func SearchImage(searchRequest *models.Search) ([]*models.ImageTags, error) {
	if searchRequest.Type == "" || strings.ToLower(searchRequest.Type) == models.DockerHub {
		it, err := repository.DockerHubTags(searchRequest)
		if err != nil {
			return nil, err
		}
		return it, nil
	}
	if strings.ToLower(searchRequest.Type) == models.Nexus {
		it, err := repository.NexusSearchTags(searchRequest)
		if err != nil {
			return nil, err
		}
		return it, nil
	}
	if strings.ToLower(searchRequest.Type) == models.Harbor {
		it, err := repository.HarborTags(searchRequest)
		if err != nil {
			return nil, err
		}
		return it, nil
	}
	return nil, errors.New("镜像仓库类型不正确，支持nexus/harbor/dockerhub")
}

// 计算最新版本
func QueryReleaseLatestVersion(it []*models.ImageTags, version string) string {
	if len(it) == 0 {
		return ""
	}

	//latestImage := ""
	prefixVersion := []int{}
	suffixVersion := 0
	for _, t := range it {
		// 只处理符合v1.8.2-10 格式的数据
		verify, _ := regexp.Match("^v(\\d+\\.?){2,3}-\\d+", []byte(t.ImageTag))
		if !verify {
			continue
		}
		if version != "" && strings.HasPrefix(t.ImageTag, version) {
			n, err := strconv.Atoi(t.ImageTag[len(version):])
			if err != nil && n > suffixVersion {
				suffixVersion = n
			}
			continue
		}
		if version == "" {
			tmpPrefixVersion := []int{}
			x := strings.Split(t.ImageTag[1:], "-")
			x1 := strings.Split(x[0], ".")

			n, err := strconv.Atoi(x1[1])
			if err != nil {
				continue
			}
			tmpSuffixVersion := n

			for _, x2 := range x1 {
				n, err := strconv.Atoi(x2)
				if err != nil {
					tmpPrefixVersion = []int{}
					break
				}
				tmpPrefixVersion = append(tmpPrefixVersion, n)
			}
			if len(tmpPrefixVersion) == 0 {
				continue
			}
			// 大版本长度有为两位，有为三位的，先把两位的最后以为补充为-1方便比较
			if len(tmpPrefixVersion) == 2 {
				tmpPrefixVersion = append(tmpPrefixVersion, -1)
			}
			// 与现在的最新版本比较，替换最新版本
			if len(prefixVersion) == 0 {
				prefixVersion = tmpPrefixVersion
				suffixVersion = tmpSuffixVersion
				continue
			}
			if tmpPrefixVersion[0] > prefixVersion[0] {
				prefixVersion = tmpPrefixVersion
				suffixVersion = tmpSuffixVersion
				continue
			}

		}
	}
	return ""
}
