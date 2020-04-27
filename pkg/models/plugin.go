package models

type Plugins struct {
	Plugins []*Plugin `json:"plugins"`
}

type Plugin struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Version        string `json:"version"`
	InstallLogFile string `json:"install-log-file"`
}

type PluginList struct {
	Plugins []*PluginDetail `json:"plugins"`
}

type PluginDetail struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Versions    []*PluginVersion `json:"versions"`
}

type PluginVersion struct {
	Version     string `json:"version"`
	DownloadUrl string `json:"download-url"`
}

type PluginSearch struct {
	Url     string // 插件地址
	Name    string // 插件名称
	Version string // 插件版本
}
