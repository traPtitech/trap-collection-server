package repository

import (
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"

	"mime/multipart"
	"runtime"
)

//UploadGame ゲームをConoHaオブジェクトストレージにアップロードする関数
func UploadGame(gameFile *multipart.FileHeader, fileName string, containerName string) error {
	file, err := gameFile.Open()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	opt := objects.CreateOpts{
		Content: file,
	}

	result := objects.Create(client, containerName, fileName, opt)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

//DownloadGame ゲームをConoHaオブジェクトストレージからダウンロードする関数
func DownloadGame(name string) ([]byte, error) {
	container, fileName, err := GetContainerAndFileName(name)
	if err != nil {
		return nil, err
	}
	result := objects.Download(client, container, fileName, nil)
	if result.Err != nil {
		return nil, result.Err
	}
	content, err := result.ExtractContent()
	if err != nil {
		return nil, err
	}
	return content, nil
}

//DeleteGameFromConoHa ゲームをConoHaオブジェクトストレージから削除する関数
func DeleteGameFromConoHa(name string) error {
	container, fileName, err := GetContainerAndFileName(name)
	if err != nil {
		return err
	}
	result := objects.Delete(client, container, fileName, nil)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

//GetGameList 全ゲーム用コンテナ内のゲーム一覧
func GetGameList() ([]string, error) {
	var containerList []string
	Db.Select(&containerList, "SELECT DISTINCT container FROM game")
	l := len(containerList)
	var gameList []string
	chErr := make(chan error)
	chGameList := make(chan []string)
	for _, v := range containerList {
		go syncGameList(v, chErr, chGameList)
	}
	for i := 0; i < l; i++ {
		err := <-chErr
		if err != nil {
			return nil, err
		}
		result := <-chGameList
		gameList = append(gameList, result...)
	}
	return gameList, nil
}

func syncGameList(container string, chErr chan error, chGameList chan []string) {
	page, err := objects.List(client, container, objects.ListOpts{}).AllPages()
	if err != nil {
		chErr <- err
		runtime.Goexit()
	}
	result, err := objects.ExtractNames(page)
	if err != nil {
		chErr <- err
		runtime.Goexit()
	}
	chGameList <- result
}

//GetContainerList コンテナ一覧を取得する関数
func GetContainerList() ([]containers.Container, error) {
	listOpt := containers.ListOpts{
		Full:   true,
		Prefix: "conoha_",
	}

	page, err := containers.List(client, listOpt).AllPages()
	if err != nil {
		return nil, err
	}
	containerList, err := containers.ExtractInfo(page)
	if err != nil {
		return nil, err
	}
	return containerList, nil
}

//IsThereFile ConoHaのオブジェクトストレージ内に同名ファイルが存在するか調べる関数
func IsThereFile(name string) (bool, error) {
	gameList, err := GetGameList()
	if err != nil {
		return false, err
	}
	for _, x := range gameList {
		if x == name {
			return true, nil
		}
	}
	return false, nil
}
