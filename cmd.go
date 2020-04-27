package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/antmoveh/micro-version-management/pkg/models"
	"github.com/antmoveh/micro-version-management/pkg/repository"
	"github.com/antmoveh/micro-version-management/pkg/utils"
	"github.com/urfave/cli"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var searchCommand = cli.Command{
	Name:  "search",
	Usage: "镜像仓库中搜索镜像Tag列表: app search -t nexus -url http://username:password/xxx -name imageName -v v1.3 -f image_name_list.conf",
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
			Required: false,
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "指定过滤版本: v1.9",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "-f执行文件/搜索镜像最新版本",
		},
	},

	Action: func(context *cli.Context) error {

		searchRequest := &models.Search{
			Type:    context.String("t"),
			Url:     context.String("url"),
			Name:    context.String("name"),
			Version: context.String("v"),
			File:    context.String("f"),
		}
		if searchRequest.Type != "" && strings.ToLower(searchRequest.Type) != models.DockerHub && searchRequest.Url == "" {
			log.Fatal("镜像仓库类型不为dockerHub的必须指定镜像仓库地址，例：-url http://username:password@repository.service.cloud.com:8444")
		}
		if searchRequest.Name == "" && searchRequest.File == "" {
			log.Fatal("镜像名称或镜像名称列表文件必需存在一个，若指定文件应用程序会逐行读取镜像名称然后获取该镜像最新版本")
		}
		if searchRequest.File != "" {
			printLatestImage(searchRequest)
		} else {
			printRemoteImage(searchRequest)
		}
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
			Usage: "镜像仓库门户页地址: http://username:password@repository.xxx.com:8081",
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

func printLatestImage(searchRequest *models.Search) {
	f, err := os.Open(searchRequest.File)
	if err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewReader(f)
	for {
		b, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		searchRequest.Name = strings.Replace(string(b), "\r\n", "", -1)
		it, err := SearchImage(searchRequest)
		if err != nil {
			log.Fatal(err)
		}
		latestVersion := QueryReleaseLatestVersion(it, searchRequest.Version)
		if latestVersion == "" {
			fmt.Println(searchRequest.Name + "： 未查询到最新版本")
			continue
		}
		fmt.Println(searchRequest.Name + ":" + latestVersion)
	}
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
	if releaseRequest.Prefix != "" && !strings.HasSuffix(releaseRequest.Prefix, "/") {
		releaseRequest.Prefix = releaseRequest.Prefix + "/"
	}
	if releaseRequest.Domain != "" && !strings.HasSuffix(releaseRequest.Domain, "/") {
		releaseRequest.Domain = releaseRequest.Domain + "/"
	}

	_ = os.RemoveAll(releaseRequest.ReleasePath)
	time.Sleep(1 * time.Second)
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
		log.Println("模板yaml：" + path)
		if strings.HasSuffix(path, ".yaml") {
			path1 := strings.Replace(path, "\\", "/", -1)
			path2 := strings.Split(path1[:len(path1)-5], "/")
			imageName := fmt.Sprintf("%s%s", releaseRequest.Prefix, path2[len(path2)-1])
			imageNameList = append(imageNameList, imageName)
			imagePathMap[imageName] = path
		}
		return nil

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
		log.Println("搜索该镜像所有tag：" + searchRequest.Name)
		it, err := SearchImage(searchRequest)
		if err != nil {
			log.Fatal("镜像查询失败")
		}
		latestVersion := QueryReleaseLatestVersion(it, releaseRequest.Version)
		if latestVersion != "" {
			// 替换yaml中{{image}}并将yaml挪到指定位置
			imageName := fmt.Sprintf("%s%s:%s", releaseRequest.Domain, name, latestVersion)
			log.Println("最新镜像: " + imageName)
			err = utils.MoveYamlToReleaseDir(releaseRequest.TemplatePath, releaseRequest.ReleasePath, imageName, imagePathMap[name])
			if err != nil {
				log.Fatal("yaml迁移失败：" + err.Error())
			}
		}
	}
	if releaseRequest.Apply {
		log.Println("此命令需要在kubernetes master节点执行")
		log.Println(fmt.Sprintf("kubectl delete -f %s && kubectl apply -f %s", releaseRequest.ReleasePath, releaseRequest.ReleasePath))
		cmd := exec.Command("kubectl", "delete", "-f", releaseRequest.ReleasePath, "&&", "kubectl", "apply", "-f", releaseRequest.ReleasePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("执行kubectl 命令失败：" + err.Error())
		}
	}
	log.Println("release命令执行完成，生成yaml文件目录：" + releaseRequest.ReleasePath)
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
	log.Println("计算最新镜像版本...")
	if len(it) == 0 {
		return ""
	}

	latestImageTag := ""
	for _, t := range it {
		// 只处理符合v1.8.2-10 格式的数据
		if verify, _ := regexp.Match("^v(\\d+\\.?){2,3}-\\d+", []byte(t.ImageTag)); !verify {
			continue
		}
		if version != "" && strings.HasPrefix(t.ImageTag, version+"-") {
			latestImageTag = utils.VersionCompare(latestImageTag, t.ImageTag)
		}
		if version == "" {
			latestImageTag = utils.VersionCompare(latestImageTag, t.ImageTag)
		}
	}
	return latestImageTag
}
