package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLauncherEnvironmentAcceptGameFileTypes(t *testing.T) {
	t.Parallel()

	type test struct {
		description     string
		environment     *LauncherEnvironment
		acceptFileTypes []GameFileType
	}

	testCases := []test{
		{
			description: "windowsの場合、jarとwindowsを許可する",
			environment: &LauncherEnvironment{
				os: LauncherEnvironmentOSWindows,
			},
			acceptFileTypes: []GameFileType{
				GameFileTypeJar,
				GameFileTypeWindows,
			},
		},
		{
			description: "macの場合、jarとmacを許可する",
			environment: &LauncherEnvironment{
				os: LauncherEnvironmentOSMac,
			},
			acceptFileTypes: []GameFileType{
				GameFileTypeJar,
				GameFileTypeMac,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			acceptFileTypes := testCase.environment.AcceptGameFileTypes()

			assert.Len(t, acceptFileTypes, len(testCase.acceptFileTypes))
			assert.Subset(t, acceptFileTypes, testCase.acceptFileTypes)
		})
	}
}
