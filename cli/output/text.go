package output

import (
	"fmt"

	"natneam.com/skan/model"
	"natneam.com/skan/utils"
)

func EmitText(data chan model.Match, colorOutput bool) {
	useColor := utils.IsTTY() && colorOutput

	for res := range data {
		for _, bC := range res.BeforeContext {
			fmt.Printf("%s-%d-%s\n", bC.FileName, bC.LineNumber, bC.LineText)
		}

		line := res.LineText
		if useColor {
			line = utils.HighlightLine(res.LineText, res.MatchIndexes)
		}

		fmt.Printf("%s:%d:%s\n", res.FileName, res.LineNumber, line)
		for _, aC := range res.AfterContext {
			fmt.Printf("%s-%d-%s\n", aC.FileName, aC.LineNumber, aC.LineText)
		}
		fmt.Println("---")
	}
}
