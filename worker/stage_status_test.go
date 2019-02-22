package worker_test

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
	redis "gopkg.in/redis.v5"
)

var _ = Describe("ProcessBatch Worker", func() {
	var config *viper.Viper
	var logger zap.Logger
	var redisClient *redis.Client
	var err error

	BeforeEach(func() {
		config = GetConf()

		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)

		redisClient, err = extensions.NewRedis("workers", config, logger)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("stage status creation", func() {
		It("should be possible to create", func() {
			ss, err := worker.NewStageStatus(redisClient, "job1", "1", "first stage", 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(ss).NotTo(BeNil())
		})

		It("should not be possible to create a stage status with 0 max progress", func() {
			ss, err := worker.NewStageStatus(redisClient, "job1", "1", "first stage", 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("can't create a stage with 0 maxProgress"))
			Expect(ss).To(BeNil())
		})
	})

	Describe("stage status should be report to redis", func() {
		It("about a stage progress", func() {
			stage := "1"
			description := "first stage"
			maxProgress := 2

			ss, err := worker.NewStageStatus(redisClient, "job1", stage, description, maxProgress)
			Expect(err).NotTo(HaveOccurred())
			Expect(ss).NotTo(BeNil())

			jobStages := redisClient.HGetAll("job1").Val()
			Expect(jobStages).To(HaveKey(stage))
			Expect(jobStages["1"]).To(Equal("job1-1"))

			stageStats := redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["description"]).To(BeEquivalentTo(description))
			Expect(stageStats["current"]).To(BeEquivalentTo("0"))
			Expect(stageStats["max"]).To(BeEquivalentTo("2"))

			ss.IncrProgress()
			stageStats = redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["current"]).To(BeEquivalentTo("1"))

			ss.IncrProgress()
			stageStats = redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["current"]).To(BeEquivalentTo("2"))
		})

		It("about inner stages progress", func() {
			description1 := "stage 1"
			maxProgress1 := 2

			description2 := "stage 2"
			maxProgress2 := 20

			s1, err := worker.NewStageStatus(redisClient, "job1", "1", description1, maxProgress1)
			Expect(err).NotTo(HaveOccurred())
			Expect(s1).NotTo(BeNil())

			s2, err := worker.NewStageStatus(redisClient, "job1", "2", description2, maxProgress2)
			Expect(err).NotTo(HaveOccurred())
			Expect(s2).NotTo(BeNil())

			description1_1 := "stage 1.1"
			maxProgress1_1 := 3
			s1_1, err := s1.NewSubStage(description1_1, maxProgress1_1)
			Expect(err).NotTo(HaveOccurred())
			Expect(s1_1).NotTo(BeNil())
			Expect(len(s1.SubStageStatus)).To(Equal(1))

			description1_1_1 := "stage 1.1.1"
			maxProgress1_1_1 := 5
			s1_1_1, err := s1_1.NewSubStage(description1_1_1, maxProgress1_1_1)
			Expect(err).NotTo(HaveOccurred())
			Expect(s1_1_1).NotTo(BeNil())
			Expect(len(s1_1.SubStageStatus)).To(Equal(1))

			description1_1_2 := "stage 1.1.2"
			maxProgress1_1_2 := 7
			s1_1_2, err := s1_1.NewSubStage(description1_1_2, maxProgress1_1_2)
			Expect(err).NotTo(HaveOccurred())
			Expect(s1_1_2).NotTo(BeNil())
			Expect(len(s1_1.SubStageStatus)).To(Equal(2))

			description1_2 := "stage 1.2"
			maxProgress1_2 := 4
			s1_2, err := s1.NewSubStage(description1_2, maxProgress1_2)
			Expect(err).NotTo(HaveOccurred())
			Expect(s1_2).NotTo(BeNil())
			Expect(len(s1.SubStageStatus)).To(Equal(2))

			jobStages := redisClient.HGetAll("job1").Val()
			Expect(jobStages).To(HaveKey("1"))
			Expect(jobStages).To(HaveKey("1.1"))
			Expect(jobStages).To(HaveKey("1.1.1"))
			Expect(jobStages).To(HaveKey("1.1.2"))
			Expect(jobStages).To(HaveKey("1.2"))
			Expect(jobStages).To(HaveKey("2"))
			Expect(jobStages["1"]).To(Equal("job1-1"))
			Expect(jobStages["1.1"]).To(Equal("job1-1.1"))
			Expect(jobStages["1.1.1"]).To(Equal("job1-1.1.1"))
			Expect(jobStages["1.1.2"]).To(Equal("job1-1.1.2"))
			Expect(jobStages["1.2"]).To(Equal("job1-1.2"))
			Expect(jobStages["2"]).To(Equal("job1-2"))

			stage1Stats := redisClient.HGetAll("job1-1").Val()
			Expect(stage1Stats["description"]).To(Equal(description1))
			Expect(stage1Stats["current"]).To(Equal("0"))
			Expect(stage1Stats["max"]).To(Equal(strconv.Itoa(maxProgress1)))

			stage2Stats := redisClient.HGetAll("job1-2").Val()
			Expect(stage2Stats["description"]).To(Equal(description2))
			Expect(stage2Stats["current"]).To(Equal("0"))
			Expect(stage2Stats["max"]).To(Equal(strconv.Itoa(maxProgress2)))

			stage1_1Stats := redisClient.HGetAll("job1-1.1").Val()
			Expect(stage1_1Stats["description"]).To(Equal(description1_1))
			Expect(stage1_1Stats["current"]).To(Equal("0"))
			Expect(stage1_1Stats["max"]).To(Equal(strconv.Itoa(maxProgress1_1)))

			stage1_1_1Stats := redisClient.HGetAll("job1-1.1.1").Val()
			Expect(stage1_1_1Stats["description"]).To(Equal(description1_1_1))
			Expect(stage1_1_1Stats["current"]).To(Equal("0"))
			Expect(stage1_1_1Stats["max"]).To(Equal(strconv.Itoa(maxProgress1_1_1)))

			stage1_2Stats := redisClient.HGetAll("job1-1.2").Val()
			Expect(stage1_2Stats["description"]).To(Equal(description1_2))
			Expect(stage1_2Stats["current"]).To(Equal("0"))
			Expect(stage1_2Stats["max"]).To(Equal(strconv.Itoa(maxProgress1_2)))

			s1.IncrProgress()
			s1.IncrProgress()
			stage1Stats = redisClient.HGetAll("job1-1").Val()
			Expect(stage1Stats["current"]).To(BeEquivalentTo("2"))

			s1_1.IncrProgress()
			s1_1.IncrProgress()
			s1_1.IncrProgress()
			stage1_1Stats = redisClient.HGetAll("job1-1.1").Val()
			Expect(stage1_1Stats["current"]).To(BeEquivalentTo("3"))

			s1_1_1.IncrProgress()
			s1_1_1.IncrProgress()
			s1_1_1.IncrProgress()
			s1_1_1.IncrProgress()
			stage1_1_1Stats = redisClient.HGetAll("job1-1.1.1").Val()
			Expect(stage1_1_1Stats["current"]).To(BeEquivalentTo("4"))

			stage1_2Stats = redisClient.HGetAll("job1-1.2").Val()
			Expect(stage1_2Stats["current"]).To(BeEquivalentTo("0"))
		})

		It("should not be possible to increase a status more than its max progress", func() {
			stage := "1"
			description := "first stage"
			maxProgress := 1

			ss, err := worker.NewStageStatus(redisClient, "job1", stage, description, maxProgress)
			Expect(err).NotTo(HaveOccurred())
			Expect(ss).NotTo(BeNil())

			jobStages := redisClient.HGetAll("job1").Val()
			Expect(jobStages).To(HaveKey(stage))
			Expect(jobStages["1"]).To(Equal("job1-1"))

			stageStats := redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["description"]).To(BeEquivalentTo(description))
			Expect(stageStats["current"]).To(BeEquivalentTo("0"))
			Expect(stageStats["max"]).To(BeEquivalentTo("1"))

			err = ss.IncrProgress()
			Expect(err).NotTo(HaveOccurred())
			stageStats = redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["current"]).To(BeEquivalentTo("1"))

			err = ss.IncrProgress()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("stage is already finished"))
			stageStats = redisClient.HGetAll("job1-1").Val()
			Expect(stageStats["current"]).To(BeEquivalentTo("1"))

		})
	})
})
