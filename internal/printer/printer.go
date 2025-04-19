package printer

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintGreen(s string) {
	color.Set(color.FgHiGreen)
	fmt.Print(s)
	color.Unset()
}
