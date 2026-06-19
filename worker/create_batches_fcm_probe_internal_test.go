/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 *
 * Internal (package worker) tests for the dispatch-time FCM-first routing
 * support on the create-batches read path: the fcm_token column probe, its
 * process-lifetime cache, and the column-aware SELECT in getUserBatchFromPG.
 */

package worker

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// confPath mirrors testing.GetConfPath(); the testing package can't be imported
// here (it pulls in api -> worker, an import cycle for an internal test).
const confPath = "../config/test.yaml"

var _ = Describe("CreateBatches FCM token probe", func() {
	logger := zap.New(zap.NewJSONEncoder(zap.NoTime()), zap.FatalLevel)
	w := NewWorker(logger, confPath)
	b := &CreateBatchesWorker{Workers: w}

	dropTables := func(tables ...string) {
		for _, t := range tables {
			_, err := w.PushDB.Exec("DROP TABLE IF EXISTS " + t)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	BeforeEach(func() {
		// Reset the process-lifetime cache so each spec starts clean; the cache
		// is package-global and otherwise leaks across specs.
		fcmTokenColumnCacheMu.Lock()
		fcmTokenColumnCache = map[string]bool{}
		fcmTokenColumnCacheMu.Unlock()
	})

	Describe("pushTableHasFcmToken", func() {
		It("reports true for a table carrying the fcm_token column", func() {
			dropTables("probewith_apns")
			_, err := w.PushDB.Exec(`CREATE TABLE probewith_apns (
				user_id text, token text, locale text, tz text, fcm_token text)`)
			Expect(err).NotTo(HaveOccurred())
			defer dropTables("probewith_apns")

			Expect(b.pushTableHasFcmToken("probewith_apns")).To(BeTrue())
		})

		It("reports false for a table without the fcm_token column", func() {
			dropTables("probewithout_apns")
			_, err := w.PushDB.Exec(`CREATE TABLE probewithout_apns (
				user_id text, token text, locale text, tz text)`)
			Expect(err).NotTo(HaveOccurred())
			defer dropTables("probewithout_apns")

			Expect(b.pushTableHasFcmToken("probewithout_apns")).To(BeFalse())
		})

		It("reports false when the table does not exist", func() {
			dropTables("probemissing_apns")
			Expect(b.pushTableHasFcmToken("probemissing_apns")).To(BeFalse())
		})

		It("caches the first result for the process lifetime (migrate-while-running needs a restart)", func() {
			dropTables("probecache_apns")
			_, err := w.PushDB.Exec(`CREATE TABLE probecache_apns (
				user_id text, token text, locale text, tz text)`)
			Expect(err).NotTo(HaveOccurred())
			defer dropTables("probecache_apns")

			// First probe: no column -> false, and that's cached.
			Expect(b.pushTableHasFcmToken("probecache_apns")).To(BeFalse())

			// Add the column after the cache was populated.
			_, err = w.PushDB.Exec("ALTER TABLE probecache_apns ADD COLUMN fcm_token text")
			Expect(err).NotTo(HaveOccurred())

			// Still false: the cached value is served, mirroring push-api's AR
			// column cache. A restart (cache reset) is required to see the column.
			Expect(b.pushTableHasFcmToken("probecache_apns")).To(BeFalse())

			fcmTokenColumnCacheMu.Lock()
			fcmTokenColumnCache = map[string]bool{}
			fcmTokenColumnCacheMu.Unlock()

			// After a "restart" the freshly-probed table now reports the column.
			Expect(b.pushTableHasFcmToken("probecache_apns")).To(BeTrue())
		})
	})

	Describe("getUserBatchFromPG", func() {
		It("populates User.FcmToken from a migrated table", func() {
			dropTables("fcmread_apns")
			_, err := w.PushDB.Exec(`CREATE TABLE fcmread_apns (
				user_id text, token text, locale text, tz text, fcm_token text)`)
			Expect(err).NotTo(HaveOccurred())
			defer dropTables("fcmread_apns")
			_, err = w.PushDB.Exec(`INSERT INTO fcmread_apns (user_id, token, locale, tz, fcm_token)
				VALUES ('u1', 'apns-tok-1', 'en', '-0300', 'fcm-tok-1'),
				       ('u2', 'apns-tok-2', 'en', '-0300', NULL)`)
			Expect(err).NotTo(HaveOccurred())

			job := &model.Job{Service: "apns", App: model.App{Name: "fcmread"}}
			ids := []string{"u1", "u2"}
			users := *b.getUserBatchFromPG(&ids, job)

			byID := map[string]User{}
			for _, u := range users {
				byID[u.UserID] = u
			}
			Expect(byID).To(HaveLen(2))
			Expect(byID["u1"].FcmToken).To(Equal("fcm-tok-1"))
			Expect(byID["u1"].Token).To(Equal("apns-tok-1"))
			// NULL fcm_token decodes to the empty string -> APNs fallback at dispatch.
			Expect(byID["u2"].FcmToken).To(Equal(""))
			Expect(byID["u2"].Token).To(Equal("apns-tok-2"))
		})

		It("reads a non-migrated table without error and leaves FcmToken empty", func() {
			dropTables("nofcmread_apns")
			_, err := w.PushDB.Exec(`CREATE TABLE nofcmread_apns (
				user_id text, token text, locale text, tz text)`)
			Expect(err).NotTo(HaveOccurred())
			defer dropTables("nofcmread_apns")
			_, err = w.PushDB.Exec(`INSERT INTO nofcmread_apns (user_id, token, locale, tz)
				VALUES ('u1', 'apns-tok-1', 'en', '-0300')`)
			Expect(err).NotTo(HaveOccurred())

			job := &model.Job{Service: "apns", App: model.App{Name: "nofcmread"}}
			ids := []string{"u1"}
			users := *b.getUserBatchFromPG(&ids, job)

			Expect(users).To(HaveLen(1))
			Expect(users[0].Token).To(Equal("apns-tok-1"))
			Expect(users[0].FcmToken).To(Equal(""))
		})
	})
})
