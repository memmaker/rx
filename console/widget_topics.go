package console

import (
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"os"
	"path"
	"strings"
)

type TextTopic struct {
	Title string
	Text  string
}

func (u *UI) openDirectoryAsTopics(dirname string) {
	directoryName := path.Join(u.settings.DataRootDir, dirname)
	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return
	}

	var topics []TextTopic
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := path.Join(directoryName, entry.Name())
		lines := fxtools.ReadFileAsLines(fileName)

		topicName := path.Base(entry.Name())
		if len(lines) > 0 && strings.HasPrefix(lines[0], "%title:") {
			topicName = strings.TrimPrefix(lines[0], "%title:")
			lines = lines[1:]
		}

		topics = append(topics, TextTopic{
			Title: strings.TrimSpace(topicName),
			Text:  strings.Join(lines, "\n"),
		})
	}

	u.openTopicViewer(topics)
}

func (u *UI) openTopicViewer(data []TextTopic) {

	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	borderFgFocus := u.uiTheme.GetUIColorForTcell(UIColorBorderForegroundFocus)

	topicViewer := NewTopicViewer(data, fg, bg, borderFgFocus, func() {
		u.pages.RemovePanel("Topics")
		u.popFocus()
	})
	u.pages.AddPanel("Topics", topicViewer, true, true)
	u.lockFocusToPrimitive(topicViewer)
}

type TopicViewer struct {
	*cview.Grid
	onClose func()
	topics  *cview.List
	text    *cview.TextView
}

func (b TopicViewer) SetOnClose(onClose func()) {
	b.onClose = onClose
}

func NewTopicViewer(data []TextTopic, fg, bg, borderFgFocus tcell.Color, close func()) TopicViewer {
	closeOnEscape := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		return event
	}

	grid := cview.NewGrid()

	grid.SetInputCapture(closeOnEscape)

	widthNeeded := longestTitle(data) + 2
	grid.SetColumns(widthNeeded, 0)
	grid.SetRows(0, 1)

	helpText := cview.NewTextView()
	helpText.SetText("↑/↓/Home/End/PgDn/PgUp to navigate, ESC to close")
	helpText.SetTextColor(fg)
	helpText.SetBackgroundColor(bg)
	helpText.SetDynamicColors(true)
	helpText.SetWrap(false)
	helpText.SetBorder(false)
	helpText.SetTextAlign(cview.AlignCenter)

	textView := cview.NewTextView()
	textView.SetBorder(true)
	textView.SetDynamicColors(true)
	textView.SetTitleColor(fg)
	textView.SetTextColor(fg)
	textView.SetBorderColor(fg)
	textView.SetScrollBarColor(fg)
	textView.SetBackgroundColor(bg)
	textView.SetWrap(true)
	textView.SetWordWrap(true)

	topicList := cview.NewList()

	topicList.SetBorder(true)
	topicList.SetWrapAround(true)
	topicList.SetHover(true)
	topicList.ShowSecondaryText(false)
	topicList.SetForceSelectedTextColor(true)
	topicList.SetChangedFunc(func(index int, listItem *cview.ListItem) {
		item := data[index]
		textView.SetText(item.Text)
	})
	topicList.SetScrollBarColor(fg)
	//topicList.SetHighlightFullLine(true)

	topicList.SetTitleColor(fg)
	topicList.SetMainTextColor(fg)
	topicList.SetSecondaryTextColor(fg)

	topicList.SetBorderColor(fg)
	topicList.SetBorderColorFocused(borderFgFocus)

	topicList.SetBackgroundColor(bg)

	topicList.SetShortcutColor(fg)

	topicList.SetSelectedTextColor(bg)
	topicList.SetSelectedBackgroundColor(fg)
	topicList.SetHover(false)
	topicList.SetHighlightDisabled(false)
	topicList.SetHighlightFullLine(true)
	topicList.SetSelectedAlwaysVisible(true)

	//topicList.SetIndicators(">", "", " ", "")
	for _, topic := range data {
		item := cview.NewListItem(topic.Title)
		item.SetSelectedFunc(func() {
			textView.SetText(topic.Text)
		})
		topicList.AddItem(item)
	}

	topicList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		if event.Key() == tcell.KeyPgDn || event.Key() == tcell.KeyPgUp {
			textView.InputHandler()(event, nil)
			return nil
		}
		if event.Key() == tcell.KeyHome {
			textView.InputHandler()(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil)
			return nil
		}
		if event.Key() == tcell.KeyEnd {
			textView.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
			return nil
		}
		return event
	})

	grid.AddItem(topicList, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(textView, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(helpText, 1, 0, 1, 2, 0, 0, false)
	p := TopicViewer{
		Grid:   grid,
		topics: topicList,
		text:   textView,
	}

	//grid.SetDrawFunc(p.drawInside)
	//grid.SetBackgroundTransparent(true)
	//grid.SetInputCapture(p.handleKey)

	return p
}

func longestTitle(data []TextTopic) int {
	longest := 0
	for _, topic := range data {
		if cview.TaggedStringWidth(topic.Title) > longest {
			longest = cview.TaggedStringWidth(topic.Title)
		}
	}
	return longest
}
