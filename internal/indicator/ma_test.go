package indicator

import (
	"fmt"
	"testing"
)

func TestCalculateMa(t *testing.T) {
	fmt.Println(calculateMa([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 9))
}
