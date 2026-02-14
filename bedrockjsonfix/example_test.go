package bedrockjsonfix_test

import (
	"fmt"

	"github.com/SkaticNET/bedrock-jsonfix/bedrockjsonfix"
)

func ExampleFixBytes_report() {
	opt := bedrockjsonfix.DefaultOptions()
	res, err := bedrockjsonfix.FixBytes([]byte("{//comment\n\"a\":1,}\n"), opt)
	if err != nil {
		panic(err)
	}
	fmt.Printf("line_comments=%d trailing_commas=%d valid=%t\n",
		res.Report.StrippedLineComments,
		res.Report.RemovedTrailingCommas,
		res.Report.ValidJSON,
	)
	// Output: line_comments=1 trailing_commas=1 valid=true
}
