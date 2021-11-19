package values

type LauncherEnvironment struct {
	os LauncherEnvironmentOS
}

func NewLauncherEnvironment(os LauncherEnvironmentOS) *LauncherEnvironment {
	return &LauncherEnvironment{
		os: os,
	}
}

func (le *LauncherEnvironment) AcceptGameFileTypes() []GameFileType {
	switch le.os {
	case LauncherEnvironmentOSWindows:
		return []GameFileType{
			GameFileTypeJar,
			GameFileTypeWindows,
		}
	case LauncherEnvironmentOSMac:
		return []GameFileType{
			GameFileTypeJar,
			GameFileTypeMac,
		}
	}

	return nil
}

type (
	LauncherEnvironmentOS int
)

const (
	LauncherEnvironmentOSWindows LauncherEnvironmentOS = iota
	// LauncherEnvironmentOSMac 今のところ稼働させる予定なし
	LauncherEnvironmentOSMac
)
