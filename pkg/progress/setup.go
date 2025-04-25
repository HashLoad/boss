package progress

import (
	"github.com/pterm/pterm"
	"github.com/snakeice/gogress/format"
)

func Setup() {
	format.DefaultFormat.BoxStart = "|"
	format.DefaultFormat.BoxEnd = "|"
	format.DefaultFormat.Empty = pterm.Gray("█")
	format.DefaultFormat.Current = pterm.Normal("█")
	format.DefaultFormat.Completed = pterm.Cyan("█")
}
