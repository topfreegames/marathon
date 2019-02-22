package worker

import (
	"errors"
	"fmt"

	redis "gopkg.in/redis.v5"
)

// StageStatus holds information about a stage from a worker pipeline in Redis
type StageStatus struct {
	JobID       string
	Stage       string
	StageKey    string
	Description string
	Client      *redis.Client

	MaxProgress     int
	CurrentProgress int
	Completed       bool

	SubStageStatus []*StageStatus
}

// NewStageStatus returns a new StageStatus instance
func NewStageStatus(client *redis.Client,
	jobID, stage, description string,
	maxProgress int) (*StageStatus, error) {

	if maxProgress == 0 {
		return nil, errors.New("can't create a stage with 0 maxProgress")
	}

	ss := &StageStatus{
		JobID:       jobID,
		Stage:       stage,
		StageKey:    fmt.Sprintf("%s-%s", jobID, stage),
		Description: description,
		Client:      client,

		MaxProgress:     maxProgress,
		CurrentProgress: 0,
		Completed:       false,

		SubStageStatus: make([]*StageStatus, 0),
	}

	ss.Client.HSet(ss.StageKey, "description", description)
	ss.Client.HSet(ss.JobID, ss.Stage, ss.StageKey)
	ss.Client.HSet(ss.StageKey, "max", maxProgress)
	ss.Client.HSet(ss.StageKey, "current", 0)

	return ss, nil
}

// NewSubStage creates a new StageStatus from a previous one and add it to its SubStages list
func (s *StageStatus) NewSubStage(
	description string,
	maxProgress int,
) (*StageStatus, error) {
	ss, err := NewStageStatus(
		s.Client,
		s.JobID, fmt.Sprintf("%s.%d", s.Stage, len(s.SubStageStatus)+1), description,
		maxProgress,
	)
	if err != nil {
		return nil, err
	}

	s.SubStageStatus = append(s.SubStageStatus, ss)
	return ss, err
}

// IncrProgress increments the StageStatus progress by 1 unit
func (s *StageStatus) IncrProgress() error {
	if s.Completed {
		return errors.New("stage is already finished")
	}

	val := s.Client.HIncrBy(s.StageKey, "current", 1).Val()
	if val == int64(s.MaxProgress) {
		s.Completed = true
	}

	return nil
}
