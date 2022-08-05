package values

type (
	GameManagementRole int
)

const (
	// GameManagerRoleAdministrator
	// ゲームの管理者。
	// ゲーム更新とCollaborationの追加権限がある。
	GameManagementRoleAdministrator GameManagementRole = iota
	// GameManagerRoleCollaborator
	// ゲームの共同管理者。
	// ゲーム更新ができる。
	GameManagementRoleCollaborator
)

func (gmr GameManagementRole) HaveGameUpdatePermission() bool {
	return gmr == GameManagementRoleAdministrator || gmr == GameManagementRoleCollaborator
}

func (gmr GameManagementRole) HaveUpdateManagementRolePermission() bool {
	return gmr == GameManagementRoleAdministrator
}
