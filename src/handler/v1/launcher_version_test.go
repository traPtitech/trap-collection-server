package v1

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetVersions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherVersionService := mock.NewMockLauncherVersion(ctrl)

	launcherVersionHandler := NewLauncherVersion(mockLauncherVersionService)

	type test struct {
		description            string
		launcherVersions       []*domain.LauncherVersion
		GetLauncherVersionsErr error
		expect                 []*openapi.Version
		isErr                  bool
		err                    error
		statusCode             int
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
			},
		},
		{
			description:            "GetLauncherVersionsがエラーなので500",
			GetLauncherVersionsErr: errors.New("GetLauncherVersions error"),
			isErr:                  true,
			statusCode:             http.StatusInternalServerError,
		},
		{
			description: "Questionnaireありのランチャーバージョンでもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "https://example.com",
					CreatedAt: now,
				},
			},
		},
		{
			description:      "ランチャーバージョンがなくてもエラーなし",
			launcherVersions: []*domain.LauncherVersion{},
			expect:           []*openapi.Version{},
		},
		{
			description: "ランチャーバージョンが複数でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID2,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(launcherVersionID2).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionService.
				EXPECT().
				GetLauncherVersions(gomock.Any()).
				Return(testCase.launcherVersions, testCase.GetLauncherVersionsErr)

			launcherVersions, err := launcherVersionHandler.GetVersions()

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Len(t, launcherVersions, len(testCase.expect))

			for i, expect := range testCase.expect {
				assert.Equal(t, *expect, *launcherVersions[i])
			}
		})
	}
}
