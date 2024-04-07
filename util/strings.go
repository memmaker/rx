package util

import (
	"RogueUI/geometry"
	"fmt"
	"strings"
)

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

func MaxLen(text []string) int {
	maxLength := 0
	for _, line := range text {
		asRunes := []rune(line)
		if len(asRunes) > maxLength {
			maxLength = len(asRunes)
		}
	}
	return maxLength
}
func RightPad(s string, pLen int) string {
	return s + strings.Repeat(" ", pLen-len(s))
}

func LeftPad(s string, pLen int) string {
	return strings.Repeat(" ", pLen-len(s)) + s
}
func RightPadCount(s string, count int) string {
	return s + strings.Repeat(" ", count)
}

func LeftPadCount(s string, count int) string {
	return strings.Repeat(" ", count) + s
}

func RightPadCountMulti(lines []string, count int) []string {
	for i, line := range lines {
		lines[i] = RightPadCount(line, count)
	}
	return lines
}

func LeftPadCountMulti(lines []string, count int) []string {
	for i, line := range lines {
		lines[i] = LeftPadCount(line, count)
	}
	return lines
}

func rightAlignColumns(cols []string, width []int) string {
	for i, col := range cols {
		if len(col) < width[i] {
			cols[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", width[i]-len(col)), col)
		}
	}
	return strings.Join(cols, "")
}

type TableRow struct {
	Columns []string
}

func TableLine(labelWidth, colWidth int, label string, columns ...string) string {
	colWidths := make([]int, len(columns))
	for i, _ := range columns {
		colWidths[i] = colWidth
	}
	return fmt.Sprintf("%s%s", RightPad(label, labelWidth+1), rightAlignColumns(columns, colWidths))
}
func TableLayout(tableData []TableRow, alignments []TextAlignment) []string {
	colWidths := make([]int, len(tableData[0].Columns))

	for _, row := range tableData {
		for i, col := range row.Columns {
			if len(col)+1 > colWidths[i] {
				colWidths[i] = len(col) + 1
			}
		}
	}

	var result []string
	for _, row := range tableData {
		var rowText string
		for i, col := range row.Columns {
			var paddedCol string
			if alignments[i] == AlignRight {
				paddedCol = LeftPad(col, colWidths[i])
			} else if alignments[i] == AlignCenter {
				paddedCol = CenterPad(col, colWidths[i])
			} else {
				paddedCol = RightPad(col, colWidths[i])
			}
			rowText += paddedCol
		}
		result = append(result, rowText)
	}
	return result
}

func CenterPad(col string, neededWidth int) string {
	if len(col) >= neededWidth {
		return col
	}
	pad := strings.Repeat(" ", (neededWidth-len(col))/2)
	return fmt.Sprintf("%s%s%s", pad, col, pad)
}

func AutoLayoutArray(inputText []string, width int) []string {
	return AutoLayout(strings.Join(inputText, " "), width)
}
func splitIntoTokens(inputText []string) []string {
	oneString := strings.Join(inputText, " ")
	return strings.Split(oneString, " ")
}

// AutoLayout splits a string into lines of maximum width
func AutoLayout(inputText string, width int) []string {
	return AutoLayoutWithBreakingPrefix(inputText, width, "")
}

func AutoLayoutWithBreakingPrefix(inputText string, width int, prefix string) []string {
	// split on spaces and newlines
	inputText = strings.ReplaceAll(inputText, "\n", " ")
	tokens := strings.Split(inputText, " ")

	var lines []string
	currentLine := ""
	for i, token := range tokens {
		if len(currentLine)+len(token)+1 > width {
			lines = append(lines, currentLine)
			currentLine = prefix + strings.TrimSpace(token)
		} else if i == 0 {
			currentLine = prefix + strings.TrimSpace(token)
		} else {
			currentLine += " " + strings.TrimSpace(token)
		}
	}
	lines = append(lines, currentLine)
	return lines
}
func AutoLayoutPages(inputText string, width int, height int) [][]string {
	inputText = strings.ReplaceAll(inputText, "\n", " ")
	tokens := strings.Split(inputText, " ")
	var allPages [][]string
	var currentPage []string
	lastDelim := geometry.Point{X: -1, Y: -1}
	currentLine := ""
	firstTokenInPage := true
	for len(tokens) > 0 {
		// pop
		token := tokens[0]
		tokens = tokens[1:]

		if strings.TrimSpace(token) == "" {
			continue
		}

		indexOfDelimInToken := strings.IndexAny(token, ".!?")
		indexOfDelimInLine := -1
		if len(currentLine)+len(token)+1 > width {
			currentPage = append(currentPage, currentLine)
			if len(currentPage) == height {
				if lastDelim.X > 0 {
					// we have a sensible split?
					lineToSplitOn := currentPage[lastDelim.Y]

					splitIndex := lastDelim.X + 1
					if len(lineToSplitOn) > splitIndex && lineToSplitOn[splitIndex] == '"' {
						splitIndex++
					}
					// split the lines
					partOfCurrPage := strings.TrimSpace(lineToSplitOn[:splitIndex])
					partOfNextPage := strings.TrimSpace(lineToSplitOn[splitIndex:])

					// get all the lines up to here
					currPageLines := currentPage[:lastDelim.Y]
					// append the split part
					currPageLines = append(currPageLines, partOfCurrPage)

					// get the rest of the lines
					nextPageLines := currentPage[lastDelim.Y+1:]

					// prepend the split part
					nextPageLines = append([]string{partOfNextPage}, nextPageLines...)
					restOfTheTokens := splitIntoTokens(nextPageLines)

					// re-add the token we popped
					restOfTheTokens = append(restOfTheTokens, token)

					// prepend to our token queue
					tokens = append(restOfTheTokens, tokens...)

					// append the current page
					allPages = append(allPages, currPageLines)

					// reset
					currentPage = []string{}
					lastDelim = geometry.Point{X: -1, Y: -1}
					currentLine = ""
					firstTokenInPage = true
					continue
				} else {
					// just split
					allPages = append(allPages, currentPage)
					currentPage = []string{}
				}
			}
			currentLine = token
			if indexOfDelimInToken != -1 {
				indexOfDelimInLine = indexOfDelimInToken
			}
		} else if firstTokenInPage {
			firstTokenInPage = false
			currentLine = token
			if indexOfDelimInToken != -1 {
				indexOfDelimInLine = indexOfDelimInToken
			}
		} else {
			currentLine += " " + token
			if indexOfDelimInToken != -1 {
				indexOfDelimInLine = len(currentLine) - len(token) + indexOfDelimInToken
			}
		}
		indexOfCurrentLine := len(currentPage)
		if indexOfDelimInLine != -1 {
			lastDelim = geometry.Point{X: indexOfDelimInLine, Y: indexOfCurrentLine}
		}
	}
	currentPage = append(currentPage, currentLine)
	allPages = append(allPages, currentPage)
	return allPages
}
