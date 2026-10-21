package services

import (
	"testing"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
)

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
