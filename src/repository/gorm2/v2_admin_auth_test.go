package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestAddAdminV2(t *testing.T) {
	t.Parallel()

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
			if testCase.beforeAdmins != nil && len(testCase.beforeAdmins) != 0 {
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
