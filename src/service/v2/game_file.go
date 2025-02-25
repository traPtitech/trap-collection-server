package v2

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"golang.org/x/sync/errgroup"
)

type GameFile struct {
	db                 repository.DB
	gameRepository     repository.GameV2
	gameFileRepository repository.GameFileV2
	gameFileStorage    storage.GameFile
}

func NewGameFile(
	db repository.DB,
	gameRepository repository.GameV2,
	gameFileRepository repository.GameFileV2,
	gameFileStorage storage.GameFile,
) *GameFile {
	return &GameFile{
		db:                 db,
		gameRepository:     gameRepository,
		gameFileRepository: gameFileRepository,
		gameFileStorage:    gameFileStorage,
	}
}

func (*GameFile) checkZip(_ context.Context, reader io.Reader) (zr *zip.Reader, ok bool, err error) {
	f, err := os.CreateTemp("", "game_file")
	if err != nil {
		return nil, false, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		err = os.Remove(f.Name())
	}()
	defer func() {
		err = f.Close()
	}()

	_, err = io.Copy(f, reader)
	if err != nil {
		return nil, false, fmt.Errorf("failed to copy file: %w", err)
	}

	fInfo, err := f.Stat()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get file info: %w", err)
	}
	zr, err = zip.NewReader(f, fInfo.Size())
	if errors.Is(err, zip.ErrFormat) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to open zip file: %w", err)
	}

	return zr, true, nil
}

func zipFileContains(zr *zip.Reader, filePath string, isDir bool) bool {
	return slices.ContainsFunc(zr.File, func(zf *zip.File) bool {
		if isDir {
			return path.Clean(zf.Name) == filePath && zf.FileInfo().IsDir()
		}
		return zf.Name == filePath && !zf.FileInfo().IsDir()
	})
}

// エントリーポイントが存在し、それがディレクトリでないことを確認。
// 一般的なエントリーポイントの存在確認に使う。
func (*GameFile) checkEntryPointExist(_ context.Context, zr *zip.Reader, entryPoint values.GameFileEntryPoint) (bool, error) {
	entryPointExists := zipFileContains(zr, string(entryPoint), false)

	if !entryPointExists {
		return false, nil
	}

	return true, nil
}

// macOSのアプリケーション(*.app)のエントリーポイントが正しいか確認。
// 仕様は [Appleの開発者向けページ] を参照。
//
// 具体的には、
//   - エントリーポイントがディレクトリで .app で終わること
//   - エントリーポイント/Contents/MacOS というディレクトリが存在すること
//   - エントリーポイント/Contents/Info.plist というファイルが存在すること
//
// [Appleの開発者向けページ]: https://developer.apple.com/library/archive/documentation/CoreFoundation/Conceptual/CFBundles/BundleTypes/BundleTypes.html#//apple_ref/doc/uid/10000123i-CH101-SW1
func (*GameFile) checkMacOSAppEntryPointValid(_ context.Context, zr *zip.Reader, entryPoint values.GameFileEntryPoint) (bool, error) {
	if !strings.HasSuffix(string(entryPoint), ".app") || !zipFileContains(zr, string(entryPoint), true) {
		return false, nil
	}

	requiredDirs := []string{
		path.Join(string(entryPoint), "Contents"),
		path.Join(string(entryPoint), "Contents", "MacOS"),
	}

	for _, dir := range requiredDirs {
		if !zipFileContains(zr, dir, true) {
			return false, nil
		}
	}

	requiredFiles := []string{
		path.Join(string(entryPoint), "Contents", "Info.plist"),
	}

	for _, file := range requiredFiles {
		if !zipFileContains(zr, file, false) {
			return false, nil
		}
	}

	return true, nil
}

