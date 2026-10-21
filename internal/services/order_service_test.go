package services

import (
	"context"
	"errors"
	"testing"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"
)

type orderRepoStub struct {
	updateStatusOrder models.Order
	updateStatusErr   error
}

func (s orderRepoStub) Checkout(
	context.Context,
	string,
	string,
	[]models.CheckoutItemRequest,
) (models.Order, error) {
	return models.Order{}, nil
}

func (s orderRepoStub) GetByID(
	context.Context,
	string,
) (models.Order, error) {
	return models.Order{}, nil
}

func (s orderRepoStub) ListByUser(
	context.Context,
	string,
	models.OrderListParams,
) ([]models.Order, int64, error) {
	return nil, 0, nil
}

func (s orderRepoStub) ListAll(
	context.Context,
	models.OrderListParams,
) ([]models.Order, int64, error) {
	return nil, 0, nil
}

func (s orderRepoStub) UpdateStatus(
	context.Context,
	string,
	models.OrderStatus,
) (models.Order, error) {
	return s.updateStatusOrder, s.updateStatusErr
}

func TestNormalizeOrderListParams(t *testing.T) {
	testCases := []struct {
		name  string
		input models.OrderListParams
		want  models.OrderListParams
	}{
		{
			name:  "use default values",
			input: models.OrderListParams{},
			want: models.OrderListParams{
				Page:  1,
				Limit: 10,
			},
		},
		{
			name: "limit maximum is capped",
			input: models.OrderListParams{
				Page:  2,
				Limit: 200,
			},
			want: models.OrderListParams{
				Page:  2,
				Limit: 100,
			},
		},
		{
			name: "valid values are preserved",
			input: models.OrderListParams{
				Page:  3,
				Limit: 25,
			},
			want: models.OrderListParams{
				Page:  3,
				Limit: 25,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := normalizeOrderListParams(testCase.input)

			if got != testCase.want {
				t.Fatalf(
					"normalizeOrderListParams() = %+v, want %+v",
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestBuildOrderPaginationMeta(t *testing.T) {
	params := models.OrderListParams{
		Page:  2,
		Limit: 10,
	}

	got := buildOrderPaginationMeta(params, 25)

	want := models.OrderPaginationMeta{
		Page:       2,
		Limit:      10,
		Total:      25,
		TotalPages: 3,
	}

	if got != want {
		t.Fatalf(
			"buildOrderPaginationMeta() = %+v, want %+v",
			got,
			want,
		)
	}
}

func TestUpdateStatusMapsRepositoryErrors(t *testing.T) {
	testCases := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{
			name:    "order not found",
			repoErr: repo.ErrOrderNotFound,
			wantErr: ErrOrderNotFound,
		},
		{
			name:    "invalid status",
			repoErr: repo.ErrInvalidOrderStatus,
			wantErr: ErrInvalidOrderStatus,
		},
		{
			name:    "invalid status transition",
			repoErr: repo.ErrInvalidStatusTransition,
			wantErr: ErrInvalidStatusTransition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewOrderService(orderRepoStub{
				updateStatusErr: testCase.repoErr,
			})

			_, err := service.UpdateStatus(
				context.Background(),
				"order-id",
				models.UpdateOrderStatusRequest{
					Status: models.OrderStatusCompleted,
				},
			)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf(
					"UpdateStatus() error = %v, want %v",
					err,
					testCase.wantErr,
				)
			}
		})
	}
}

func TestUpdateStatusReturnsUpdatedOrder(t *testing.T) {
	want := models.Order{
		ID:     "order-id",
		Status: models.OrderStatusCompleted,
	}

	service := NewOrderService(orderRepoStub{
		updateStatusOrder: want,
	})

	got, err := service.UpdateStatus(
		context.Background(),
		"order-id",
		models.UpdateOrderStatusRequest{
			Status: models.OrderStatusCompleted,
		},
	)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if got.ID != want.ID || got.Status != want.Status {
		t.Fatalf("UpdateStatus() = %+v, want %+v", got, want)
	}
}