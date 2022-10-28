package v2

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateEdition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		name             values.LauncherVersionName
		questionnaireURL types.Option[values.LauncherVersionQuestionnaireURL]
		gameVersionIDs   []values.GameVersionID
	}
	type mockInfo struct {
		gameVersions []*repository.GameVersionInfoWithGameID

		executeGetGameVersionsByIDs      bool
		executeSaveEdition               bool
		executeUpdateEditionGameVersions bool

		errGetGameVersionsByIDs      error
		errSaveEdition               error
		errUpdateEditionGameVersions error
	}

	type test struct {
		description     string
		args            args
		mockInfo        mockInfo
		expectedEdition *domain.LauncherVersion
		isErr           bool
		err             error
	}

	urlStr := "https://example.com"
	urlLink, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	name := values.LauncherVersionName("v1.0.0")
	now := time.Now()

	gameVersionIDs1, gameVersions1 := generateGameVersionsForEditionTests(t, 1)
	gameVersionIDs2, gameVersions2 := generateGameVersionsForEditionTests(t, 1)
	gameVersionIDs3, gameVersions3 := generateGameVersionsForEditionTests(t, 2)
	gameVersionIDs5, _ := generateGameVersionsForEditionTests(t, 1)
	gameVersionIDs6, _ := generateGameVersionsForEditionTests(t, 1)
	gameVersionIDs7, gameVersions7 := generateGameVersionsForEditionTests(t, 1)
	gameVersionIDs8, gameVersions8 := generateGameVersionsForEditionTests(t, 1)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs1,
			},
			mockInfo: mockInfo{
				gameVersions: gameVersions1,

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: true,
			},
			expectedEdition: domain.NewLauncherVersionWithQuestionnaire(values.NewLauncherVersionID(), name, values.NewLauncherVersionQuestionnaireURL(urlLink), now),
		},
		{
			description: "URLなしだがエラーなし",
			args: args{
				name:             name,
				questionnaireURL: types.Option[values.LauncherVersionQuestionnaireURL]{},
				gameVersionIDs:   gameVersionIDs2,
			},
			mockInfo: mockInfo{
				gameVersions: gameVersions2,

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: true,
			},
			expectedEdition: domain.NewLauncherVersionWithoutQuestionnaire(values.NewLauncherVersionID(), name, now),
		},
		{
			description: "gameVersionIDsの要素が複数だがエラーなし",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs3,
			},
			mockInfo: mockInfo{
				gameVersions: gameVersions3,

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: true,
			},
			expectedEdition: domain.NewLauncherVersionWithQuestionnaire(values.NewLauncherVersionID(), name, values.NewLauncherVersionQuestionnaireURL(urlLink), now),
		},
		{
			description: "gameVersionIDsが空だがエラーなし",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   []values.GameVersionID{},
			},
			mockInfo: mockInfo{
				gameVersions: []*repository.GameVersionInfoWithGameID{},

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: true,
			},
			expectedEdition: domain.NewLauncherVersionWithQuestionnaire(values.NewLauncherVersionID(), name, values.NewLauncherVersionQuestionnaireURL(urlLink), now),
		},
		{
			description: "GetGameVersionsByIDsがエラーなのでエラー",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs5,
			},
			mockInfo: mockInfo{
				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               false,
				executeUpdateEditionGameVersions: false,

				errGetGameVersionsByIDs: errors.New("error"),
			},
			isErr: true,
		},
		{
			description: "gameVersionsの数が違うのでエラー",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs6, // 要素数1
			},
			mockInfo: mockInfo{
				gameVersions: []*repository.GameVersionInfoWithGameID{},

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               false,
				executeUpdateEditionGameVersions: false,
			},
			isErr: true,
			err:   service.ErrInvalidGameVersionID,
		},
		{
			description: " SaveEditionでエラーなのでエラー",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs7,
			},
			mockInfo: mockInfo{
				gameVersions: gameVersions7,

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: false,

				errSaveEdition: errors.New("error"),
			},
			isErr: true,
		},
		{
			description: "  UpdateEditionGameVersionsでエラーなのでエラー",
			args: args{
				name:             name,
				questionnaireURL: types.NewOption[values.LauncherVersionQuestionnaireURL](urlLink),
				gameVersionIDs:   gameVersionIDs8,
			},
			mockInfo: mockInfo{
				gameVersions: gameVersions8,

				executeGetGameVersionsByIDs:      true,
				executeSaveEdition:               true,
				executeUpdateEditionGameVersions: true,

				errUpdateEditionGameVersions: errors.New("error"),
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mockRepository.NewMockDB(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)

			editionService := NewEdition(
				mockDB,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
			)

			if testCase.mockInfo.executeGetGameVersionsByIDs {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersionsByIDs(ctx, testCase.args.gameVersionIDs, repository.LockTypeRecord).
					Return(testCase.mockInfo.gameVersions, testCase.mockInfo.errGetGameVersionsByIDs)
			}
			if testCase.mockInfo.executeSaveEdition {
				mockEditionRepository.
					EXPECT().
					SaveEdition(ctx, gomock.Any()). // newEditionについてはmockできない
					Return(testCase.mockInfo.errSaveEdition)
			}
			if testCase.mockInfo.executeUpdateEditionGameVersions {
				mockEditionRepository.
					EXPECT().
					UpdateEditionGameVersions(ctx, gomock.Any(), testCase.args.gameVersionIDs).
					Return(testCase.mockInfo.errUpdateEditionGameVersions)
			}

			got, err := editionService.CreateEdition(ctx, testCase.args.name, testCase.args.questionnaireURL, testCase.args.gameVersionIDs)

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

			expected := testCase.expectedEdition
			// IDについてはチェックできない
			assert.Equal(t, expected.GetName(), got.GetName())

			expectedURL, expectedErr := expected.GetQuestionnaireURL()
			gotURL, gotErr := got.GetQuestionnaireURL()
			assert.Equal(t, expectedURL, gotURL)
			assert.Equal(t, expectedErr, gotErr)

			assert.WithinDuration(t, expected.GetCreatedAt(), got.GetCreatedAt(), time.Second*2)
		})
	}

}

