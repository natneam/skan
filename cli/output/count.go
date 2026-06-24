package output

import (
	"fmt"

	"natneam.com/skan/model"
)

func EmitCount(data chan model.Match) {
	count := 0
	for range data {
		count++
	}
	fmt.Println(count)
}
