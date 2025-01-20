package console

import (
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"path/filepath"
	"strings"
	"time"
)

func (u *UI) ShowGameOver(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	u.blockInput(3 * time.Second)
	u.animator.CancelAll()
	u.gameIsOver = true
	cview.FadeToBlack(u.application, u.settings.AnimationDelay, 10, false)

	if scoreInfo.Escaped {
		u.showWinScreen(scoreInfo, highScores)
	} else {
		u.showDeathScreen(scoreInfo, highScores)
	}
}

func (u *UI) showWinScreen(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(true)
	textView.SetScrollable(false)
	textView.SetScrollBarVisibility(cview.ScrollBarNever)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)

	winMessage := fxtools.ReadFileAsLines(filepath.Join(u.settings.DataRootDir, "win.txt"))

	gameOverMessage := []string{
		"",
		"",
		"",
		fmt.Sprintf("%s", scoreInfo.PlayerName),
		fmt.Sprintf("Cash: %d", scoreInfo.Gold),
		fmt.Sprintf("%s", scoreInfo.DescriptiveMessage),
		"",
		"",
		"",
	}
	gameOverMessage = append(gameOverMessage, winMessage...)

	pressSpace := []string{
		"",
		"",
		"",
		"",
		fmt.Sprintf("Press [#FFFFFF::b]SPACE[-:-:-] to continue"),
	}
	gameOverMessage = append(gameOverMessage, pressSpace...)
	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	textView.SetInputCapture(u.popOnSpaceWithNotification(panelName, func() {
		u.resetFocusToMain()
		u.showHighscoresAndRestart(highScores)
	}))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.lockFocusToPrimitive(textView)
}
func (u *UI) showDeathScreen(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetBorder(true)
	textView.SetScrollable(false)
	textView.SetScrollBarVisibility(cview.ScrollBarNever)
	textView.SetTitle("You died")

	gameOverMessage := []string{
		"",
		fmt.Sprintf("%s", scoreInfo.PlayerName),
		fmt.Sprintf("Cash: %d", scoreInfo.Gold),
		fmt.Sprintf("Cause of Death: %s", scoreInfo.DescriptiveMessage),
	}
	restartText := []string{
		"",
		"",
		"[#fccc2b::b]Do you want to play again? (y/n)[-:-:-]",
		"",
		"",
	}
	scoreTable := toLinesOfText(highScores)

	gameOverMessage = append(gameOverMessage, restartText...)
	gameOverMessage = append(gameOverMessage, scoreTable...)

	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.lockFocusToPrimitive(textView)
	textView.SetInputCapture(u.yesNoReceiver(u.reset, u.QuitGame))
}

func (u *UI) reset() {
	u.mapOverlay.ClearAll()
	u.pages.RemovePanel("modal")
	u.resetFocusToMain()
	u.gameIsOver = false
	u.game.Reset()
}
func (u *UI) showHighscoresAndRestart(highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetTitle("Game Over")

	restartText := []string{
		"",
		"[#fccc2b::b]Do you want to play again? (y/n)[-:-:-]",
		"",
	}
	scoreTable := toLinesOfText(highScores)
	gameOverMessage := append(restartText, scoreTable...)

	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
	textView.SetInputCapture(u.yesNoReceiver(u.reset, u.QuitGame))
}

func (u *UI) ShowHighScoresOnly(highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetTitle("High Scores")

	scoreTable := toLinesOfText(highScores)
	u.setColoredText(textView, strings.Join(scoreTable, "\n"))

	panelName := "main"
	if u.pages.HasPanel("main") {
		panelName = "fullscreen"
	}

	textView.SetInputCapture(u.popOnAnyKeyWithNotification(panelName, u.QuitGame))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
}

func (u *UI) QuitGame() {
	u.lifeCycle.QuitGame(u.application)
}