func TestGetEditions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type mockInfo struct {
		editions []*domain.LauncherVersion

		errGetEditions error
	}

	type test struct {
		description      string
		mockInfo         mockInfo
		expectedEditions []*domain.LauncherVersion
		isErr            bool
		err              error
	}

	editions1 := generateEditions(t, true, 1)
	editions2 := generateEditions(t, false, 1)
	editions3 := generateEditions(t, true, 2)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			mockInfo: mockInfo{
				editions: editions1,
			},
			expectedEditions: editions1,
		},
		{
			description: "URLなしでもエラーなし",
			mockInfo: mockInfo{
				editions: editions2,
			},
			expectedEditions: editions2,
		},
		{
			description: "対象が複数でもエラーなし",
			mockInfo: mockInfo{
				editions: editions3,
			},
			expectedEditions: editions3,
		},
		{
			description: "GetEditionsがエラーなのでエラー",
			mockInfo: mockInfo{
				errGetEditions: errors.New("error"),
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mockRepository.NewMockDB(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)

			editionService := NewEdition(
				mockDB,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
			)

			mockEditionRepository.
				EXPECT().
				GetEditions(ctx, repository.LockTypeNone).
				Return(testCase.mockInfo.editions, testCase.mockInfo.errGetEditions)

			gotEditions, err := editionService.GetEditions(ctx)

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

			if !assert.Len(t, gotEditions, len(testCase.expectedEditions)) {
				return
			}
			for i, got := range gotEditions {
				expected := testCase.expectedEditions[i]

				assert.Equal(t, expected.GetID(), got.GetID())
				assert.Equal(t, expected.GetName(), got.GetName())

				expectedURL, expectedErr := expected.GetQuestionnaireURL()
				gotURL, gotErr := got.GetQuestionnaireURL()
				assert.Equal(t, expectedURL, gotURL)
				assert.Equal(t, expectedErr, gotErr)

				assert.WithinDuration(t, expected.GetCreatedAt(), got.GetCreatedAt(), time.Second*2)
			}
		})
	}
}

