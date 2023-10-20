package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetDirectoryFiles 不遍历文件夹 只获取当前文件目录下的文件
func GetDirectoryFiles(dir string) ([]string, error) {

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			fullPath := filepath.Join(dir, entry.Name())
			files = append(files, fullPath)
		}
	}

	return files, nil
}

// GetFilesRecursively 递归地获取指定文件夹下的文件列表
func GetFilesRecursively(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// GetDirFiles 通过标记决定返回当前目录下的文件 还是 遍历当前目录下的所有文件
// flag true 则只遍历本层目录下的文件
func GetDirFiles(dir string, flag bool) ([]string, error) {
	if flag {
		return GetDirectoryFiles(dir)
	} else {
		return GetFilesRecursively(dir)
	}
}

// IsFileType 判断文件类型是否是指定类型
func IsFileType(filePath string, fileTypeList []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, fileType := range fileTypeList {
		if ext == strings.ToLower(fileType) {
			return true
		}
	}
	return false
}

// DeleteTypeFiles 删除文件
func DeleteTypeFiles(dirs []string, filetypeList []string) []string {
	var removeDirs []string
	for _, file := range dirs {
		isType := IsFileType(file, filetypeList)
		if isType {
			_ = os.Remove(file)
		} else {
			removeDirs = append(removeDirs, file)
		}
	}
	return removeDirs
}

// RemoveTypeFiles 删除文件
func RemoveTypeFiles(dirs []string, filetypeList []string) []string {
	var removeDirs []string
	for _, file := range dirs {
		isType := IsFileType(file, filetypeList)
		if isType {
			removeDirs = append(removeDirs, file)
		}
	}
	return removeDirs
}

type FileInfoStruct struct {
	Name       string    `json:"name"`
	ModTime    time.Time `json:"modTime"`
	FilePath   string    `json:"filePath"`
	ReName     string    `json:"reName"`
	ReFilePath string    `json:"reFilePath"`
}

// GetFileInfo 获取文件详细信息
func GetFileInfo(filePath string, filePrefix []string) (FileInfoStruct, bool) {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Println(filePath + "文件不存在")
	}

	isPrefix := false
	for _, prefix := range filePrefix {
		if strings.HasPrefix(fileInfo.Name(), prefix) {
			isPrefix = true
			break
		}
	}

	var fileInfoStruct FileInfoStruct
	if isPrefix {
		fileInfoStruct.Name = fileInfo.Name()
		fileInfoStruct.ModTime = fileInfo.ModTime()
		fileInfoStruct.FilePath = filePath
	}
	return fileInfoStruct, isPrefix
}

// GetPrefixFileInfo 获取指定文件名前缀的文件信息
func GetPrefixFileInfo(directories []string, prefixes []string) []FileInfoStruct {
	var fileInfoStructList []FileInfoStruct
	for _, directory := range directories {
		fileInfoOut, isPrefix := GetFileInfo(directory, prefixes)
		if isPrefix {
			fileInfoStructList = append(fileInfoStructList, fileInfoOut)
		}
	}
	return fileInfoStructList
}

// ReviseFileInfo 按照需求添加需要移动和重命名文件的信息
func ReviseFileInfo(fileInfoStructList []FileInfoStruct, isCreateFolderFlag bool, folderLayout string, outFilePath string, fileLayout string) []FileInfoStruct {
	var fileInfoStructListNew []FileInfoStruct
	if isCreateFolderFlag {
		for _, fileInfoStruct := range fileInfoStructList {
			ext := filepath.Ext(fileInfoStruct.Name)
			reName := fileInfoStruct.ModTime.Format(fileLayout) + ext
			reFilePath := outFilePath + "/" + fileInfoStruct.ModTime.Format(folderLayout) + "/"

			fileInfoStruct.ReName = reName
			fileInfoStruct.ReFilePath = reFilePath
			fileInfoStructListNew = append(fileInfoStructListNew, fileInfoStruct)
		}
	}
	return fileInfoStructListNew
}

// CreateFolder 创建文件夹
func CreateFolder(destinationPath string) {
	// 创建目标文件夹
	destinationFolder := filepath.Dir(destinationPath)
	err := os.MkdirAll(destinationFolder, os.ModePerm)
	if err != nil {
		fmt.Println("创建目标文件夹失败:", err)
	}
}

// ProcessFile 真正的对文件进行处理
func ProcessFile(fileInfoStructList []FileInfoStruct) {
	for _, fileInfoStruct := range fileInfoStructList {
		CreateFolder(fileInfoStruct.ReFilePath)
		err := os.Rename(fileInfoStruct.FilePath, fileInfoStruct.ReFilePath+fileInfoStruct.ReName)
		if err != nil {
			fmt.Println("移动文件失败:", err)
		}
	}
}

type ConfigData struct {
	// 需要处理的文件目录
	Path string `json:"path"`
	// 是否递归目录
	Prefixes []string `json:"prefixes"`
	// 需要处理的文件头
	IsRecursive bool `json:"isRecursive"`
	// 需要处理的文件后缀
	RemoveFileType []string `json:"removeFileType"`
	// 需要删除的文件后缀
	DeleteFileType []string `json:"deleteFileType"`
	// 是否新建文件夹
	IsCreateFolderFlag bool `json:"isCreateFolderFlag"`
	// 创建文件夹名规则
	FolderLayout string `json:"folderLayout"`
	// 重命名文件名规则
	FileLayout string `json:"fileLayout"`
	// 输出文件位置
	OutFilePath string `json:"outFilePath"`
}

// ParseConfig 解析配置文件
func ParseConfig() ConfigData {
	var data ConfigData

	// 读取 JSON 文件
	jsonData, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal("读取 JSON 文件失败:", err)
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Fatal("解析 JSON 数据失败:", err)
	}

	return data
}

// ExportJson 写出json
func ExportJson(fileInfo []FileInfoStruct) bool {
	// 将结构体转换为 JSON 字符串
	jsonData, err := json.MarshalIndent(fileInfo, "", "	")
	if err != nil {
		log.Fatal("转换为 JSON 失败:", err)
	}
	timeString := time.Now().Format("2006-01-02_15-04-05")
	// 将 JSON 数据写入文件
	err = os.WriteFile(timeString+"_output.json", jsonData, 0644)
	if err != nil {
		log.Fatal("写入文件失败:", err)
		return false
	}
	return true
}

func main() {
	config := ParseConfig()

	directories, err := GetDirFiles(config.Path, false)
	if err != nil {
		log.Fatal(err)
	}

	directories = DeleteTypeFiles(directories, config.DeleteFileType)

	directories = RemoveTypeFiles(directories, config.RemoveFileType)

	fileInfoStructList := GetPrefixFileInfo(directories, config.Prefixes)

	fileInfoStructListNew := ReviseFileInfo(fileInfoStructList, config.IsCreateFolderFlag, config.FolderLayout, config.OutFilePath, config.FileLayout)

	// 只有json 写入成功才能对文件进行操作
	if ExportJson(fileInfoStructListNew) {
		ProcessFile(fileInfoStructListNew)
	}
}
