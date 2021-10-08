package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

type TraPMember struct {
	id     values.TraPMemberID
	name   values.TraPMemberName
	status values.TraPMemberStatus
	role   values.TraPMemberRole
}

func NewTraPMember(
	id values.TraPMemberID,
	name values.TraPMemberName,
	status values.TraPMemberStatus,
	role values.TraPMemberRole,
) *TraPMember {
	return &TraPMember{
		id:     id,
		name:   name,
		status: status,
		role:   role,
	}
}

func (tm *TraPMember) GetID() values.TraPMemberID {
	return tm.id
}

func (tm *TraPMember) GetName() values.TraPMemberName {
	return tm.name
}

func (tm *TraPMember) GetStatus() values.TraPMemberStatus {
	return tm.status
}

func (tm *TraPMember) SetStatus(status values.TraPMemberStatus) {
	tm.status = status
}

func (tm *TraPMember) GetRole() values.TraPMemberRole {
	return tm.role
}
