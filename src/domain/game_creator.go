package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameCreator struct {
	id        values.GameCreatorID
	userID    values.TraPMemberID
	userName  values.TraPMemberName
	createdAt time.Time
}

func NewGameCreator(id values.GameCreatorID, userID values.TraPMemberID, userName values.TraPMemberName, createdAt time.Time) *GameCreator {
	return &GameCreator{
		id:        id,
		userID:    userID,
		userName:  userName,
		createdAt: createdAt,
	}
}

func (creator *GameCreator) ID() values.GameCreatorID        { return creator.id }
func (creator *GameCreator) UserID() values.TraPMemberID     { return creator.userID }
func (creator *GameCreator) UserName() values.TraPMemberName { return creator.userName }
func (creator *GameCreator) CreatedAt() time.Time            { return creator.createdAt }

type GameCreatorJob struct {
	id          values.GameCreatorJobID
	displayName values.GameCreatorJobDisplayName
	createdAt   time.Time
}

func NewGameCreatorJob(id values.GameCreatorJobID, displayName values.GameCreatorJobDisplayName, createdAt time.Time) *GameCreatorJob {
	return &GameCreatorJob{
		id:          id,
		displayName: displayName,
		createdAt:   createdAt,
	}
}

func (job *GameCreatorJob) ID() values.GameCreatorJobID                   { return job.id }
func (job *GameCreatorJob) DisplayName() values.GameCreatorJobDisplayName { return job.displayName }
func (job *GameCreatorJob) CreatedAt() time.Time                          { return job.createdAt }

type GameCreatorCustomJob struct {
	id          values.GameCreatorJobID
	displayName values.GameCreatorJobDisplayName
	gameID      values.GameID
	createdAt   time.Time
}

func NewGameCreatorCustomJob(id values.GameCreatorJobID, displayName values.GameCreatorJobDisplayName, gameID values.GameID, createdAt time.Time) *GameCreatorCustomJob {
	return &GameCreatorCustomJob{
		id:          id,
		displayName: displayName,
		gameID:      gameID,
		createdAt:   createdAt,
	}
}

func (customJob *GameCreatorCustomJob) ID() values.GameCreatorJobID { return customJob.id }
func (customJob *GameCreatorCustomJob) DisplayName() values.GameCreatorJobDisplayName {
	return customJob.displayName
}
func (customJob *GameCreatorCustomJob) GameID() values.GameID { return customJob.gameID }
func (customJob *GameCreatorCustomJob) CreatedAt() time.Time  { return customJob.createdAt }

type GameCreatorWithJobs struct {
	gameCreator *GameCreator
	jobs        []*GameCreatorJob
	customJobs  []*GameCreatorCustomJob
}

func NewGameCreatorWithJobs(gameCreator *GameCreator, jobs []*GameCreatorJob, customJobs []*GameCreatorCustomJob) *GameCreatorWithJobs {
	return &GameCreatorWithJobs{
		gameCreator: gameCreator,
		jobs:        jobs,
		customJobs:  customJobs,
	}
}

func (gcj *GameCreatorWithJobs) GameCreator() *GameCreator           { return gcj.gameCreator }
func (gcj *GameCreatorWithJobs) Jobs() []*GameCreatorJob             { return gcj.jobs }
func (gcj *GameCreatorWithJobs) CustomJobs() []*GameCreatorCustomJob { return gcj.customJobs }
