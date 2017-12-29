/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/api"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("User Handler", func() {
	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	app := GetDefaultTestApp(logger)
	faultyDb := GetFaultyTestDB(app)
	var testUser *model.User
	BeforeEach(func() {
		app.DB.Exec("TRUNCATE TABLE users;")
		testUser = CreateTestUser(app.DB, map[string]interface{}{"email": "test@test.com", "isAdmin": true})
	})

	Describe("Get /users", func() {
		Describe("Successfully", func() {
			It("should return 200 and a single user", func() {
				status, body := Get(app, "/users", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(1))
			})

			It("should return 200 and a list of users", func() {
				testUsers := append([]*model.User{testUser}, CreateTestUsers(app.DB, 9)...)
				status, body := Get(app, "/users", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, user := range response {
					Expect(user["id"]).ToNot(BeNil())
					Expect(user["email"]).To(Equal(testUsers[idx].Email))
					Expect(user["isAdmin"]).To(Equal(testUsers[idx].IsAdmin))
					Expect(user["allowedApps"]).To(HaveLen(len(testUsers[idx].AllowedApps)))
					Expect(user["createdBy"]).To(Equal(testUsers[idx].CreatedBy))
					Expect(user["createdAt"]).ToNot(BeNil())
					Expect(user["createdAt"]).ToNot(Equal(0))
					Expect(user["updatedAt"]).ToNot(BeNil())
					Expect(user["updatedAt"]).ToNot(Equal(0))
				}
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Get(app, "/users", "")
				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 401 if invalid user", func() {
				status, _ := Get(app, "/users", "test2@test.com")
				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, "/users", "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})
		})
	})

	Describe("Post /users", func() {
		Describe("Successfully", func() {
			It("should return 201 and the created user", func() {
				payload := GetUserPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/users", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["id"]).ToNot(BeNil())
				Expect(response["email"]).To(Equal(payload["email"]))
				Expect(response["isAdmin"]).To(Equal(payload["isAdmin"]))
				Expect(response["createdBy"]).To(Equal("test@test.com"))
				Expect(response["createdAt"]).ToNot(BeNil())
				Expect(response["createdAt"]).ToNot(Equal(0))
				Expect(response["updatedAt"]).ToNot(BeNil())
				Expect(response["updatedAt"]).ToNot(Equal(0))

				id, err := uuid.FromString(response["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbUser := &model.User{
					ID: id,
				}
				err = app.DB.Select(dbUser)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser).NotTo(BeNil())
				Expect(dbUser.Email).To(Equal(payload["email"]))
				Expect(dbUser.IsAdmin).To(Equal(payload["isAdmin"]))
				Expect(dbUser.CreatedBy).To(Equal("test@test.com"))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, "/users", "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				payload := GetUserPayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, "/users", string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 409 if user with same email already exists", func() {
				existingUser := CreateTestUser(app.DB)
				payload := GetUserPayload(map[string]interface{}{"email": existingUser.Email})
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/users", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uix_users_email"))
			})

			It("should return 422 if missing email", func() {
				payload := GetUserPayload()
				delete(payload, "email")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/users", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid email"))
			})
		})
	})

	Describe("Get /users/:id", func() {
		Describe("Successfully", func() {
			It("should return 200 and the requested user", func() {
				existingUser := CreateTestUser(app.DB)
				status, body := Get(app, fmt.Sprintf("/users/%s", existingUser.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["id"]).To(Equal(existingUser.ID.String()))
				Expect(response["email"]).To(Equal(existingUser.Email))
				Expect(response["isAdmin"]).To(Equal(existingUser.IsAdmin))
				Expect(response["allowedApps"]).To(HaveLen(len(existingUser.AllowedApps)))
				Expect(response["createdBy"]).To(Equal(existingUser.CreatedBy))
				Expect(response["createdAt"]).ToNot(BeNil())
				Expect(response["createdAt"]).ToNot(Equal(0))
				Expect(response["updatedAt"]).ToNot(BeNil())
				Expect(response["updatedAt"]).ToNot(Equal(0))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Get(app, "/users/1234", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				existingUser := CreateTestUser(app.DB)
				app.DB = faultyDb
				status, _ := Get(app, fmt.Sprintf("/users/%s", existingUser.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the user does not exist", func() {
				status, _ := Get(app, fmt.Sprintf("/users/%s", uuid.NewV4().String()), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if user id is not UUID", func() {
				status, body := Get(app, "/users/not-uuid", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})

	Describe("Put /users/:uid", func() {
		Describe("successfully", func() {
			It("should return 200 and the updated user without updating createdBy", func() {
				existingUser := CreateTestUser(app.DB)
				payload := GetUserPayload(map[string]interface{}{"isAdmin": false, "allowedApps": []uuid.UUID{uuid.NewV4(), uuid.NewV4()}})
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/users/%s", existingUser.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["id"]).ToNot(BeNil())
				Expect(response["email"]).To(Equal(existingUser.Email))
				Expect(response["isAdmin"]).To(Equal(payload["isAdmin"]))
				Expect(response["createdBy"]).To(Equal(existingUser.CreatedBy))
				Expect(int64(response["createdAt"].(float64))).To(Equal(existingUser.CreatedAt))
				Expect(response["updatedAt"]).ToNot(Equal(existingUser.UpdatedAt))
				Expect(response["allowedApps"]).To(HaveLen(2))

				id, err := uuid.FromString(response["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbUser := &model.User{
					ID: id,
				}
				err = app.DB.Select(dbUser)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser).NotTo(BeNil())
				Expect(dbUser.Email).To(Equal(existingUser.Email))
				Expect(dbUser.IsAdmin).To(Equal(payload["isAdmin"]))
				Expect(dbUser.CreatedBy).To(Equal(existingUser.CreatedBy))
				Expect(dbUser.AllowedApps).To(HaveLen(2))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 500 if some error occurred", func() {
				existingUser := CreateTestUser(app.DB)
				payload := GetUserPayload()
				pl, _ := json.Marshal(payload)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Put(app, fmt.Sprintf("/users/%s", existingUser.ID), string(pl), "update@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 403 if user is not admin", func() {
				existingUser := CreateTestUser(app.DB)
				CreateTestUser(app.DB, map[string]interface{}{"email": "test2@test.com", "isAdmin": false})
				payload := GetUserPayload()
				pl, _ := json.Marshal(payload)
				status, _ := Put(app, fmt.Sprintf("/users/%s", existingUser.ID), string(pl), "test2@test.com")
				Expect(status).To(Equal(http.StatusForbidden))
			})

			It("should return 422 if user id is not UUID", func() {
				payload := GetUserPayload()
				pl, _ := json.Marshal(payload)

				status, body := Put(app, "/users/not-uuid", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})

	Describe("Delete /users/:id", func() {
		Describe("successfully", func() {
			It("should return 204 ", func() {
				existingUser := CreateTestUser(app.DB)
				status, _ := Delete(app, fmt.Sprintf("/users/%s", existingUser.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusNoContent))

				dbUser := &model.User{
					ID: existingUser.ID,
				}
				err := app.DB.Select(dbUser)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(api.RecordNotFoundString))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Delete(app, "/users/1234", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingUser := CreateTestUser(app.DB)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Delete(app, fmt.Sprintf("/users/%s", existingUser.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the user does not exist", func() {
				status, _ := Delete(app, fmt.Sprintf("/user/%s", uuid.NewV4().String()), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if user id is not UUID", func() {
				status, body := Delete(app, "/users/not-uuid", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})
})