func TestGetEdition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		editionID values.LauncherVersionID
	}
	type mockInfo struct {
		edition *domain.LauncherVersion

		errGetEdition error
	}

	type test struct {
		description     string
		args            args
		mockInfo        mockInfo
		expectedEdition *domain.LauncherVersion
		isErr           bool
		err             error
	}

	editionID1, edition1 := generateEdition(t, true)
	editionID2, edition2 := generateEdition(t, false)
	editionID3, _ := generateEdition(t, true)
	editionID4, _ := generateEdition(t, true)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			args: args{
				editionID: editionID1,
			},
			mockInfo: mockInfo{
				edition: edition1,
			},
			expectedEdition: edition1,
		},
		{
			description: "URLなしでもエラーなし",
			args: args{
				editionID: editionID2,
			},
			mockInfo: mockInfo{
				edition: edition2,
			},
			expectedEdition: edition2,
		},
		{
			description: "GetEditionでErrRecordNotFoundなのでエラー",
			args: args{
				editionID: editionID3,
			},
			mockInfo: mockInfo{
				errGetEdition: repository.ErrRecordNotFound,
			},
			isErr: true,
			err:   service.ErrInvalidEditionID,
		},
		{
			description: "GetEditionでその他のエラーなのでエラー",
			args: args{
				editionID: editionID4,
			},
			mockInfo: mockInfo{
				errGetEdition: errors.New("error"),
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mockRepository.NewMockDB(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)

			editionService := NewEdition(
				mockDB,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
			)

			mockEditionRepository.
				EXPECT().
				GetEdition(ctx, testCase.args.editionID, repository.LockTypeNone).
				Return(testCase.mockInfo.edition, testCase.mockInfo.errGetEdition)

			got, err := editionService.GetEdition(ctx, testCase.args.editionID)
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

			expected := testCase.expectedEdition

			assert.Equal(t, expected.GetID(), got.GetID())
			assert.Equal(t, expected.GetName(), got.GetName())

			expectedURL, expectedErr := expected.GetQuestionnaireURL()
			gotURL, gotErr := got.GetQuestionnaireURL()
			assert.Equal(t, expectedURL, gotURL)
			assert.Equal(t, expectedErr, gotErr)

			assert.WithinDuration(t, expected.GetCreatedAt(), got.GetCreatedAt(), time.Second*2)
		})
	}
}

