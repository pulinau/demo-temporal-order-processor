package temporal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"
	temporalmocks "github.com/pulinau/demo-temporal-order-processor/internal/temporal/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
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
	// Setup
	inventoryChecker := temporalmocks.NewMockInventoryChecker(s.T())
	inventoryChecker.EXPECT().
		CheckInventory(mock.Anything, uuid.MustParse("ba320a5d-62ed-46d0-b491-084514598721"), int32(1)).
		Return(true, nil)

	activities := temporal.NewOrderActivities(inventoryChecker)
	s.env.RegisterActivity(activities.Validate)

	// Invoke
	_, err := s.env.ExecuteActivity(activities.Validate, temporal.Order{
		ID: uuid.MustParse(dummyOrderID),
		LineItems: []temporal.LineItem{
			{
				ProductID:    uuid.MustParse("ba320a5d-62ed-46d0-b491-084514598721"),
				Quantity:     1,
				PricePerItem: decimal.RequireFromString("123.45"),
			},
		},
	})

	// Assert
	s.Require().NoError(err)
}

func (s *ActivityTestSuite) TestValidate_Fail() {

	tests := []struct {
		name       string
		input      temporal.Order
		setupMocks func(t *testing.T, mockIC *temporalmocks.MockInventoryChecker)
		err        string
	}{
		{
			name:  "Missing order ID",
			input: temporal.Order{},
			err:   "order must have a valid order ID",
		},
		{
			name: "No items in order",
			input: temporal.Order{
				ID:        uuid.MustParse(dummyOrderID),
				LineItems: []temporal.LineItem{},
			},
			err: "order must have at least one item",
		},
		{
			name: "No inventory for line item",
			input: temporal.Order{
				ID: uuid.MustParse(dummyOrderID),
				LineItems: []temporal.LineItem{
					{
						ProductID:    uuid.MustParse("ba320a5d-62ed-46d0-b491-084514598721"),
						Quantity:     1,
						PricePerItem: decimal.RequireFromString("123.45"),
					},
				},
			},
			setupMocks: func(t *testing.T, mockIC *temporalmocks.MockInventoryChecker) {
				mockIC.EXPECT().CheckInventory(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
			},
			err: "insufficient inventory for product",
		},
		{
			name: "Inventory checker error",
			input: temporal.Order{
				ID: uuid.MustParse(dummyOrderID),
				LineItems: []temporal.LineItem{
					{
						ProductID:    uuid.MustParse("ba320a5d-62ed-46d0-b491-084514598721"),
						Quantity:     1,
						PricePerItem: decimal.RequireFromString("123.45"),
					},
				},
			},
			setupMocks: func(t *testing.T, mockIC *temporalmocks.MockInventoryChecker) {
				mockIC.EXPECT().CheckInventory(mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("test error"))
			},
			err: "failed to check inventory for product",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup
			inventoryChecker := temporalmocks.NewMockInventoryChecker(s.T())
			if tt.setupMocks != nil {
				tt.setupMocks(s.T(), inventoryChecker)
			}

			activities := temporal.NewOrderActivities(inventoryChecker)
			s.env.RegisterActivity(activities.Validate)

			// Invoke
			_, err := s.env.ExecuteActivity(activities.Validate, tt.input)

			// Assert
			s.Require().ErrorContains(err, tt.err)
		})
	}
}
