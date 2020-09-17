package storage

import (
	"fmt"
	"io"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
)

// Swift オブジェクトストレージへのアップロード用の構造体
type Swift struct {
	client    *gophercloud.ServiceClient
	container string
}

// NewSwiftStorage Swiftのコンストラクタ
func NewSwiftStorage(container string) (*Swift, error) {
	option, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return &Swift{}, fmt.Errorf("Failed In Reading Auth Env:%w", err)
	}

	provider, err := openstack.AuthenticatedClient(option)
	if err != nil {
		return &Swift{}, fmt.Errorf("Failed In Authorization:%w", err)
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return &Swift{}, fmt.Errorf("Failed In Reading Connecting To Storage:%w", err)
	}

	result := containers.Create(client, container, nil)
	if result.Err != nil {
		return &Swift{}, fmt.Errorf("Failed In Making New Storage:%w", err)
	}

	swift := &Swift{
		client:    client,
		container: container,
	}

	return swift, nil
}

// Save ファイルの保存
func (s *Swift) Save(fileName string, src io.Reader) error {
	opt := objects.CreateOpts{
		Content: src,
	}

	result := objects.Create(s.client, s.container, fileName, opt)
	if result.Err != nil {
		return fmt.Errorf("Failed In Saving File: %w", result.Err)
	}

	return nil
}

// Open ファイルを開く
func (s *Swift) Open(fileName string) (io.ReadCloser, error) {
	result := objects.Download(s.client, s.container, fileName, nil)
	if result.Err != nil {
		return nil, fmt.Errorf("Failed In Downloading File: %w", result.Err)
	}

	return result.Body, nil
}

// Delete ファイルの削除
func (s *Swift) Delete(fileName string) error {
	result := objects.Delete(s.client, s.container, fileName, nil)
	if result.Err != nil {
		return fmt.Errorf("Failed In Deleting File: %w", result.Err)
	}

	return nil
}
