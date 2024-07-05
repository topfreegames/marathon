package worker_test

import (
	goworkers2 "github.com/digitalocean/go-workers2"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/model"
)

type ValidateCreateBatchesMiddleware struct {
	IsTestRunning       bool
	ResultChan          *chan bool
	JobID               uuid.UUID
	MarathonDB          interfaces.DB
	ExpectedTotalUsers  int
	ExpectedTotalTokens int
}

func (v *ValidateCreateBatchesMiddleware) Call(queue string, m *goworkers2.Manager, next goworkers2.JobFunc) goworkers2.JobFunc {
	return func(message *goworkers2.Msg) error {
		if !v.IsTestRunning {
			return next(message)
		}
		if queue == "process_batch_worker" {
			dbUpdated := false // check database state
			job := &model.Job{}
			err := v.MarathonDB.Model(job).Where("id = ?", v.JobID).Select()
			if err == nil && job.TotalUsers == v.ExpectedTotalUsers && job.TotalTokens == v.ExpectedTotalTokens {
				dbUpdated = true
			}

			*v.ResultChan <- dbUpdated
		}
		return next(message)
	}

}
