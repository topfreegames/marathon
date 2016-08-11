package models_test

import (
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

type Table struct {
	TableName string `db:"tablename"`
}

var _ = Describe("Models", func() {
	var (
		db models.DB
	)
	BeforeEach(func() {
		_db, dbErr := models.GetTestDB()
		Expect(dbErr).To(BeNil())
		Expect(_db).NotTo(BeNil())
		db = _db

		// Truncate all tables
		var tables []Table
		_, _ = db.Select(&tables, "SELECT tablename from pg_tables where schemaname='public'")
		var tableNames []string
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName)
		}
		if len(tableNames) > 0 {
			_, err := db.Exec(fmt.Sprintf("TRUNCATE %s", strings.Join(tableNames, ",")))
			Expect(err).To(BeNil())
		}
	})

	Describe("UserToken", func() {
		Describe("Create UserTokens Table", func() {
			It("Should succeed when data is correct", func() {
				app := "app_test_1_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())

				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))
			})

			It("Should succeed creating 2 tables", func() {
				app1 := "app_test_2_1"
				service1 := "apns"
				createdTable1, err := models.CreateUserTokensTable(db, app1, service1)
				Expect(err).To(BeNil())
				Expect(createdTable1.TableName).To(Equal(models.GetTableName(app1, service1)))

				app2 := "app_test_2_2"
				service2 := "apns"
				createdTable2, err := models.CreateUserTokensTable(db, app2, service2)
				Expect(err).To(BeNil())

				Expect(createdTable2.TableName).To(Equal(models.GetTableName(app2, service2)))
			})
		})

		Describe("Create UserToken", func() {
			It("Should create a user token through a factory", func() {
				app := "app_test_2_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userToken, err := CreateUserTokenFactory(db, map[string]interface{}{})
				Expect(err).To(BeNil())
				err = db.Insert(userToken)
				Expect(err).To(BeNil())

				dbUserToken, err := models.GetUserTokenByID(db, app, service, userToken.ID)
				Expect(err).To(BeNil())
				Expect(dbUserToken.Token).To(Equal(userToken.Token))
				Expect(dbUserToken.UserID).To(Equal(userToken.UserID))
				Expect(dbUserToken.Locale).To(Equal(userToken.Locale))
				Expect(dbUserToken.Region).To(Equal(userToken.Region))
				Expect(dbUserToken.Tz).To(Equal(userToken.Tz))
				Expect(dbUserToken.BuildN).To(Equal(userToken.BuildN))
			})

			It("Should upsert a userToken when new userToken", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+04:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

				userToken, err := models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				dbUserToken, err := models.GetUserTokenByID(db, app, service, userToken.ID)
				Expect(err).To(BeNil())
				Expect(dbUserToken.Token).To(Equal(userToken.Token))
				Expect(dbUserToken.UserID).To(Equal(userToken.UserID))
				Expect(dbUserToken.Locale).To(Equal(userToken.Locale))
				Expect(dbUserToken.Region).To(Equal(userToken.Region))
				Expect(dbUserToken.Tz).To(Equal(userToken.Tz))
				Expect(dbUserToken.BuildN).To(Equal(userToken.BuildN))
			})

			It("Should upsert a userToken when userToken exists", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+03:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

				userToken, err := models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				locale = uuid.NewV4().String()[:2]
				region = uuid.NewV4().String()[:2]
				tz = "GMT+04:00"
				buildN = uuid.NewV4().String()
				optOut = []string{uuid.NewV4().String(), uuid.NewV4().String()}

				userToken, err = models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				dbUserToken, err := models.GetUserTokenByID(db, app, service, userToken.ID)
				Expect(err).To(BeNil())
				Expect(dbUserToken.Token).To(Equal(userToken.Token))
				Expect(dbUserToken.UserID).To(Equal(userToken.UserID))
				Expect(dbUserToken.Locale).To(Equal(userToken.Locale))
				Expect(dbUserToken.Region).To(Equal(userToken.Region))
				Expect(dbUserToken.Tz).To(Equal(userToken.Tz))
				Expect(dbUserToken.BuildN).To(Equal(userToken.BuildN))
			})

			It("Should upsert a userToken when userToken exists", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+03:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

				userToken, err := models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				userID = uuid.NewV4().String()
				locale = uuid.NewV4().String()[:2]
				region = uuid.NewV4().String()[:2]
				tz = "GMT+04:00"
				buildN = uuid.NewV4().String()
				optOut = []string{uuid.NewV4().String(), uuid.NewV4().String()}

				userToken, err = models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				dbUserToken, err := models.GetUserTokenByID(db, app, service, userToken.ID)
				Expect(err).To(BeNil())
				Expect(dbUserToken.Token).To(Equal(userToken.Token))
				Expect(dbUserToken.UserID).To(Equal(userToken.UserID))
				Expect(dbUserToken.Locale).To(Equal(userToken.Locale))
				Expect(dbUserToken.Region).To(Equal(userToken.Region))
				Expect(dbUserToken.Tz).To(Equal(userToken.Tz))
				Expect(dbUserToken.BuildN).To(Equal(userToken.BuildN))
			})

			It("Should find userToken by filters", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+03:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}
				userToken, err := models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				filters := [][]interface{}{
					{"user_id", userID},
				}
				modifiers := [][]interface{}{
					{"ORDER BY", "updated_at ASC"},
					{"LIMIT", 100},
					{"OFFSET", 0},
				}

				dbUserTokens, err := models.GetUserTokensBatchByFilters(db, app, service, filters, modifiers)
				Expect(err).To(BeNil())

				Expect(len(dbUserTokens)).To(Equal(1))
				Expect(dbUserTokens[0].Token).To(Equal(userToken.Token))
				Expect(dbUserTokens[0].UserID).To(Equal(userToken.UserID))
				Expect(dbUserTokens[0].Locale).To(Equal(userToken.Locale))
				Expect(dbUserTokens[0].Region).To(Equal(userToken.Region))
				Expect(dbUserTokens[0].Tz).To(Equal(userToken.Tz))
				Expect(dbUserTokens[0].BuildN).To(Equal(userToken.BuildN))
			})

			It("Should find userTokens by filters", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+03:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}
				userToken, err := models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				userID2 := uuid.NewV4().String()
				token2 := uuid.NewV4().String()
				locale2 := uuid.NewV4().String()[:2]
				region2 := uuid.NewV4().String()[:2]
				tz2 := "GMT+04:00"
				optOut2 := []string{uuid.NewV4().String(), uuid.NewV4().String()}
				userToken2, err := models.UpsertToken(
					db, app, service, userID2, token2, locale2, region2, tz2, buildN, optOut2,
				)
				Expect(err).To(BeNil())

				filters := [][]interface{}{
					{"build_n", buildN},
				}
				modifiers := [][]interface{}{
					{"ORDER BY", "updated_at ASC"},
					{"LIMIT", 100},
					{"OFFSET", 0},
				}

				dbUserTokens, err := models.GetUserTokensBatchByFilters(db, app, service, filters, modifiers)
				Expect(err).To(BeNil())

				Expect(len(dbUserTokens)).To(Equal(2))

				Expect(dbUserTokens[0].Token).To(Equal(userToken.Token))
				Expect(dbUserTokens[0].UserID).To(Equal(userToken.UserID))
				Expect(dbUserTokens[0].Locale).To(Equal(userToken.Locale))
				Expect(dbUserTokens[0].Region).To(Equal(userToken.Region))
				Expect(dbUserTokens[0].Tz).To(Equal(userToken.Tz))
				Expect(dbUserTokens[0].BuildN).To(Equal(userToken.BuildN))

				Expect(dbUserTokens[1].Token).To(Equal(userToken2.Token))
				Expect(dbUserTokens[1].UserID).To(Equal(userToken2.UserID))
				Expect(dbUserTokens[1].Locale).To(Equal(userToken2.Locale))
				Expect(dbUserTokens[1].Region).To(Equal(userToken2.Region))
				Expect(dbUserTokens[1].Tz).To(Equal(userToken2.Tz))
				Expect(dbUserTokens[1].BuildN).To(Equal(userToken2.BuildN))
			})

			It("Should ont find userToken by filters when there are no matches", func() {
				app := "app_test_3_1"
				service := "apns"
				createdTable, err := models.CreateUserTokensTable(db, app, service)
				Expect(err).To(BeNil())
				Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

				userID := uuid.NewV4().String()
				token := uuid.NewV4().String()
				locale := uuid.NewV4().String()[:2]
				region := uuid.NewV4().String()[:2]
				tz := "GMT+03:00"
				buildN := uuid.NewV4().String()
				optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}
				_, err = models.UpsertToken(
					db, app, service, userID, token, locale, region, tz, buildN, optOut,
				)
				Expect(err).To(BeNil())

				filters := [][]interface{}{
					{"user_id", uuid.NewV4().String()},
				}
				modifiers := [][]interface{}{
					{"ORDER BY", "updated_at ASC"},
					{"LIMIT", 100},
					{"OFFSET", 0},
				}

				dbUserTokens, err := models.GetUserTokensBatchByFilters(db, app, service, filters, modifiers)
				Expect(err).To(BeNil())

				Expect(len(dbUserTokens)).To(Equal(0))
			})
		})
	})
})
