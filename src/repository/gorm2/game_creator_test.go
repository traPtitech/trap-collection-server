package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestGetGameCreatorsByGameID(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID1 := values.NewGameID()
	game1 := &schema.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "Test Game 1",
		VisibilityTypeID: 1,
	}
	gameID2 := values.NewGameID()
	game2 := &schema.GameTable2{
		ID:               uuid.UUID(gameID2),
		Name:             "Test Game 2",
		VisibilityTypeID: 1,
	}
	err = db.Create([]*schema.GameTable2{game1, game2}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	job1 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "job1",
	}
	job2 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "job2",
	}
	err = db.Create([]*schema.GameCreatorJobTable{job1, job2}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameCreatorJobTable{}).Error
		require.NoError(t, err)
	})

	customJob1 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob1",
		GameID:      uuid.UUID(gameID1),
	}
	customJob2 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob2",
		GameID:      uuid.UUID(gameID2),
	}
	customJob3 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob3",
		GameID:      uuid.UUID(gameID2),
	}
	err = db.Create([]*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameCreatorCustomJobTable{}).Error
		require.NoError(t, err)
	})

	creator1 := &schema.GameCreatorTable{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		UserName: "creator1",
		GameID:   uuid.UUID(gameID1),
		CreatorJobs: []schema.GameCreatorJobTable{
			*job1,
		},
		CustomCreatorJobs: []schema.GameCreatorCustomJobTable{
			*customJob1,
		},
		CreatedAt: now,
	}
	creator2 := &schema.GameCreatorTable{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		UserName: "creator2",
		GameID:   uuid.UUID(gameID2),
		CreatorJobs: []schema.GameCreatorJobTable{
			*job2,
		},
		CustomCreatorJobs: []schema.GameCreatorCustomJobTable{
			*customJob2,
			*customJob3,
		},
		CreatedAt: now,
	}
	creator3 := &schema.GameCreatorTable{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		UserName:  "creator3",
		GameID:    uuid.UUID(gameID2),
		CreatedAt: now.Add(1 * time.Hour),
	}

	testCases := map[string]struct {
		gameID      values.GameID
		allCreators []*schema.GameCreatorTable
		expected    []*domain.GameCreatorWithJobs
		err         error
	}{
		"一人を正しく取得できる": {
			gameID:      gameID1,
			allCreators: []*schema.GameCreatorTable{creator1, creator2},
			expected: []*domain.GameCreatorWithJobs{
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator1.ID),
						values.TraPMemberID(creator1.UserID),
						values.NewGameIDFromUUID(creator1.GameID),
						values.TraPMemberName(creator1.UserName),
						creator1.CreatedAt),
					[]*domain.GameCreatorJob{
						domain.NewGameCreatorJob(
							values.GameCreatorJobID(job1.ID),
							values.GameCreatorJobDisplayName(job1.DisplayName),
							job1.CreatedAt,
						),
					},
					[]*domain.GameCreatorCustomJob{
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob1.ID),
							values.GameCreatorJobDisplayName(customJob1.DisplayName),
							values.GameID(customJob1.GameID),
							customJob1.CreatedAt,
						),
					},
				),
			},
		},
		"複数人を正しく取得できる": {
			gameID:      gameID2,
			allCreators: []*schema.GameCreatorTable{creator1, creator2, creator3},
			expected: []*domain.GameCreatorWithJobs{
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator2.ID),
						values.TraPMemberID(creator2.UserID),
						values.NewGameIDFromUUID(creator2.GameID),
						values.TraPMemberName(creator2.UserName),
						creator2.CreatedAt),
					[]*domain.GameCreatorJob{
						domain.NewGameCreatorJob(
							values.GameCreatorJobID(job2.ID),
							values.GameCreatorJobDisplayName(job2.DisplayName),
							job2.CreatedAt,
						),
					},
					[]*domain.GameCreatorCustomJob{
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob2.ID),
							values.GameCreatorJobDisplayName(customJob2.DisplayName),
							values.GameID(customJob2.GameID),
							customJob2.CreatedAt,
						),
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob3.ID),
							values.GameCreatorJobDisplayName(customJob3.DisplayName),
							values.GameID(customJob3.GameID),
							customJob3.CreatedAt,
						),
					},
				),
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator3.ID),
						values.TraPMemberID(creator3.UserID),
						values.NewGameIDFromUUID(creator3.GameID),
						values.TraPMemberName(creator3.UserName),
						creator3.CreatedAt),
					[]*domain.GameCreatorJob{},
					[]*domain.GameCreatorCustomJob{},
				),
			},
		},
		"該当するゲームクリエイターがいない場合は空配列を返す": {
			gameID:      values.NewGameID(),
			allCreators: []*schema.GameCreatorTable{creator1, creator2},
			expected:    []*domain.GameCreatorWithJobs{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			err = db.Create(testCase.allCreators).Error
			require.NoError(t, err)
			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Select(clause.Associations).Delete(testCase.allCreators).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			creatorsWithJobs, err := repo.GetGameCreatorsByGameID(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)

			assert.Len(t, creatorsWithJobs, len(testCase.expected))

			for i, creatorWithJobs := range creatorsWithJobs {
				creator := creatorWithJobs.GetGameCreator()
				expectedCreator := testCase.expected[i].GetGameCreator()
				assert.Equal(t, expectedCreator.GetID(), creator.GetID())
				assert.Equal(t, expectedCreator.GetUserID(), creator.GetUserID())
				assert.Equal(t, expectedCreator.GetGameID(), creator.GetGameID())
				assert.Equal(t, expectedCreator.GetUserName(), creator.GetUserName())
				assert.WithinDuration(t, expectedCreator.GetCreatedAt(), creator.GetCreatedAt(), time.Second)

				actualJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(creatorWithJobs.GetJobs()))
				for _, job := range creatorWithJobs.GetJobs() {
					actualJobsMap[job.GetID()] = job
				}
				assert.Len(t, actualJobsMap, len(testCase.expected[i].GetJobs()))
				for _, expectedJob := range testCase.expected[i].GetJobs() {
					actualJob, ok := actualJobsMap[expectedJob.GetID()]
					assert.Truef(t, ok, "expected job ID %v not found in actual jobs", expectedJob.GetID())
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}

				actualCustomJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorCustomJob, len(creatorWithJobs.GetCustomJobs()))
				for _, job := range creatorWithJobs.GetCustomJobs() {
					actualCustomJobsMap[job.GetID()] = job
				}
				assert.Len(t, actualCustomJobsMap, len(testCase.expected[i].GetCustomJobs()))
				for _, expectedJob := range testCase.expected[i].GetCustomJobs() {
					actualJob, ok := actualCustomJobsMap[expectedJob.GetID()]
					assert.Truef(t, ok, "expected custom job ID %v not found in actual custom jobs", expectedJob.GetID())
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.Equal(t, expectedJob.GetGameID(), actualJob.GetGameID())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}

func TestGetGameCreatorPresetJobs(t *testing.T) {
	now := time.Now()
	job1 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "プログラマー",
		CreatedAt:   now,
	}
	job2 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "デザイナー",
		CreatedAt:   now,
	}

	testCases := map[string]struct {
		presetJobs []*schema.GameCreatorJobTable
		expected   []*domain.GameCreatorJob
		err        error
	}{
		"プリセットジョブが存在する場合は取得できる": {
			presetJobs: []*schema.GameCreatorJobTable{job1, job2},
			expected: []*domain.GameCreatorJob{
				domain.NewGameCreatorJob(
					values.GameCreatorJobID(job1.ID),
					values.GameCreatorJobDisplayName(job1.DisplayName),
					job1.CreatedAt,
				),
				domain.NewGameCreatorJob(
					values.GameCreatorJobID(job2.ID),
					values.GameCreatorJobDisplayName(job2.DisplayName),
					job2.CreatedAt,
				),
			},
		},
		"プリセットジョブが存在しない場合は空配列を返す": {
			presetJobs: []*schema.GameCreatorJobTable{},
			expected:   []*domain.GameCreatorJob{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.presetJobs) > 0 {
				err = db.Create(testCase.presetJobs).Error
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(testCase.presetJobs).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			jobs, err := repo.GetGameCreatorPresetJobs(t.Context())

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, jobs, len(testCase.expected))

			expectedMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(testCase.expected))
			for _, job := range testCase.expected {
				expectedMap[job.GetID()] = job
			}

			for _, job := range jobs {
				expected, ok := expectedMap[job.GetID()]
				if !assert.Truef(t, ok, "unexpected job ID: %v", job.GetID()) {
					continue
				}
				assert.Equal(t, expected.GetDisplayName(), job.GetDisplayName())
				assert.WithinDuration(t, expected.GetCreatedAt(), job.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetGameCreatorCustomJobsByGameID(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID1 := values.NewGameID()
	game1 := &schema.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "test game 1",
		VisibilityTypeID: 1,
	}
	gameID2 := values.NewGameID()
	game2 := &schema.GameTable2{
		ID:               uuid.UUID(gameID2),
		Name:             "test game 2",
		VisibilityTypeID: 1,
	}
	err = db.Create([]*schema.GameTable2{game1, game2}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	customJob1 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "ボーカル",
		GameID:      uuid.UUID(gameID1),
		CreatedAt:   now,
	}
	customJob2 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "作曲者",
		GameID:      uuid.UUID(gameID2),
		CreatedAt:   now,
	}
	customJob3 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "イラストレーター",
		GameID:      uuid.UUID(gameID2),
		CreatedAt:   now,
	}

	testCases := map[string]struct {
		gameID     values.GameID
		customJobs []*schema.GameCreatorCustomJobTable
		expected   []*domain.GameCreatorCustomJob
		err        error
	}{
		"対象ゲームのカスタムジョブを1件取得できる": {
			gameID:     gameID1,
			customJobs: []*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3},
			expected: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					values.GameCreatorJobID(customJob1.ID),
					values.GameCreatorJobDisplayName(customJob1.DisplayName),
					values.GameID(customJob1.GameID),
					customJob1.CreatedAt,
				),
			},
		},
		"対象ゲームのカスタムジョブを複数件取得できる": {
			gameID:     gameID2,
			customJobs: []*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3},
			expected: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					values.GameCreatorJobID(customJob2.ID),
					values.GameCreatorJobDisplayName(customJob2.DisplayName),
					values.GameID(customJob2.GameID),
					customJob2.CreatedAt,
				),
				domain.NewGameCreatorCustomJob(
					values.GameCreatorJobID(customJob3.ID),
					values.GameCreatorJobDisplayName(customJob3.DisplayName),
					values.GameID(customJob3.GameID),
					customJob3.CreatedAt,
				),
			},
		},
		"対象ゲームにカスタムジョブが存在しない場合は空配列": {
			gameID:     gameID1,
			customJobs: []*schema.GameCreatorCustomJobTable{},
			expected:   []*domain.GameCreatorCustomJob{},
		},
		"gameIDが存在しない場合は空配列": {
			gameID:     values.NewGameID(),
			customJobs: []*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3},
			expected:   []*domain.GameCreatorCustomJob{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.customJobs) > 0 {
				err = db.Create(testCase.customJobs).Error
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(testCase.customJobs).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			customJobs, err := repo.GetGameCreatorCustomJobsByGameID(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, customJobs, len(testCase.expected))

			expectedMap := make(map[values.GameCreatorJobID]*domain.GameCreatorCustomJob, len(testCase.expected))
			for _, job := range testCase.expected {
				expectedMap[job.GetID()] = job
			}

			for _, job := range customJobs {
				expected, ok := expectedMap[job.GetID()]
				if !assert.Truef(t, ok, "unexpected custom job ID: %v", job.GetID()) {
					continue
				}
				assert.Equal(t, expected.GetDisplayName(), job.GetDisplayName())
				assert.Equal(t, expected.GetGameID(), job.GetGameID())
				assert.WithinDuration(t, expected.GetCreatedAt(), job.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestCreateGameCreatorCustomJobs(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	invalidGameID := values.NewGameID()

	games := []*schema.GameTable2{
		{
			ID:               uuid.UUID(gameID1),
			Name:             "クロノ・トリガーtest",
			VisibilityTypeID: 1,
		},
		{
			ID:               uuid.UUID(gameID2),
			Name:             "ポケモソ不思議のダンジョンtest",
			VisibilityTypeID: 1,
		},
	}
	err = db.Create(games).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("id IN ?", []uuid.UUID{uuid.UUID(gameID1), uuid.UUID(gameID2)}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	jobID1 := values.NewGameCreatorJobID()
	jobID2 := values.NewGameCreatorJobID()
	jobIDFKFail := values.NewGameCreatorJobID()
	dupJobID := values.NewGameCreatorJobID()
	dupJobIDExtra := values.NewGameCreatorJobID()
	mixValidJobID := values.NewGameCreatorJobID()
	mixInvalidJobID := values.NewGameCreatorJobID()

	testCases := map[string]struct {
		addJobs         []*domain.GameCreatorCustomJob
		beforeJobs      []*schema.GameCreatorCustomJobTable
		wantDelta       int
		wantErr         error
		wantInsertedIDs []values.GameCreatorJobID
	}{
		"複数件のカスタムジョブを作成できる": {
			addJobs: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					jobID1,
					values.GameCreatorJobDisplayName("カスタムジョブご飯"),
					gameID1,
					now,
				),
				domain.NewGameCreatorCustomJob(
					jobID2,
					values.GameCreatorJobDisplayName("カスタムジョブおにぎり"),
					gameID2,
					now,
				),
			},
			beforeJobs: []*schema.GameCreatorCustomJobTable{
				{
					ID:          uuid.UUID(values.NewGameCreatorJobID()),
					DisplayName: "既存のカスタムジョブ",
					GameID:      uuid.UUID(gameID1),
					CreatedAt:   now.Add(-1 * time.Hour),
				},
			},
			wantDelta:       2,
			wantErr:         nil,
			wantInsertedIDs: []values.GameCreatorJobID{jobID1, jobID2},
		},
		"空配列を渡してもエラーにならない": {
			addJobs:    []*domain.GameCreatorCustomJob{},
			beforeJobs: []*schema.GameCreatorCustomJobTable{},
			wantDelta:  0,
			wantErr:    nil,
		},
		"nil入力でエラーにならない": {
			addJobs:    nil,
			beforeJobs: []*schema.GameCreatorCustomJobTable{},
			wantDelta:  0,
			wantErr:    nil,
		},
		"存在しないgame_idのジョブを作成するとエラーになる": {
			addJobs: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					jobIDFKFail,
					values.GameCreatorJobDisplayName("カスタムジョブ無効"),
					invalidGameID,
					now,
				),
			},
			beforeJobs:      []*schema.GameCreatorCustomJobTable{},
			wantDelta:       0,
			wantErr:         repository.ErrForeignKeyViolated,
			wantInsertedIDs: []values.GameCreatorJobID{},
		},
		"重複するIDのジョブを作成するとエラーになる": {
			addJobs: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					dupJobID,
					values.GameCreatorJobDisplayName("カスタムジョブご飯"),
					gameID1,
					now,
				),
				domain.NewGameCreatorCustomJob(
					dupJobID,
					values.GameCreatorJobDisplayName("カスタムジョブご飯(重複)"),
					gameID1,
					now,
				),
				domain.NewGameCreatorCustomJob(
					dupJobIDExtra,
					values.GameCreatorJobDisplayName("カスタムジョブおにぎり"),
					gameID2,
					now,
				),
			},
			beforeJobs:      []*schema.GameCreatorCustomJobTable{},
			wantDelta:       0,
			wantErr:         repository.ErrDuplicatedUniqueKey,
			wantInsertedIDs: []values.GameCreatorJobID{},
		},
		"有効ジョブと無効ジョブを同時に投入するとロールバック": {
			addJobs: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					mixValidJobID,
					values.GameCreatorJobDisplayName("カスタムジョブご飯"),
					gameID1,
					now,
				),
				domain.NewGameCreatorCustomJob(
					mixInvalidJobID,
					values.GameCreatorJobDisplayName("カスタムジョブ無効"),
					invalidGameID,
					now,
				),
			},
			beforeJobs:      []*schema.GameCreatorCustomJobTable{},
			wantDelta:       0,
			wantErr:         repository.ErrForeignKeyViolated,
			wantInsertedIDs: []values.GameCreatorJobID{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.beforeJobs) > 0 {
				err = db.Create(testCase.beforeJobs).Error
				require.NoError(t, err)
			}
			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("game_id IN ?", []uuid.UUID{uuid.UUID(gameID1), uuid.UUID(gameID2)}).Delete(&schema.GameCreatorCustomJobTable{}).Error
				require.NoError(t, err)
			})

			beforeCount := int64(len(testCase.beforeJobs))

			repo := NewGameCreator(testDB)
			err = repo.CreateGameCreatorCustomJobs(t.Context(), testCase.addJobs)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
			}

			var afterCount int64
			err = db.Model(&schema.GameCreatorCustomJobTable{}).
				Where("game_id IN ?", []uuid.UUID{uuid.UUID(gameID1), uuid.UUID(gameID2)}).
				Count(&afterCount).Error
			require.NoError(t, err)
			assert.Equal(t, beforeCount+int64(testCase.wantDelta), afterCount)

			if len(testCase.wantInsertedIDs) == 0 {
				return
			}
			wantIDs := make([]uuid.UUID, 0, len(testCase.wantInsertedIDs))
			for _, id := range testCase.wantInsertedIDs {
				wantIDs = append(wantIDs, uuid.UUID(id))
			}

			var insertedJobs []schema.GameCreatorCustomJobTable
			err = db.Where("id IN ?", wantIDs).Find(&insertedJobs).Error
			require.NoError(t, err)
			assert.Len(t, insertedJobs, len(testCase.wantInsertedIDs))

			expectedJobsMap := make(map[uuid.UUID]*domain.GameCreatorCustomJob, len(testCase.addJobs))
			for _, job := range testCase.addJobs {
				expectedJobsMap[uuid.UUID(job.GetID())] = job
			}
			for _, job := range insertedJobs {
				expected, ok := expectedJobsMap[job.ID]
				if !assert.Truef(t, ok, "unexpected ID inserted: %v", job.ID) {
					continue
				}
				assert.Equal(t, uuid.UUID(expected.GetID()), job.ID)
				assert.Equal(t, uuid.UUID(expected.GetGameID()), job.GameID)
				assert.Equal(t, string(expected.GetDisplayName()), job.DisplayName)
				assert.WithinDuration(t, expected.GetCreatedAt(), job.CreatedAt, time.Second)
			}
		})
	}
}