func (gameFile *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, gameID values.GameID, fileType values.GameFileType, entryPoint values.GameFileEntryPoint) (*domain.GameFile, error) {

	var file *domain.GameFile
	err := gameFile.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		fileID := values.NewGameFileID()

		eg, ctx := errgroup.WithContext(ctx)
		hashPr, hashPw := io.Pipe()
		filePr, filePw := io.Pipe()
		entryPointPr, entryPointPw := io.Pipe()

		eg.Go(func() error {
			defer hashPr.Close()

			hash, err := values.NewGameFileHash(hashPr)
			if err != nil {
				return fmt.Errorf("failed to get hash: %w", err)
			}

			file = domain.NewGameFile(
				fileID,
				fileType,
				entryPoint,
				hash,
				time.Now(),
			)

			err = gameFile.gameFileRepository.SaveGameFile(ctx, gameID, file)
			if err != nil {
				return fmt.Errorf("failed to save game file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gameFile.gameFileStorage.SaveGameFile(ctx, filePr, fileID)
			if err != nil {
				return fmt.Errorf("failed to save game file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer entryPointPr.Close()

			zr, ok, err := gameFile.checkZip(ctx, entryPointPr)
			if err != nil {
				return fmt.Errorf("failed to check zip: %w", err)
			}
			if !ok {
				return service.ErrNotZipFile
			}

			// これらのどれか一つで成功した場合(trueが返ってきた場合)、有効なエントリーポイントとして扱う
			checkers := []func(context.Context, *zip.Reader, values.GameFileEntryPoint) (bool, error){
				gameFile.checkEntryPointExist,
				gameFile.checkMacOSAppEntryPointValid,
			}
			for _, checker := range checkers {
				ok, err = checker(ctx, zr, entryPoint)
				if err != nil {
					return fmt.Errorf("failed to check entry point: %w", err)
				}
				if ok {
					return nil
				}
			}

			return service.ErrInvalidEntryPoint
		})

		eg.Go(func() error {
			defer hashPw.Close()
			defer filePw.Close()
			defer entryPointPw.Close()

			mw := io.MultiWriter(hashPw, filePw, entryPointPw)
			_, err = io.Copy(mw, reader)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

			return nil
		})

		err = eg.Wait()
		if err != nil {
			return fmt.Errorf("failed to save game file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return file, nil
}

func (gameFile *GameFile) GetGameFile(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (values.GameFileTmpURL, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	var url *url.URL
	err = gameFile.db.Transaction(ctx, nil, func(ctx context.Context) error {
		file, err := gameFile.gameFileRepository.GetGameFile(ctx, fileID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameFileID
		}
		if err != nil {
			return fmt.Errorf("failed to get game file: %w", err)
		}

		if file.GameID != gameID {
			// gameIdに対応したゲームにゲームファイルが紐づいていない場合も、
			// 念の為閲覧権限がないゲームに紐づいたファイルIDを知ることができないようにするため、
			// ファイルが存在しない場合と同じErrInvalidGameFileIDを返す
			return service.ErrInvalidGameFileID
		}

		url, err = gameFile.gameFileStorage.GetTempURL(ctx, file.GameFile, time.Minute)
		if err != nil {
			return fmt.Errorf("failed to get game file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return url, nil
}

func (gameFile *GameFile) GetGameFiles(ctx context.Context, gameID values.GameID) ([]*domain.GameFile, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameFiles, err := gameFile.gameFileRepository.GetGameFiles(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game files: %w", err)
	}

	return gameFiles, nil
}

func (gameFile *GameFile) GetGameFileMeta(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (*domain.GameFile, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	file, err := gameFile.gameFileRepository.GetGameFile(ctx, fileID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameFileID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game file: %w", err)
	}

	if file.GameID != gameID {
		// gameIdに対応したゲームにゲームファイルが紐づいていない場合も、
		// 念の為閲覧権限がないゲームに紐づいたファイルIDを知ることができないようにするため、
		// ファイルが存在しない場合と同じErrInvalidGameFileIDを返す
		return nil, service.ErrInvalidGameFileID
	}

	return file.GameFile, nil
}
