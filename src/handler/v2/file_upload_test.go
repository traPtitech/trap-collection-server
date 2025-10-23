package v2

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isFileUploadRequest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		method string
		path   string
		gameID string
		want   bool
	}{
		"POST /api/v2/games/:gameID/files": {
			method: http.MethodPost,
			path:   "/api/v2/games/123/files",
			gameID: "123",
			want:   true,
		},
		"POST /api/v2/games/:gameID/images": {
			method: http.MethodPost,
			path:   "/api/v2/games/123/images",
			gameID: "123",
			want:   true,
		},
		"POST /api/v2/games/:gameID/videos": {
			method: http.MethodPost,
			path:   "/api/v2/games/123/videos",
			gameID: "123",
			want:   true,
		},
		"GET /api/v2/games/:gameID/files": {
			method: http.MethodGet,
			path:   "/api/v2/games/123/files",
			gameID: "123",
			want:   false,
		},
		"POST /api/v2/games/:gameID/others": {
			method: http.MethodPost,
			path:   "/api/v2/games/123/others",
			gameID: "123",
			want:   false,
		},
		"POST /api/v2/games/files": {
			method: http.MethodPost,
			path:   "/api/v2/games/files",
			gameID: "",
			want:   false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, _, _ := setupTestRequest(t, testCase.method, testCase.path, nil)
			c.SetParamNames("gameID")
			c.SetParamValues(testCase.gameID)

			got := isFileUploadRequest(c)

			assert.Equal(t, testCase.want, got)
		})
	}
}
