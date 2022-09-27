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
	case "s3":
		return config.StorageTypeS3, nil
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

type StorageS3 struct{}

func NewStorageS3() *StorageS3 {
	return &StorageS3{}
}

func (*StorageS3) AccessKeyID() (string, error) {
	s3AccessKeyID, ok := os.LookupEnv(envKeyS3AccessKeyID)
	if !ok {
		return "", errors.New("S3_ACCESS_KEY_ID is not set")
	}

	return s3AccessKeyID, nil
}

func (*StorageS3) SecretAccessKey() (string, error) {
	s3SecretAccessKey, ok := os.LookupEnv(envKeyS3SecretAccessKey)
	if !ok {
		return "", errors.New("S3_SECRET_ACCESS_KEY is not set")
	}

	return s3SecretAccessKey, nil
}

func (*StorageS3) Region() (string, error) {
	s3Region, ok := os.LookupEnv(envKeyS3Region)
	if !ok {
		return "", errors.New("S3_REGION is not set")
	}

	return s3Region, nil
}

func (*StorageS3) Bucket() (string, error) {
	s3Bucket, ok := os.LookupEnv(envKeyS3Bucket)
	if !ok {
		return "", errors.New("S3_BUCKET is not set")
	}

	return s3Bucket, nil
}

func (*StorageS3) Endpoint() (string, error) {
	s3Endpoint, ok := os.LookupEnv(envKeyS3Endpoint)
	if !ok {
		return "", errors.New("S3_ENDPOINT is not set")
	}

	return s3Endpoint, nil
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
