package output

import (
	"fmt"
	"os"

	"natneam.com/skan/model"
	"natneam.com/skan/utils"
)

func EmitText(data chan model.Match, colorOutput bool) {
	useColor := utils.IsTTY() && colorOutput
	countMatches := 0
	filesCount := make(map[string]bool)

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
		countMatches++
		filesCount[res.FileName] = true
	}
	if countMatches == 0 {
		return
	}
	fmt.Printf("%d Matches across %d files\n", countMatches, len(filesCount))
}

func EmitErrorInfo(errorCount int, errOutputFile string) {
	if errorCount == 0 {
		return
	}
	if errOutputFile != "" {
		fmt.Fprintf(os.Stderr, "Encountered %d errors, written to %s\n", errorCount, errOutputFile)
	} else {
		fmt.Fprintf(os.Stderr, "Encountered %d errors\n", errorCount)
	}
}
