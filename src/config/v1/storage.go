package v1

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/traPtitech/trap-collection-server/src/config"
)

type Storage struct{}

func NewStorage() *Storage {
	return &Storage{}
}

func (*Storage) Type() (config.StorageType, error) {
	storage, ok := os.LookupEnv(envKeyStorage)
	if !ok {
		return config.StorageTypeSwift, nil
	}

	switch storage {
	case "swift":
		return config.StorageTypeSwift, nil
	case "local":
		return config.StorageTypeLocal, nil
	}

	return 0, errors.New("invalid storage")
}

type StorageSwift struct{}

func NewStorageSwift() *StorageSwift {
	return &StorageSwift{}
}

func (*StorageSwift) AuthURL() (*url.URL, error) {
	strSwiftAuthURL, ok := os.LookupEnv(envKeySwiftAuthURL)
	if !ok {
		return nil, errors.New("OS_AUTH_URL is not set")
	}

	swiftAuthURL, err := url.Parse(strSwiftAuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse swiftAuthURL: %w", err)
	}

	return swiftAuthURL, nil
}

func (*StorageSwift) UserName() (string, error) {
	swiftUserName, ok := os.LookupEnv(envKeySwiftUserName)
	if !ok {
		return "", errors.New("OS_USERNAME is not set")
	}

	return swiftUserName, nil
}

func (*StorageSwift) Password() (string, error) {
	swiftPassword, ok := os.LookupEnv(envKeySwiftPassword)
	if !ok {
		return "", errors.New("OS_PASSWORD is not set")
	}

	return swiftPassword, nil
}

func (*StorageSwift) TenantID() (string, error) {
	swiftTenantID, ok := os.LookupEnv(envKeySwiftTenantID)
	if !ok {
		return "", errors.New("OS_TENANT_ID is not set")
	}

	return swiftTenantID, nil
}

func (*StorageSwift) TenantName() (string, error) {
	swiftTenantName, ok := os.LookupEnv(envKeySwiftTenantName)
	if !ok {
		return "", errors.New("OS_TENANT_NAME is not set")
	}

	return swiftTenantName, nil
}

func (*StorageSwift) Container() (string, error) {
	swiftContainer, ok := os.LookupEnv(envKeySwiftContainer)
	if !ok {
		return "", errors.New("OS_CONTAINER is not set")
	}

	return swiftContainer, nil
}

func (*StorageSwift) TmpURLKey() (string, error) {
	swiftTmpURLKey, ok := os.LookupEnv(envKeySwiftTmpURLKey)
	if !ok {
		return "", errors.New("OS_TMP_URL_KEY is not set")
	}

	return swiftTmpURLKey, nil
}

type StorageLocal struct{}

func NewStorageLocal() *StorageLocal {
	return &StorageLocal{}
}

func (*StorageLocal) Path() (string, error) {
	filePath, ok := os.LookupEnv(envKeyFilePath)
	if !ok {
		return "", errors.New("FILE_PATH is not set")
	}

	return filePath, nil
}
