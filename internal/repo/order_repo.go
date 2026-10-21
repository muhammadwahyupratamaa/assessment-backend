package repo

import (
	"context"
	"errors"
	"sort"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrOrderProductNotFound    = errors.New("order product not found")
	ErrInsufficientStock       = errors.New("insufficient product stock")
	ErrInvalidCheckoutItem     = errors.New("invalid checkout item")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
)

type OrderRepo interface {
	Checkout(
		ctx context.Context,
		userID string,
		idempotencyKey string,
		items []models.CheckoutItemRequest,
	) (models.Order, error)
	GetByID(ctx context.Context, orderID string) (models.Order, error)
	ListByUser(
		ctx context.Context,
		userID string,
		params models.OrderListParams,
	) ([]models.Order, int64, error)
	ListAll(
		ctx context.Context,
		params models.OrderListParams,
	) ([]models.Order, int64, error)
	UpdateStatus(
		ctx context.Context,
		orderID string,
		status models.OrderStatus,
	) (models.Order, error)
}

type orderRepo struct {
	db *pgxpool.Pool
}

type checkoutProduct struct {
	ID       string
	Name     string
	Price    int64
	Stock    int
	Quantity int
	Subtotal int64
}

func NewOrderRepo(db *pgxpool.Pool) OrderRepo {
	return &orderRepo{
		db: db,
	}
}

