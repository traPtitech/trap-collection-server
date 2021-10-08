package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

func TestAuthorize(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	oidcService := NewOIDC(mockOIDCAuth, common.ClientID("clientID"))

	client, session, err := oidcService.Authorize(ctx)
	assert.NoError(t, err)

	assert.Equal(t, oidcService.client, client)
	assert.Equal(t, values.OIDCCodeChallengeMethodSha256, session.GetCodeChallengeMethod())
}

func TestCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	oidcService := NewOIDC(mockOIDCAuth, common.ClientID("clientID"))

	type test struct {
		description       string
		GetOIDCSessionErr error
		isErr             bool
		err               error
	}

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
		},
		{
			description:       "GetOIDCSessionでエラー",
			GetOIDCSessionErr: errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			codeVerifier, err := values.NewOIDCCodeVerifier()
			assert.NoError(t, err)

			code := values.NewOIDCAuthorizationCode("")
			authState := domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			)
			var session *domain.OIDCSession
			if testCase.GetOIDCSessionErr == nil {
				session = domain.NewOIDCSession(
					values.NewAccessToken("access token"),
					time.Now(),
				)
			}
			mockOIDCAuth.
				EXPECT().
				GetOIDCSession(ctx, oidcService.client, code, authState).
				Return(session, testCase.GetOIDCSessionErr)

			actualSession, err := oidcService.Callback(ctx, authState, code)

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

			assert.Equal(t, session, actualSession)
		})
	}
}
