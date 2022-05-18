// Referenced code at https://go.dev/play/p/v5mP-_KEce
package currency

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// The base unit is in 1/1000 of a cent.
const (
	cents   = 1000
	dollars = 100 * cents
)

type MicroUSD int64

func ParseMicroUSD(amount string) (MicroUSD, error) {
	if matched, _ := regexp.MatchString(`^(0|[1-9][0-9]*).[0-9]{2}$`, amount); !matched {
		return 0, errors.New("invalid syntax")
	}

	f, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, err
	}

	if f < 0.00 || f > 4294967295.99 {
		return 0, errors.New("amount out of bounds")
	}

	return MicroUSD(f * dollars), nil
}

func (musd MicroUSD) String() string {
	return fmt.Sprintf("%.2f", float64(musd)/dollars)
}