func (r *orderRepo) Checkout(
	ctx context.Context,
	userID string,
	idempotencyKey string,
	items []models.CheckoutItemRequest,
) (models.Order, error) {
	quantities := make(map[string]int)

	for _, item := range items {
		if item.ProductID == "" || item.Quantity < 1 {
			return models.Order{}, ErrInvalidCheckoutItem
		}

		quantities[item.ProductID] += item.Quantity
	}

	productIDs := make([]string, 0, len(quantities))
	for productID := range quantities {
		productIDs = append(productIDs, productID)
	}

	sort.Strings(productIDs)

	if idempotencyKey != "" {
		existingOrder, err := r.getOrderByIdempotencyKey(
			ctx,
			userID,
			idempotencyKey,
		)
		if err == nil {
			return existingOrder, nil
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, err
		}
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Order{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	products := make([]checkoutProduct, 0, len(productIDs))
	var totalAmount int64

	for _, productID := range productIDs {
		var product checkoutProduct

		err := tx.QueryRow(
			ctx,
			`SELECT id, name, price, stock
			FROM products
			WHERE id = $1 AND deleted_at IS NULL
			FOR UPDATE`,
			productID,
		).Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Stock,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrOrderProductNotFound
		}
		if err != nil {
			return models.Order{}, err
		}

		product.Quantity = quantities[productID]

		if product.Stock < product.Quantity {
			return models.Order{}, ErrInsufficientStock
		}

		product.Subtotal = product.Price * int64(product.Quantity)
		totalAmount += product.Subtotal

		products = append(products, product)
	}

	var order models.Order
	var status string

	err = tx.QueryRow(
		ctx,
		`INSERT INTO orders (
			user_id,
			total_amount,
			idempotency_key
		)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, user_id, total_amount, status, created_at`,
		userID,
		totalAmount,
		idempotencyKey,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&status,
		&order.CreatedAt,
	)
	if err != nil {
		if idempotencyKey != "" && isUniqueViolation(err) {
			_ = tx.Rollback(ctx)

			return r.getOrderByIdempotencyKey(
				ctx,
				userID,
				idempotencyKey,
			)
		}

		return models.Order{}, err
	}

	order.Status = models.OrderStatus(status)

	for _, product := range products {
		_, err := tx.Exec(
			ctx,
			`UPDATE products
			SET stock = stock - $1, updated_at = now()
			WHERE id = $2`,
			product.Quantity,
			product.ID,
		)
		if err != nil {
			return models.Order{}, err
		}

		var orderItemID string

		err = tx.QueryRow(
			ctx,
			`INSERT INTO order_items (
				order_id,
				product_id,
				quantity,
				unit_price,
				subtotal
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			order.ID,
			product.ID,
			product.Quantity,
			product.Price,
			product.Subtotal,
		).Scan(&orderItemID)
		if err != nil {
			return models.Order{}, err
		}

		order.Items = append(order.Items, models.OrderItem{
			ID:          orderItemID,
			OrderID:     order.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    product.Quantity,
			Price:       product.Price,
			Subtotal:    product.Subtotal,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *orderRepo) GetByID(
	ctx context.Context,
	orderID string,
) (models.Order, error) {
	order, err := r.getOrder(ctx, r.db, orderID)
	if err != nil {
		return models.Order{}, err
	}

	items, err := r.getOrderItems(ctx, r.db, order.ID)
	if err != nil {
		return models.Order{}, err
	}

	order.Items = items

	return order, nil
}

func (r *orderRepo) ListByUser(
	ctx context.Context,
	userID string,
	params models.OrderListParams,
) ([]models.Order, int64, error) {
	var total int64

	err := r.db.QueryRow(
		ctx,
		`SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		ctx,
		`SELECT id, user_id, total_amount, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID,
		params.Limit,
		(params.Page-1)*params.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders, err := r.scanOrderList(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepo) ListAll(
	ctx context.Context,
	params models.OrderListParams,
) ([]models.Order, int64, error) {
	var total int64

	if err := r.db.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM orders`,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		ctx,
		`SELECT id, user_id, total_amount, status, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		params.Limit,
		(params.Page-1)*params.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders, err := r.scanOrderList(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepo) UpdateStatus(
	ctx context.Context,
	orderID string,
	status models.OrderStatus,
) (models.Order, error) {
	if status != models.OrderStatusPending &&
		status != models.OrderStatusCompleted &&
		status != models.OrderStatusCancelled {
		return models.Order{}, ErrInvalidOrderStatus
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Order{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var currentStatus string

	err = tx.QueryRow(
		ctx,
		`SELECT status
		FROM orders
		WHERE id = $1
		FOR UPDATE`,
		orderID,
	).Scan(&currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return models.Order{}, err
	}

	currentOrderStatus := models.OrderStatus(currentStatus)

	if currentOrderStatus == status {
		if err := tx.Commit(ctx); err != nil {
			return models.Order{}, err
		}

		return r.GetByID(ctx, orderID)
	}

	if currentOrderStatus != models.OrderStatusPending {
		return models.Order{}, ErrInvalidStatusTransition
	}

	if status == models.OrderStatusCancelled {
		rows, err := tx.Query(
			ctx,
			`SELECT product_id, quantity
			FROM order_items
			WHERE order_id = $1`,
			orderID,
		)
		if err != nil {
			return models.Order{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var productID string
			var quantity int

			if err := rows.Scan(&productID, &quantity); err != nil {
				return models.Order{}, err
			}

			_, err := tx.Exec(
				ctx,
				`UPDATE products
				SET stock = stock + $1, updated_at = now()
				WHERE id = $2`,
				quantity,
				productID,
			)
			if err != nil {
				return models.Order{}, err
			}
		}

		if err := rows.Err(); err != nil {
			return models.Order{}, err
		}
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE orders
		SET status = $1
		WHERE id = $2`,
		status,
		orderID,
	)
	if err != nil {
		return models.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Order{}, err
	}

	return r.GetByID(ctx, orderID)
}

func (r *orderRepo) scanOrderList(
	ctx context.Context,
	rows pgx.Rows,
) ([]models.Order, error) {
	orders := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order
		var status string

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&status,
			&order.CreatedAt,
		); err != nil {
			return nil, err
		}

		order.Status = models.OrderStatus(status)

		items, err := r.getOrderItems(ctx, r.db, order.ID)
		if err != nil {
			return nil, err
		}

		order.Items = items
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) getOrder(
	ctx context.Context,
	queryer interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	},
	orderID string,
) (models.Order, error) {
	var order models.Order
	var status string

	err := queryer.QueryRow(
		ctx,
		`SELECT id, user_id, total_amount, status, created_at
		FROM orders
		WHERE id = $1`,
		orderID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&status,
		&order.CreatedAt,
	)
	if err != nil {
		return models.Order{}, err
	}

	order.Status = models.OrderStatus(status)

	return order, nil
}

func (r *orderRepo) getOrderItems(
	ctx context.Context,
	queryer interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	},
	orderID string,
) ([]models.OrderItem, error) {
	rows, err := queryer.Query(
		ctx,
		`SELECT
			oi.id,
			oi.order_id,
			oi.product_id,
			p.name,
			oi.unit_price,
			oi.quantity,
			oi.subtotal
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.id`,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)

	for rows.Next() {
		var item models.OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.Price,
			&item.Quantity,
			&item.Subtotal,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *orderRepo) getOrderByIdempotencyKey(
	ctx context.Context,
	userID string,
	idempotencyKey string,
) (models.Order, error) {
	var orderID string

	err := r.db.QueryRow(
		ctx,
		`SELECT id
		FROM orders
		WHERE user_id = $1 AND idempotency_key = $2`,
		userID,
		idempotencyKey,
	).Scan(&orderID)
	if err != nil {
		return models.Order{}, err
	}

	return r.GetByID(ctx, orderID)
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError

	return errors.As(err, &pgError) && pgError.Code == "23505"
}
