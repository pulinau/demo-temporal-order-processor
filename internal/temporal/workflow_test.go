package temporal_test

import (
	"testing"
	"time"

	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"
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

	activities *temporal.OrderActivities
}

func (s *WorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.activities = &temporal.OrderActivities{}
}

func (s *WorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *WorkflowTestSuite) TestWorkflow_Success() {
	// Mock activity implementations.

	s.env.OnActivity(s.activities.Validate, mock.Anything, temporal.Order{}).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("pickOrder", nil)
	}, time.Minute)

	s.env.OnActivity(s.activities.Process, mock.Anything, temporal.Order{}).Return("PROCESSED", nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("shipOrder", nil)
	}, 2*time.Hour)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("markOrderAsDelivered", nil)
	}, 5*24*time.Hour)

	// Execute workflow.

	s.env.ExecuteWorkflow(temporal.ProccessOrder, temporal.Params{Order: temporal.Order{}})

	// Assert execution and order status.

	s.Require().NoError(s.env.GetWorkflowError())

	val, err := s.env.QueryWorkflow("GetOrderStatus")
	s.Require().NoError(err, "workflow should be queryable")
	var got temporal.OrderStatus
	err = val.Get(&got)
	s.Require().NoError(err, "query result should be a temporal.OrderStatus")
	s.Equal(temporal.Completed, got, "order should be completed")
}

func (s *WorkflowTestSuite) TestWorkflow_Cancelled() {
	// Mock activity implementations.

	s.env.OnActivity(s.activities.Validate, mock.Anything, temporal.Order{}).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("cancelOrder", nil)
	}, time.Minute)

	// Execute workflow.

	s.env.ExecuteWorkflow(temporal.ProccessOrder, temporal.Params{Order: temporal.Order{}})

	// Assert execution and order status.

	s.Require().NoError(s.env.GetWorkflowError())

	val, err := s.env.QueryWorkflow("GetOrderStatus")
	s.Require().NoError(err, "workflow should be queryable")
	var got temporal.OrderStatus
	err = val.Get(&got)
	s.Require().NoError(err, "query result should be a temporal.OrderStatus")
	s.Equal(temporal.Cancelled, got, "order should be cancelled")
}
