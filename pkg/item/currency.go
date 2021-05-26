package item

import "fmt"

// Currency representa el tipo de dato moneda
type Currency int64

// ToCurrency converts a float64 to Currency
// e.g. 1.23 to $1.23, 1.345 to $1.35
func ToCurrency(f float64) Currency {
	return Currency((f * 100) + 0.5)
}

// Float64 converts a USD to float64
func (c Currency) Float64() float64 {
	x := float64(c)
	x = x / 100
	return x
}

// Multiply safely multiplies a USD value by a float64, rounding
// to the nearest cent.
func (c Currency) Multiply(f float64) Currency {
	x := (float64(c) * f) + 0.5
	return Currency(x)
}

// String returns a formatted Currency value
func (c Currency) String() string {
	x := float64(c)
	x = x / 100
	return fmt.Sprintf("$%.2f", x)
}
