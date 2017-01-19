/*
 * Copyright (c) 2017 Felipe Cavalcanti <fjfcavalcanti@gmail.com>
 * Author: Felipe Cavalcanti <fjfcavalcanti@gmail.com>
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

package feedback

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Feedback Handler", func() {
	var logger zap.Logger
	var config *viper.Viper
	var jobID uuid.UUID
	var handler *Handler

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()),
			zap.FatalLevel,
		)
		config = viper.New()
		jobID = uuid.NewV4()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
		var err error
		handler, err = NewHandler(config, logger, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create a new instance", func() {
		It("should return a configured handler", func() {
			l, err := NewHandler(config, logger, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Config).NotTo(BeNil())
			Expect(l.FeedbackCache).NotTo(BeNil())
			Expect(l.FlushInterval).NotTo(BeNil())
			Expect(l.Logger).NotTo(BeNil())
		})

		It("should return a configured handler", func() {
			l, err := NewHandler(config, logger, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Config).NotTo(BeNil())
			Expect(l.FeedbackCache).NotTo(BeNil())
			Expect(l.FlushInterval).NotTo(BeNil())
			Expect(l.Logger).NotTo(BeNil())
		})
	})

	Describe("generatePGIncrJSON", func() {
		It("should generate the valid postgres query", func() {
			m := map[string]int{
				"ack":       10,
				"bad_token": 20,
			}
			q := handler.generatePGIncrJSON(jobID.String(), m)
			Expect(q).Should(SatisfyAny(
				Equal(fmt.Sprintf(`UPDATE jobs SET feedbacks = feedbacks || CONCAT('{"ack":',COALESCE(feedbacks->>'ack','0')::int + 10,',"bad_token":',COALESCE(feedbacks->>'bad_token','0')::int + 20,'}')::jsonb WHERE id = '%s';`, jobID.String())),
				Equal(fmt.Sprintf(`UPDATE jobs SET feedbacks = feedbacks || CONCAT('{"bad_token":',COALESCE(feedbacks->>'bad_token','0')::int + 20,',"ack":',COALESCE(feedbacks->>'ack','0')::int + 10,'}')::jsonb WHERE id = '%s';`, jobID.String())),
			))
		})

		It("should generate the valid postgres query", func() {
			m := map[string]int{
				"bad_token": 20,
			}
			q := handler.generatePGIncrJSON(jobID.String(), m)
			Expect(q).To(Equal(fmt.Sprintf(`UPDATE jobs SET feedbacks = feedbacks || CONCAT('{"bad_token":',COALESCE(feedbacks->>'bad_token','0')::int + 20,'}')::jsonb WHERE id = '%s';`, jobID.String())))
		})
	})

	Describe("handleMessage", func() {
		It("should handle a error message", func() {
			Expect(len(handler.FeedbackCache)).To(Equal(0))
			m := fmt.Sprintf("{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"nack\",\"error\":\"BAD_REGISTRATION\",\"category\":\"\",\"metadata\":{\"jobId\":\"%s\"}}", jobID.String())
			handler.handleMessage([]byte(m))
			Expect(len(handler.FeedbackCache)).To(Equal(1))
			Expect(handler.FeedbackCache[jobID.String()]).To(BeEquivalentTo(map[string]int{
				"BAD_REGISTRATION": 1,
			}))
		})

		It("should handle a success message", func() {
			Expect(len(handler.FeedbackCache)).To(Equal(0))
			m := fmt.Sprintf("{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"ack\",\"error\": null,\"category\":\"\",\"metadata\":{\"jobId\":\"%s\"}}", jobID.String())
			handler.handleMessage([]byte(m))
			Expect(len(handler.FeedbackCache)).To(Equal(1))
			Expect(handler.FeedbackCache[jobID.String()]).To(BeEquivalentTo(map[string]int{
				"ack": 1,
			}))
			handler.handleMessage([]byte(m))
			Expect(handler.FeedbackCache[jobID.String()]).To(BeEquivalentTo(map[string]int{
				"ack": 2,
			}))
		})

		It("should do nothing if message has no metadata", func() {
			Expect(len(handler.FeedbackCache)).To(Equal(0))
			m := "{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"ack\",\"error\": null,\"category\":\"\"}"
			handler.handleMessage([]byte(m))
			Expect(len(handler.FeedbackCache)).To(Equal(0))
		})
	})

	Describe("feedbackService", func() {
		It("should detect a valid service for a gcm message", func() {
			m := "{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"ack\",\"error\": null,\"category\":\"\"}"
			var message Message
			err := json.Unmarshal([]byte(m), &message)
			Expect(err).NotTo(HaveOccurred())
			s := handler.feedbackService(&message)
			Expect(s).To(Equal("gcm"))
		})

		It("should detect a valid service for a apns message", func() {
			m := "{\"DeviceToken\":\"\",\"ID\":\"\",\"Err\":{\"Key\":\"missing-device-token\",\"Description\":\"device token was not specified\"}}"
			var message Message
			err := json.Unmarshal([]byte(m), &message)
			Expect(err).NotTo(HaveOccurred())
			s := handler.feedbackService(&message)
			Expect(s).To(Equal("apns"))
		})
	})

	Describe("flushFeedbacks", func() {
		It("should delete map keys and exec query in postgres", func() {
			mockPG := testing.NewPGMock(0, 0, nil)
			mockDB, err := extensions.NewPGClient("db", config, logger, mockPG)
			Expect(err).NotTo(HaveOccurred())
			h, err := NewHandler(config, logger, nil, mockDB)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(handler.FeedbackCache)).To(Equal(0))
			m := fmt.Sprintf("{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"ack\",\"error\": null,\"category\":\"\",\"metadata\":{\"jobId\":\"%s\"}}", jobID.String())
			h.handleMessage([]byte(m))
			Expect(len(h.FeedbackCache)).To(Equal(1))
			h.FlushInterval = time.Duration(10) * time.Millisecond
			go h.flushFeedbacks()
			Eventually(func() int {
				return len(h.FeedbackCache)
			}).Should(Equal(0))
			Eventually(func() int {
				return len(mockPG.ExecOnes)
			}).Should(Equal(1))
		})
	})

	Describe("HandleMessages", func() {
		It("should handle messaages if HandleMessages is called", func() {
			mChan := make(chan []byte)
			go handler.HandleMessages(&mChan)
			Eventually(func() bool {
				return handler.run
			}).Should(BeTrue())
			m := fmt.Sprintf("{\"from\":\"F8DIN2OFA0X3IOV897KUVPWU9CR2GNGUIOODWUFFVMJTFGQB45CY0ZEKXV758JOY0Z46P2CUCVL9HMNI3UGE5YXZYDM1AM0DX5ENIEGESOOLOV23YCKXG39ODFJXCU3UZFIW5ZCWLSEGGM1MY7SSGT07\",\"message_id\":\"422fc070-bf0e-4005-86e9-6aafaee9f3dd\",\"message_type\":\"ack\",\"error\": null,\"category\":\"\",\"metadata\":{\"jobId\":\"%s\"}}", jobID.String())
			mChan <- []byte(m)
			Eventually(func() int {
				return len(handler.FeedbackCache)
			}).Should(Equal(1))
			Expect(handler.FeedbackCache[jobID.String()]).To(BeEquivalentTo(map[string]int{
				"ack": 1,
			}))
		})
	})

})