func TestCreateGameCreators(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID := values.NewGameID()
	game := &schema.GameTable2{
		ID:               uuid.UUID(gameID),
		Name:             "Test Game",
		VisibilityTypeID: 1,
	}
	err = db.Create(game).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	creator1 := domain.NewGameCreator(
		values.NewGameCreatorID(),
		values.NewTrapMemberID(uuid.New()),
		gameID,
		values.NewTrapMemberName("えびせん"),
		now,
	)
	creator2 := domain.NewGameCreator(
		values.NewGameCreatorID(),
		values.NewTrapMemberID(uuid.New()),
		gameID,
		values.NewTrapMemberName("ぽてと"),
		now.Add(1*time.Hour),
	)

	testCases := map[string]struct {
		addCreators       []*domain.GameCreator
		beforeCreators    []*schema.GameCreatorTable
		wantDeltaCreators int
		err               error
	}{
		"複数のゲームクリエイターを作成できる": {
			addCreators:       []*domain.GameCreator{creator1, creator2},
			beforeCreators:    []*schema.GameCreatorTable{},
			wantDeltaCreators: 2,
			err:               nil,
		},
		"空の配列を渡した場合は何もしない": {
			addCreators:       []*domain.GameCreator{},
			beforeCreators:    []*schema.GameCreatorTable{},
			wantDeltaCreators: 0,
			err:               nil,
		},
		"存在しないゲームIDのゲームクリエイターを作成しようとした場合はエラー": {
			addCreators: []*domain.GameCreator{
				domain.NewGameCreator(
					values.NewGameCreatorID(),
					values.NewTrapMemberID(uuid.New()),
					values.NewGameID(),
					values.NewTrapMemberName("存在しないゲームのクリエイター"),
					now,
				),
			},
			wantDeltaCreators: 0,
			err:               repository.ErrForeignKeyViolated,
		},
		"同一 gameID と userID の組み合わせで重複エラーが発生する": {
			addCreators: []*domain.GameCreator{
				domain.NewGameCreator(
					values.NewGameCreatorID(), // 新しいID
					creator1.GetUserID(),      // 既存と同じ userID
					creator1.GetGameID(),      // 既存と同じ gameID
					values.NewTrapMemberName("えびせんMk-2"),
					now,
				),
			},
			beforeCreators: []*schema.GameCreatorTable{
				{
					ID:        uuid.UUID(creator1.GetID()),
					UserID:    uuid.UUID(creator1.GetUserID()),
					UserName:  string(creator1.GetUserName()),
					GameID:    uuid.UUID(creator1.GetGameID()),
					CreatedAt: creator1.GetCreatedAt(),
				},
			},
			wantDeltaCreators: 0,
			err:               repository.ErrDuplicatedUniqueKey,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.beforeCreators) > 0 {
				err = db.Create(testCase.beforeCreators).Error
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameCreatorTable{}).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			err = repo.CreateGameCreators(t.Context(), testCase.addCreators)
			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
			} else {
				assert.NoError(t, err)
			}

			var creators []*schema.GameCreatorTable
			err = db.Find(&creators).Error
			require.NoError(t, err)

			assert.Equal(t, testCase.wantDeltaCreators+len(testCase.beforeCreators), len(creators))

			actualByID := make(map[values.GameCreatorID]*schema.GameCreatorTable, len(creators))
			for _, c := range creators {
				actualByID[values.GameCreatorID(c.ID)] = c
			}

			beforeIDSet := make(map[values.GameCreatorID]bool, len(testCase.beforeCreators))
			for _, c := range testCase.beforeCreators {
				beforeID := values.GameCreatorID(c.ID)
				beforeIDSet[beforeID] = true
				actual, ok := actualByID[beforeID]
				require.Truef(t, ok, "beforeCreatorsのID %v が実際のcreatorsに存在しない", beforeID)
				assert.Equal(t, beforeID, values.GameCreatorID(actual.ID))
				assert.Equal(t, c.UserID, actual.UserID)
				assert.Equal(t, c.GameID, actual.GameID)
				assert.Equal(t, c.UserName, actual.UserName)
				assert.WithinDuration(t, c.CreatedAt, actual.CreatedAt, time.Second)
			}

			if testCase.err == nil {
				for _, c := range testCase.addCreators {
					actual, ok := actualByID[c.GetID()]
					if !assert.Truef(t, ok, "追加したゲームクリエイターのID %v が実際のcreatorsに存在しない", c.GetID()) {
						continue
					}
					assert.Equal(t, c.GetID(), values.GameCreatorID(actual.ID))
					assert.Equal(t, uuid.UUID(c.GetUserID()), actual.UserID)
					assert.Equal(t, uuid.UUID(c.GetGameID()), actual.GameID)
					assert.Equal(t, string(c.GetUserName()), actual.UserName)
					assert.WithinDuration(t, c.GetCreatedAt(), actual.CreatedAt, time.Second)
				}
			} else {
				for _, added := range testCase.addCreators {
					if _, existedBefore := beforeIDSet[added.GetID()]; existedBefore {
						continue
					}
					_, created := actualByID[added.GetID()]
					assert.Falsef(t, created, "unexpected creator created on error: %v", added.GetID())
				}
			}
		})
	}
}
