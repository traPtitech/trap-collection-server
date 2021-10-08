package v1

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth/mock"
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
