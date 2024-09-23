package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestAddAdminV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	adminAuthRepository := NewAdminAuth(testDB)

	type test struct {
		description  string
		userID       values.TraPMemberID
		beforeAdmins []migrate.AdminTable
		isErr        bool
		err          error
	}

	traPMemberID1 := values.NewTrapMemberID(uuid.New())
	traPMemberID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			userID:      traPMemberID1,
		},
		{
			description: "既に登録されているのでエラー",
			userID:      traPMemberID2,
			beforeAdmins: []migrate.AdminTable{
				{
					UserID: uuid.UUID(traPMemberID2),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Delete(&migrate.AdminTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete admins: %+v\n", err)
				}
			}()

			if len(testCase.beforeAdmins) != 0 {
				err := db.Session(&gorm.Session{}).Create(&testCase.beforeAdmins).Error
				if err != nil {
					t.Fatalf("failed to create admin: %+v\n", err)
				}
			}

			err := adminAuthRepository.AddAdmin(ctx, testCase.userID)

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

			var admin migrate.AdminTable
			err = db.Session(&gorm.Session{}).Where("user_id = ?", uuid.UUID(testCase.userID)).First(&admin).Error
			if err != nil {
				t.Fatalf("failed to get admin: %+v\n", err)
			}

			assert.Equal(t, uuid.UUID(testCase.userID), admin.UserID)
		})
	}
}

func TestGetAdminsV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	adminAuthRepository := NewAdminAuth(testDB)

	type test struct {
		description  string
		beforeAdmins []migrate.AdminTable
		adminsMap    map[values.TraPMemberID]struct{} // 返り値での要素の順序が定まらないため
		isErr        bool
		err          error
	}

	traPMemberID1 := values.NewTrapMemberID(uuid.New())
	traPMemberID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			beforeAdmins: []migrate.AdminTable{
				{
					UserID: uuid.UUID(traPMemberID1),
				},
			},
			adminsMap: map[values.TraPMemberID]struct{}{
				traPMemberID1: {},
			},
		},
		{
			description:  "adminが存在しなくてもエラーなし",
			beforeAdmins: []migrate.AdminTable{},
			adminsMap:    map[values.TraPMemberID]struct{}{},
		},
		{
			description: "adminが複数でもエラーなし",
			beforeAdmins: []migrate.AdminTable{
				{
					UserID: uuid.UUID(traPMemberID1),
				},
				{
					UserID: uuid.UUID(traPMemberID2),
				},
			},
			adminsMap: map[values.TraPMemberID]struct{}{
				traPMemberID1: {},
				traPMemberID2: {},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Delete(&migrate.AdminTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete admins: %+v\n", err)
				}
			}()

			if len(testCase.beforeAdmins) != 0 {
				err := db.Create(&testCase.beforeAdmins).Error
				if err != nil {
					t.Fatalf("failed to create test data: %+v\n", err)
				}
			}

			admins, err := adminAuthRepository.GetAdmins(ctx)
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

			assert.Len(t, admins, len(testCase.adminsMap))
			for _, admin := range admins {
				_, ok := testCase.adminsMap[admin]
				assert.True(t, ok)
			}

		})
	}
}

func TestDeleteAdminV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	adminAuthRepository := NewAdminAuth(testDB)

	type test struct {
		description   string
		userID        values.TraPMemberID
		beforeAdmins  []migrate.AdminTable
		afterAdminMap map[values.TraPMemberID]struct{} // DBからの取得時の順序指定ができないため
		isErr         bool
		err           error
	}

	traPMemberID1 := values.NewTrapMemberID(uuid.New())
	traPMemberID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			userID:      traPMemberID1,
			beforeAdmins: []migrate.AdminTable{
				{
					UserID: uuid.UUID(traPMemberID1),
				},
			},
			afterAdminMap: map[values.TraPMemberID]struct{}{},
		},
		{
			description: "他のadminが存在してもエラーなし",
			userID:      traPMemberID1,
			beforeAdmins: []migrate.AdminTable{
				{
					UserID: uuid.UUID(traPMemberID1),
				},
				{
					UserID: uuid.UUID(traPMemberID2),
				},
			},
			afterAdminMap: map[values.TraPMemberID]struct{}{
				traPMemberID2: {},
			},
		},
		{
			description:   "adminが存在しないのでErrNoRecordDeleted",
			userID:        traPMemberID1,
			beforeAdmins:  []migrate.AdminTable{},
			afterAdminMap: map[values.TraPMemberID]struct{}{},
			isErr:         true,
			err:           repository.ErrNoRecordDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Delete(&migrate.AdminTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete admins: %+v\n", err)
				}
			}()

			if len(testCase.beforeAdmins) != 0 {
				err := db.Session(&gorm.Session{}).Create(&testCase.beforeAdmins).Error
				if err != nil {
					t.Fatalf("failed to create admin: %+v\n", err)
				}
			}

			err := adminAuthRepository.DeleteAdmin(ctx, testCase.userID)
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

			var admins []migrate.AdminTable
			err = db.
				Unscoped().
				Session(&gorm.Session{}).
				Find(&admins).Error
			if err != nil {
				t.Fatalf("failed to get games: %+v\n", err)
			}

			assert.Len(t, admins, len(testCase.afterAdminMap))

			for _, admin := range admins {
				_, ok := testCase.afterAdminMap[values.TraPMemberID(admin.UserID)]
				assert.True(t, ok)
			}
		})
	}
}
