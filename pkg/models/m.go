package models

const (
	Nexus     = "nexus"
	Harbor    = "harbor"
	DockerHub = "dockerhub"
)

// app search 请求参数
type Search struct {
	Type    string // 仓库类型 nexus/harbor/dockerHub 默认dockerHub
	Url     string // 仓库地址 http://username:password@repository
	Name    string // 要搜索的镜像名称
	Version string // 指定过滤版本
}

// app release 请求参数
type Release struct {
	Type         string // 仓库类型 nexus/harbor/dockerHub 默认dockerHub
	Url          string // 仓库地址 http://username:password@xxxx
	TemplatePath string // 模板文件路径
	ReleasePath  string // 生成yam路径
	Apply        bool   // 是否自动执行kubectl apply -f
	Version      string // 指定过滤版本
	Prefix       string // 镜像地址前半部分 + 后半部分来源于文件名
	Domain       string // 指定生成镜像的域名
}

// dockerHub Response Tags
type DockerHubTags struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []*DockerHubTag `json:"results"`
}

type DockerHubTag struct {
	Name                string            `json:"name"`
	FullSize            int               `json:"full_size"`
	Images              []*DockerHubImage `json:"images"`
	Id                  int               `json:"id"`
	Repository          int               `json:"repository"`
	Creator             int               `json:"creator"`
	LastUpdater         int               `json:"last_updater"`
	LastUpdaterUserName string            `json:"last_updater_user_name"`
	V2                  bool              `json:"v2"`
	LastUpdated         string            `json:"last_updated"`
}

type DockerHubImage struct {
	Size         int    `json:"size"`
	Digest       string `json:"digest"`
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
	Variant      string `json:"variant"`
}

type ImageTags struct {
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
	Source    string `json:"source"`
}

type NexusTags struct {
	Items []*NexusTag `json:"items"`
}

type NexusTag struct {
	Id         string         `json:"id"`
	Repository string         `json:"repository"`
	Format     string         `json:"format"`
	Group      string         `json:"group"`
	Name       string         `json:"name"`
	Version    string         `json:"version"`
	Assets     []*NexusAssets `json:"assets"`
}

type NexusAssets struct {
	DownloadUrl string        `json:"downloadUrl"`
	Path        string        `json:"path"`
	Id          string        `json:"id"`
	Repository  string        `json:"repository"`
	Format      string        `json:"format"`
	Checksum    NexusCheckSum `json:"checksum"`
}

type NexusCheckSum struct {
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
}

type HarborTag struct {
	Digest        string `json:"digest"`
	Name          string `json:"name"`
	Size          int    `json:"size"`
	Architecture  string `json:"architecture"`
	Os            string `json:"os"`
	DockerVersion string `json:"docker_version"`
	Author        string `json:"author"`
	Created       string `json:"created"`
	PushTime      string `json:"push_time"`
	PullTime      string `json:"pull_time"`
}
