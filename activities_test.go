package demotemporalorderprocessing_test

import (
	"testing"

	"github.com/google/uuid"
	demotemporalorderprocessing "github.com/pulinau/demo-temporal-order-processor"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

const dummyOrderID = "8c727b70-cfcb-4674-8bcd-78e66e32f723"

func TestActivities(t *testing.T) {
	suite.Run(t, new(ActivityTestSuite))
}

type ActivityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestActivityEnvironment
}

func (s *ActivityTestSuite) SetupTest() {
	s.env = s.NewTestActivityEnvironment()
}

func (s *ActivityTestSuite) TestValidate_Success() {
	activities := &demotemporalorderprocessing.OrderActivities{}
	s.env.RegisterActivity(activities.Validate)

	in := demotemporalorderprocessing.Order{
		ID: uuid.MustParse(dummyOrderID),
		LineItems: []demotemporalorderprocessing.LineItem{
			{
				ProductID:    uuid.MustParse("ba320a5d-62ed-46d0-b491-084514598721"),
				Quantity:     1,
				PricePerItem: decimal.RequireFromString("123.45"),
			},
		},
	}

	_, err := s.env.ExecuteActivity(activities.Validate, in)

	s.Require().NoError(err)
}

func (s *ActivityTestSuite) TestValidate_Fail() {

	tests := []struct {
		name  string
		input demotemporalorderprocessing.Order
		err   string
	}{
		{
			name:  "Missing order ID",
			input: demotemporalorderprocessing.Order{},
			err:   "order must have a valid order ID",
		},
		{
			name: "No items in order",
			input: demotemporalorderprocessing.Order{
				ID:        uuid.MustParse(dummyOrderID),
				LineItems: []demotemporalorderprocessing.LineItem{},
			},
			err: "order must have at least one item",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			activities := &demotemporalorderprocessing.OrderActivities{}
			s.env.RegisterActivity(activities.Validate)

			_, err := s.env.ExecuteActivity(activities.Validate, tt.input)

			s.Require().ErrorContains(err, tt.err)
		})
	}
}
