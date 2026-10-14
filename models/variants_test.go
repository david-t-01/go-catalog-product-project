package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestVariant_EffectivePrice(t *testing.T) {
	product := decimal.RequireFromString("10.99")
	tests := []struct {
		name  string
		price decimal.NullDecimal
		want  string
	}{
		{"inherits when null", decimal.NullDecimal{}, "10.99"},
		{"keeps own price", decimal.NewNullDecimal(decimal.RequireFromString("11.99")), "11.99"},
		{"keeps explicit zero", decimal.NewNullDecimal(decimal.Zero), "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Variant{Price: tt.price}
			assert.Equal(t, tt.want, v.EffectivePrice(product).String())
		})
	}
}
