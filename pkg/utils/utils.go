package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func VersionCompare(source, dst string) string {
	// 版本号情况v1.10.0-78 v1.3-14 v1.2.0-4
	// 从这三类数据选出最后版本，从左到右依次比较
	if source == "" {
		_, _, _, _, err := disassembleVersion(dst)
		if err != nil {
			return ""
		}
		return dst
	}
	// 源tag经过校验，拆解不会出现错误
	s1, s2, s3, s4, _ := disassembleVersion(source)
	d1, d2, d3, d4, err := disassembleVersion(dst)
	if err != nil {
		return source
	}
	// 依次从v1-v4进行判断
	if d1 > s1 {
		return dst
	}
	if d1 == s1 && d2 > s2 {
		return dst
	}
	if d1 == s1 && d2 == s2 && d3 > s3 {
		return dst
	}
	if d1 == s1 && d2 == s2 && d3 == s3 && d4 > s4 {
		return dst
	}
	return source
}

func disassembleVersion(version string) (int, int, int, int, error) {
	version1 := strings.Split(version[1:], "-")
	version2 := strings.Split(version1[0], ".")
	v1, v2, v3, v4 := -1, -1, -1, -1
	v4, err := strconv.Atoi(version1[1])
	if err != nil {
		return v1, v2, v3, v4, err
	}
	v1, err = strconv.Atoi(version2[0])
	if err != nil {
		return v1, v2, v3, v4, err
	}
	v2, err = strconv.Atoi(version2[1])
	if err != nil {
		return v1, v2, v3, v4, err
	}
	if len(version2) == 3 {
		v3, err = strconv.Atoi(version2[2])
		if err != nil {
			return v1, v2, v3, v4, err
		}
	}
	return v1, v2, v3, v4, nil
}

// 将模板目录下的yaml文件迁移到release目录下
func MoveYamlToReleaseDir(sourceDirPrefix, dstDirPrefix string, imageName string, yamlPath string) error {
	log.Println("读取模板yaml文件: " + yamlPath)
	b, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		log.Println("读取模板yaml文件失败：" + err.Error())
		return err
	}

	s := string(b)
	s1 := strings.Replace(s, "{{image}}", imageName, -1)

	transitionSourceDirPrefix := strings.Replace(sourceDirPrefix, "\\\\", "\\", -1)
	transitionDstDirPrefix := strings.Replace(dstDirPrefix, "\\\\", "\\", -1)
	releaseFile := strings.Replace(yamlPath, transitionSourceDirPrefix, transitionDstDirPrefix, 1)

	imagePath := strings.Split(imageName, "/")
	filename := strings.Split(imagePath[len(imagePath)-1], ":")
	filepath := strings.Replace(releaseFile, fmt.Sprintf("%s.yaml", filename[0]), "", 1)
	_, err = os.Stat(filepath)
	if err != nil {
		err = os.MkdirAll(filepath, os.ModePerm)
		if err != nil {
			log.Println("创建release yaml目录失败:" + err.Error())
			return err
		}
	}
	log.Println("生成yaml文件：" + releaseFile)
	err = ioutil.WriteFile(releaseFile, []byte(s1), os.ModePerm)
	if err != nil {
		log.Println("yaml文件写入失败: " + err.Error())
		return err
	}
	return nil
}
