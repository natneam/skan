package output

import (
	"encoding/json"
	"fmt"
	"os"

	"natneam.com/skan/model"
)

func EmitJSON(data chan model.Match) {
	for res := range data {
		s, err := json.Marshal(res)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Println(string(s))
	}
}