func TestUpdateEdition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		editionID        values.LauncherVersionID
		name             values.LauncherVersionName
		questionnaireURL types.Option[values.LauncherVersionQuestionnaireURL]
	}
	type mockInfo struct {
		edition        *domain.LauncherVersion
		updatedEdition *domain.LauncherVersion

		executeUpdateEdition bool

		errGetEdition    error
		errUpdateEdition error
	}

	type test struct {
		description     string
		args            args
		mockInfo        mockInfo
		expectedEdition *domain.LauncherVersion
		isErr           bool
		err             error
	}

	newName := values.NewLauncherVersionName("v2.0.0")
	newURL, err := url.Parse("https://example.com/new")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	newOptionalURL := types.NewOption[values.LauncherVersionQuestionnaireURL](newURL)

	editionID1, edition1 := generateEdition(t, true)
	updatedEdition1 := copyEdition(t, edition1)
	updatedEdition1.SetName(newName)
	updatedEdition1.SetQuestionnaireURL(newURL)

	editionID2, edition2 := generateEdition(t, true)
	updatedEdition2 := copyEdition(t, edition2)
	updatedEdition2.SetQuestionnaireURL(newURL)

	editionID3, edition3 := generateEdition(t, true)
	updatedEdition3 := copyEdition(t, edition3)
	updatedEdition3.SetName(values.NewLauncherVersionName(""))
	updatedEdition3.SetQuestionnaireURL(newURL)

	editionID4, edition4 := generateEdition(t, true)
	updatedEdition4 := copyEdition(t, edition4)
	updatedEdition4.SetName(newName)
	url4, _ := edition4.GetQuestionnaireURL()

	editionID5, edition5 := generateEdition(t, false)
	updatedEdition5 := copyEdition(t, edition5)
	updatedEdition5.SetName(newName)
	updatedEdition5.SetQuestionnaireURL(newURL)

	editionID6, edition6 := generateEdition(t, true)
	updatedEdition6 := copyEdition(t, edition6)
	updatedEdition6.SetName(newName)
	updatedEdition6.UnsetQuestionnaireURL()

	editionID7, _ := generateEdition(t, true)
	editionID8, _ := generateEdition(t, true)

	editionID9, edition9 := generateEdition(t, true)
	updatedEdition9 := copyEdition(t, edition9)
	updatedEdition9.SetName(newName)
	updatedEdition9.SetQuestionnaireURL(newURL)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			args: args{
				editionID:        editionID1,
				name:             newName,
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				edition:        edition1,
				updatedEdition: updatedEdition1,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition1,
		},
		{
			description: "nameが変わらないがエラーなし",
			args: args{
				editionID:        editionID2,
				name:             edition2.GetName(),
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				edition:        edition2,
				updatedEdition: updatedEdition2,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition2,
		},
		{
			description: "nameが空文字列だがエラーなし",
			args: args{
				editionID:        editionID3,
				name:             values.NewLauncherVersionName(""),
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				edition:        edition3,
				updatedEdition: updatedEdition3,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition3,
		},
		{
			description: "URLの変更なしだがエラーなし",
			args: args{
				editionID:        editionID4,
				name:             newName,
				questionnaireURL: types.NewOption(url4),
			},
			mockInfo: mockInfo{
				edition:        edition4,
				updatedEdition: updatedEdition4,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition4,
		},
		{
			description: "URLなし->ありだがエラーなし",
			args: args{
				editionID:        editionID5,
				name:             newName,
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				edition:        edition5,
				updatedEdition: updatedEdition5,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition5,
		},
		{
			description: "URLあり->なしだがエラーなし",
			args: args{
				editionID:        editionID6,
				name:             newName,
				questionnaireURL: types.Option[values.LauncherVersionQuestionnaireURL]{},
			},
			mockInfo: mockInfo{
				edition:        edition6,
				updatedEdition: updatedEdition6,

				executeUpdateEdition: true,
			},
			expectedEdition: updatedEdition6,
		},
		{
			description: "GetEditionでErrRecordNotFoundなのでエラー",
			args: args{
				editionID:        editionID7,
				name:             newName,
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				errGetEdition: repository.ErrRecordNotFound,

				executeUpdateEdition: false,
			},
			isErr: true,
			err:   service.ErrInvalidEditionID,
		},
		{
			description: "GetEditionでその他のエラーなのでエラー",
			args: args{
				editionID:        editionID8,
				name:             newName,
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				errGetEdition: errors.New("error"),

				executeUpdateEdition: false,
			},
			isErr: true,
		},
		{
			description: "UpdateEditionでエラーなのでエラー",
			args: args{
				editionID:        editionID9,
				name:             newName,
				questionnaireURL: newOptionalURL,
			},
			mockInfo: mockInfo{
				edition:        edition9,
				updatedEdition: updatedEdition9,

				executeUpdateEdition: true,

				errUpdateEdition: errors.New("true"),
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mockRepository.NewMockDB(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)

			editionService := NewEdition(
				mockDB,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
			)

			mockEditionRepository.
				EXPECT().
				GetEdition(ctx, testCase.args.editionID, repository.LockTypeRecord).
				Return(testCase.mockInfo.edition, testCase.mockInfo.errGetEdition)

			if testCase.mockInfo.executeUpdateEdition {
				mockEditionRepository.EXPECT().
					UpdateEdition(ctx, testCase.mockInfo.updatedEdition).
					Return(testCase.mockInfo.errUpdateEdition)
			}

			got, err := editionService.UpdateEdition(ctx, testCase.args.editionID, testCase.args.name, testCase.args.questionnaireURL)

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

			expected := testCase.expectedEdition

			assert.Equal(t, expected.GetID(), got.GetID())
			assert.Equal(t, expected.GetName(), got.GetName())

			expectedURL, expectedErr := expected.GetQuestionnaireURL()
			gotURL, gotErr := got.GetQuestionnaireURL()
			assert.Equal(t, expectedURL, gotURL)
			assert.Equal(t, expectedErr, gotErr)

			assert.WithinDuration(t, expected.GetCreatedAt(), got.GetCreatedAt(), time.Second*2)
		})
	}
}

func generateGameVersionsForEditionTests(t *testing.T, count int) ([]values.GameVersionID, []*repository.GameVersionInfoWithGameID) {
	t.Helper()

	gameVersionIDs := make([]values.GameVersionID, 0, count)
	gameVersions := make([]*repository.GameVersionInfoWithGameID, 0, count)

	for i := 0; i < count; i++ {
		gameVersionID := values.NewGameVersionID()
		gameVersionIDs = append(gameVersionIDs, gameVersionID)

		gameVersion := &repository.GameVersionInfoWithGameID{
			GameVersion: domain.NewGameVersion(gameVersionID, values.NewGameVersionName("v1.1.0"), values.NewGameVersionDescription("test"), time.Now()),
			GameID:      values.NewGameID(),
			ImageID:     values.NewGameImageID(),
			VideoID:     values.NewGameVideoID(),
			FileIDs:     []values.GameFileID{values.NewGameFileID()},
			URL:         types.Option[values.GameURLLink]{},
		}
		gameVersions = append(gameVersions, gameVersion)
	}

	return gameVersionIDs, gameVersions
}

func generateEdition(t *testing.T, haveQuestionnaire bool) (editionID values.LauncherVersionID, edition *domain.LauncherVersion) {
	t.Helper()

	editionID = values.NewLauncherVersionID()

	if haveQuestionnaire {
		urlStr := "https://example.com"
		urlLink, err := url.Parse(urlStr)
		if err != nil {
			t.Fatalf("failed to parse url: %v", err)
		}

		edition = domain.NewLauncherVersionWithQuestionnaire(editionID, values.NewLauncherVersionName("v1.0.0"), values.NewLauncherVersionQuestionnaireURL(urlLink), time.Now())
	} else {
		edition = domain.NewLauncherVersionWithoutQuestionnaire(editionID, values.NewLauncherVersionName("v1.0.0"), time.Now())
	}

	return editionID, edition
}

func generateEditions(t *testing.T, haveQuestionnaire bool, count int) []*domain.LauncherVersion {
	t.Helper()

	editions := make([]*domain.LauncherVersion, 0, count)
	for i := 0; i < count; i++ {
		_, edition := generateEdition(t, haveQuestionnaire)
		editions = append(editions, edition)
	}

	return editions
}

// UpdateEdition用
// 主にUpdateEdition内で変更される要素を対象にしたhard copyをする
func copyEdition(t *testing.T, base *domain.LauncherVersion) *domain.LauncherVersion {
	t.Helper()

	var newEdition *domain.LauncherVersion

	name := values.NewLauncherVersionName(strings.Clone(string(base.GetName())))

	// UpdateEdition内での変更対象ではないのでhard copyでなくてよい
	editionID := base.GetID()
	createdAt := base.GetCreatedAt()

	questionnaireURL, _ := base.GetQuestionnaireURL()
	if questionnaireURL == nil {
		newEdition = domain.NewLauncherVersionWithoutQuestionnaire(editionID, name, createdAt)
	} else {
		urlStr := (*url.URL)(questionnaireURL).String()
		urlLink, err := url.Parse(urlStr)
		if err != nil {
			t.Fatalf("failed to parse url: %v", err)
		}

		newEdition = domain.NewLauncherVersionWithQuestionnaire(editionID, name, values.NewLauncherVersionQuestionnaireURL(urlLink), createdAt)
	}

	return newEdition
}
