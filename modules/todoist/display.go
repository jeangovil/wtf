package todoist

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/wtf"
)

func (widget *Widget) display() {
	proj := widget.CurrentProject()

	if proj == nil {
		return
	}

	title := fmt.Sprintf("[green]%s[white]", proj.Project.Name)
	str := ""

	for idx, item := range proj.tasks {
		row := fmt.Sprintf(
			`[%s]| | %s[%s]`,
			widget.RowColor(idx),
			tview.Escape(item.Content),
			widget.RowColor(idx),
		)

		str += wtf.HighlightableHelper(widget.View, row, idx, len(item.Content))
	}

	widget.ScrollableWidget.Redraw(title, str, false)
}
