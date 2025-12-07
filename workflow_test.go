package demotemporalorderprocessing_test

import (
	"testing"

	demotemporalorderprocessing "github.com/pulinau/demo-temporal-order-processor"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}

type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment

	activities *demotemporalorderprocessing.OrderActivities
}

func (s *WorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.activities = &demotemporalorderprocessing.OrderActivities{}
}

func (s *WorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *WorkflowTestSuite) TestWorkflow_Success() {
	// Mock activity implementation
	s.env.OnActivity(s.activities.Validate, mock.Anything, demotemporalorderprocessing.Order{}).Return(nil)
	s.env.OnActivity(s.activities.Process, mock.Anything, demotemporalorderprocessing.Order{}).Return("PROCESSED", nil)

	s.env.ExecuteWorkflow(demotemporalorderprocessing.ProccessOrder, demotemporalorderprocessing.Params{Order: demotemporalorderprocessing.Order{}})

	s.Require().NoError(s.env.GetWorkflowError())
}
