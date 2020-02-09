package repository

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antmoveh/micro-version-management/pkg/models"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"strings"
)

func DockerHubTags(searchRequest *models.Search) ([]*models.ImageTags, error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags?page_size=100&&page=paasos-e4-api.yaml", searchRequest.Name)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)

	var tags models.DockerHubTags

	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, errors.New("未能找到匹配镜像")
	}
	var it []*models.ImageTags
	for _, t := range tags.Results {
		it = append(it, &models.ImageTags{
			ImageName: searchRequest.Name,
			ImageTag:  t.Name,
			Source:    models.DockerHub,
		})
	}
	return it, nil
}

func NexusSearchTags(searchRequest *models.Search) ([]*models.ImageTags, error) {

	userName, password, url := spiltLink(searchRequest.Url)

	tagUrl := fmt.Sprintf("%s/service/rest/v1/search?docker.imageName=%s", url, searchRequest.Name)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	cookie := ""
	if userName != "" && password != "" {
		loginUrl := fmt.Sprintf("%s/service/rapture/session", url)

		data := make(url2.Values)
		data["username"] = []string{base64.StdEncoding.EncodeToString([]byte(userName))}
		data["password"] = []string{base64.StdEncoding.EncodeToString([]byte(password))}
		resp1, err := http.PostForm(loginUrl, data)
		if err != nil {
			return nil, err
		}
		defer resp1.Body.Close()
		cookie = resp1.Header.Get("Set-Cookie")
	}

	req, err := http.NewRequest("GET", tagUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var tags models.NexusTags

	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	var it []*models.ImageTags
	for _, t := range tags.Items {
		it = append(it, &models.ImageTags{
			ImageName: searchRequest.Name,
			ImageTag:  t.Version,
			Source:    models.Nexus,
		})
	}
	return it, nil
}

func HarborTags(searchRequest *models.Search) ([]*models.ImageTags, error) {
	userName, password, url := spiltLink(searchRequest.Url)
	tagUrl := fmt.Sprintf("%s/api/repositories/%s/tags?detail=false", url, searchRequest.Name)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", tagUrl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(userName, password)
	//req.Header.Add("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var tags []*models.HarborTag

	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	var it []*models.ImageTags
	for _, t := range tags {
		it = append(it, &models.ImageTags{
			ImageName: searchRequest.Name,
			ImageTag:  t.Name,
			Source:    models.Harbor,
		})
	}
	return it, nil
}

func spiltLink(url string) (string, string, string) {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	if !strings.Contains(url, "@") {
		return "", "", url
	} else {
		a := strings.Split(url, "@")
		b := strings.Split(a[0], "//")
		c := strings.Split(b[1], ":")
		return c[0], c[1], b[0] + "//" + a[1]
	}
}
