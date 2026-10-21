package services

import (
	"context"
	"errors"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"

	"github.com/jackc/pgx/v5"
)

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrOrderProductNotFound   = errors.New("order product not found")
	ErrOrderInsufficientStock = errors.New("insufficient product stock")
	ErrInvalidCheckoutItem    = errors.New("invalid checkout item")
)

type OrderService interface {
	Checkout(
		ctx context.Context,
		userID string,
		idempotencyKey string,
		request models.CheckoutRequest,
	) (models.Order, error)
	GetMyOrder(
		ctx context.Context,
		userID string,
		orderID string,
	) (models.Order, error)
	ListMyOrders(
		ctx context.Context,
		userID string,
		params models.OrderListParams,
	) ([]models.Order, models.OrderPaginationMeta, error)
	ListAllOrders(
		ctx context.Context,
		params models.OrderListParams,
	) ([]models.Order, models.OrderPaginationMeta, error)
}

type orderService struct {
	orderRepo repo.OrderRepo
}

func NewOrderService(orderRepo repo.OrderRepo) OrderService {
	return &orderService{
		orderRepo: orderRepo,
	}
}

func (s *orderService) Checkout(
	ctx context.Context,
	userID string,
	idempotencyKey string,
	request models.CheckoutRequest,
) (models.Order, error) {
	order, err := s.orderRepo.Checkout(
		ctx,
		userID,
		idempotencyKey,
		request.Items,
	)
	if err == nil {
		return order, nil
	}

	switch {
	case errors.Is(err, repo.ErrOrderProductNotFound):
		return models.Order{}, ErrOrderProductNotFound
	case errors.Is(err, repo.ErrInsufficientStock):
		return models.Order{}, ErrOrderInsufficientStock
	case errors.Is(err, repo.ErrInvalidCheckoutItem):
		return models.Order{}, ErrInvalidCheckoutItem
	default:
		return models.Order{}, err
	}
}

func (s *orderService) GetMyOrder(
	ctx context.Context,
	userID string,
	orderID string,
) (models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return models.Order{}, err
	}

	// Customer hanya boleh melihat order miliknya sendiri.
	if order.UserID != userID {
		return models.Order{}, ErrOrderNotFound
	}

	return order, nil
}

func (s *orderService) ListMyOrders(
	ctx context.Context,
	userID string,
	params models.OrderListParams,
) ([]models.Order, models.OrderPaginationMeta, error) {
	params = normalizeOrderListParams(params)

	orders, total, err := s.orderRepo.ListByUser(ctx, userID, params)
	if err != nil {
		return nil, models.OrderPaginationMeta{}, err
	}

	return orders, buildOrderPaginationMeta(params, total), nil
}

func (s *orderService) ListAllOrders(
	ctx context.Context,
	params models.OrderListParams,
) ([]models.Order, models.OrderPaginationMeta, error) {
	params = normalizeOrderListParams(params)

	orders, total, err := s.orderRepo.ListAll(ctx, params)
	if err != nil {
		return nil, models.OrderPaginationMeta{}, err
	}

	return orders, buildOrderPaginationMeta(params, total), nil
}

func normalizeOrderListParams(
	params models.OrderListParams,
) models.OrderListParams {
	if params.Page < 1 {
		params.Page = 1
	}

	if params.Limit < 1 {
		params.Limit = 10
	}

	if params.Limit > 100 {
		params.Limit = 100
	}

	return params
}

func buildOrderPaginationMeta(
	params models.OrderListParams,
	total int64,
) models.OrderPaginationMeta {
	totalPages := int(
		(total + int64(params.Limit) - 1) / int64(params.Limit),
	)

	return models.OrderPaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
