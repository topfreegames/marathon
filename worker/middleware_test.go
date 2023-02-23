package worker_test

import (
	"github.com/jrallison/go-workers"
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

func (v *ValidateCreateBatchesMiddleware) Call(queue string, message *workers.Msg, next func() bool) (r bool) {
	{
		if !v.IsTestRunning {
			return next()
		}
		r = false
		if queue == "process_batch_worker" {

			dbUpdated := false // check database state
			job := &model.Job{}
			err := v.MarathonDB.Model(job).Where("id = ?", v.JobID).Select()
			if err == nil && job.TotalUsers == v.ExpectedTotalUsers && job.TotalTokens == v.ExpectedTotalTokens {
				dbUpdated = true
			}

			*v.ResultChan <- dbUpdated
		}
		return
	}
}
