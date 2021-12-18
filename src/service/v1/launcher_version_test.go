package v1

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
)

func TestCreateLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description              string
		name                     values.LauncherVersionName
		questionnaireURL         values.LauncherVersionQuestionnaireURL
		CreateLauncherVersionErr error
		isErr                    bool
		err                      error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description:      "questionnaireURLありでエラーなし",
			name:             values.NewLauncherVersionName("name"),
			questionnaireURL: values.NewLauncherVersionQuestionnaireURL(urlLink),
		},
		{
			description: "questionnaireURLなしでエラーなし",
			name:        values.NewLauncherVersionName("name"),
		},
		{
			description:              "CreateLauncherVersionがエラーなのでエラー",
			name:                     values.NewLauncherVersionName("name"),
			questionnaireURL:         values.NewLauncherVersionQuestionnaireURL(urlLink),
			CreateLauncherVersionErr: errors.New("error"),
			isErr:                    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				CreateLauncherVersion(ctx, gomock.Any()).
				Return(testCase.CreateLauncherVersionErr)

			launcherVersion, err := launcherVersionService.CreateLauncherVersion(ctx, testCase.name, testCase.questionnaireURL)

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

			assert.Equal(t, testCase.name, launcherVersion.GetName())
			assert.WithinDuration(t, time.Now(), launcherVersion.GetCreatedAt(), 2*time.Second)

			questionnaireURL, err := launcherVersion.GetQuestionnaireURL()

			if testCase.questionnaireURL == nil {
				assert.True(t, errors.Is(err, domain.ErrNoQuestionnaire))
			} else {
				assert.Equal(t, testCase.questionnaireURL, questionnaireURL)
			}
		})
	}
}

func TestGetLauncherVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description            string
		launcherVersions       []*domain.LauncherVersion
		GetLauncherVersionsErr error
		isErr                  bool
		err                    error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
			},
		},
		{
			description:            "GetLauncherVersionsがエラーなのでエラー",
			GetLauncherVersionsErr: errors.New("error"),
			isErr:                  true,
		},
		{
			description:      "launcherVersionsが空でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{},
		},
		{
			description: "launcherVersionsが複数でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersions(ctx).
				Return(testCase.launcherVersions, testCase.GetLauncherVersionsErr)

			launcherVersions, err := launcherVersionService.GetLauncherVersions(ctx)

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

			assert.Len(t, launcherVersions, len(testCase.launcherVersions))

			for i, launcherVersion := range launcherVersions {
				assert.Equal(t, testCase.launcherVersions[i].GetName(), launcherVersion.GetName())
				assert.WithinDuration(t, testCase.launcherVersions[i].GetCreatedAt(), launcherVersion.GetCreatedAt(), 2*time.Second)

				questionnaireURL, err := launcherVersion.GetQuestionnaireURL()

				if errors.Is(err, domain.ErrNoQuestionnaire) {
					_, err = testCase.launcherVersions[i].GetQuestionnaireURL()
					assert.True(t, errors.Is(err, domain.ErrNoQuestionnaire))
				} else {
					expectQuestionnaireURL, err := testCase.launcherVersions[i].GetQuestionnaireURL()
					assert.False(t, errors.Is(err, domain.ErrNoQuestionnaire))
					assert.Equal(t, expectQuestionnaireURL, questionnaireURL)
				}
			}
		})
	}
}
