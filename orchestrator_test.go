package patron

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkerOrchestratorTestSuite struct {
	suite.Suite

	workerOrchestrator WorkerOrchestrator
}

func TestWorkerOrchestratorTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerOrchestratorTestSuite))
}

func (suite *WorkerOrchestratorTestSuite) SetupTest() {
	workerOrchestrator := NewWorkerOrchestrator(
		NewWorkerArray(func(job *Job) error {
			time.Sleep(1 * time.Second)
			payloadName, err := job.GetPayload("name")
			if err != nil {
				return err
			}

			fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, payloadName)

			return nil
		}, 5),
	)

	suite.workerOrchestrator = workerOrchestrator
}

func (suite *WorkerOrchestratorTestSuite) TestNewWorkerOrchestrator() {
	suite.NotNil(NewWorkerOrchestrator(
		NewWorkerArray(nil, 1),
	))
}

func (suite *WorkerOrchestratorTestSuite) TestAddJobToQueue() {
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	suite.Equal(1, suite.workerOrchestrator.GetQueueLength())
}

// Note: this test is not reliable.
func (suite *WorkerOrchestratorTestSuite) TestStartAsAsync() {
	suite.T().Skip()

	workerResultCh := make(chan *WorkerResult)
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test",
		},
	})
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      11,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test2",
		},
	})

	suite.workerOrchestrator.StartAsAsync(context.Background(), workerResultCh)
	suite.NotEmpty(<-workerResultCh)
	suite.NotEmpty(<-workerResultCh)
}

func (suite *WorkerOrchestratorTestSuite) TestStartAllJobsSuccess() {
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test",
		},
	})
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      11,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test2",
		},
	})

	results := suite.workerOrchestrator.Start(context.Background())

	suite.Len(results, 2)
	suite.Equal(nil, results[0].Error)
	suite.Equal(nil, results[1].Error)
}

func (suite *WorkerOrchestratorTestSuite) TestStartOneJobFailure() {
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test",
		},
	})
	suite.workerOrchestrator.AddJobToQueue(&Job{
		ID:      11,
		Context: context.Background(),
		Payload: map[string]interface{}{},
	})

	results := suite.workerOrchestrator.Start(context.Background())

	suite.Len(results, 2)
	for _, result := range results {
		if result.JobID == 11 {
			suite.ErrorIs(result.Error, ErrJobPayloadNotFound)
		} else {
			suite.Equal(nil, result.Error)
		}
	}
}
