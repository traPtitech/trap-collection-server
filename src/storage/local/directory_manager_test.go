package local

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"go.uber.org/mock/gomock"
)

func TestSetupDirectory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	rootPath := "./directory_manager_test"
	mockConf := mock.NewMockStorageLocal(ctrl)
	mockConf.
		EXPECT().
		Path().
		Return(rootPath, nil)
	directoryManager, err := NewDirectoryManager(mockConf)
	if err != nil {
		t.Fatalf("failed to create directory manager: %v", err)
		return
	}
	defer func() {
		err := os.RemoveAll(rootPath)
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	type test struct {
		description        string
		directoryName      string
		isDirectoryExist   bool
		isFileExist        bool
		isAlreadyFileExist bool
		isErr              bool
		err                error
	}

	testCases := []test{
		{
			description:   "ディレクトリがなくてもディレクトリを作成する",
			directoryName: "a",
		},
		{
			description:      "ディレクトリがあるので何もしない",
			directoryName:    "b",
			isDirectoryExist: true,
		},
		{
			description:   "ファイルがあるのでエラー",
			directoryName: "c",
			isFileExist:   true,
			isErr:         true,
		},
		{
			description:        "ディレクトリ内にファイルがあっても壊れない",
			directoryName:      "d",
			isAlreadyFileExist: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			files := []string{}
			if testCase.isDirectoryExist {
				err := os.Mkdir(string(rootPath)+"/"+testCase.directoryName, 0777)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}

				if testCase.isAlreadyFileExist {
					_, err := os.Create(string(rootPath) + "/" + testCase.directoryName + "/file")
					if err != nil {
						t.Fatalf("failed to create file: %v", err)
					}

					files = append(files, "file")
				}
			}

			if testCase.isFileExist {
				_, err := os.Create(string(rootPath) + "/" + testCase.directoryName)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			directoryPath, err := directoryManager.setupDirectory(testCase.directoryName)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, path.Join(string(rootPath), testCase.directoryName), directoryPath)

			entries, err := os.ReadDir(string(rootPath) + "/" + testCase.directoryName)
			if err != nil {
				t.Fatalf("failed to stat directory: %v", err)
			}

			actualFiles := []string{}
			for _, entry := range entries {
				actualFiles = append(actualFiles, entry.Name())
			}

			assert.Equal(t, files, actualFiles)
		})
	}
}
