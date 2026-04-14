package app

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/pool"
	"github.com/Gaurav-Gosain/tuios/internal/terminal"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

var (
	baseButtonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
)

func getBorder() lipgloss.Border {
	return config.GetBorderForStyle()
}

func getNormalBorder() lipgloss.Border {
	return getBorder()
}

// RightString returns a right-aligned string on the top border line.
// With side borders (BorderStyle != "none") it wraps the line in ╭/╮ corner characters.
// windowBg optionally colors the left padding to match the window body when non-nil.
func RightString(str string, width int, color color.Color, windowBg ...color.Color) string {
	spaces := width - lipgloss.Width(str)
	style := pool.GetStyle()
	defer pool.PutStyle(style)
	fg := style.Foreground(color)

	if spaces < 0 {
		return ""
	}

	padStyle := fg
	if len(windowBg) > 0 && windowBg[0] != nil {
		padStyle = fg.Background(windowBg[0])
	}
	if config.HasSideBorders() {
		return padStyle.Render(config.GetWindowBorderTopLeft()+strings.Repeat(config.GetWindowBorderTop(), spaces)) +
			str +
			fg.Render(config.GetWindowBorderTopRight())
	}
	return padStyle.Render(strings.Repeat(config.GetWindowBorderTop(), spaces)) + str
}

func makeRounded(content string, fgColor color.Color, bgColor ...color.Color) string {
	style := pool.GetStyle()
	defer pool.PutStyle(style)
	s := style.Foreground(fgColor)
	if len(bgColor) > 0 && bgColor[0] != nil {
		s = s.Background(bgColor[0])
	}
	content = s.Render(config.GetWindowPillLeft()) + content + s.Render(config.GetWindowPillRight())
	return content
}

// isDefaultTitle checks if the title is the auto-generated default (e.g., "Terminal 8bf1c038").
func isDefaultTitle(title, windowID string) bool {
	if len(windowID) < 8 {
		return false
	}
	return title == "Terminal "+windowID[:8]
}

// userAtHostPrefix matches the common shell prompt prefix "user@host:" or "user@server: " at the start of a title.
var userAtHostPrefix = regexp.MustCompile(`^[^@\s]+@[^\s:]+:\s*`)

// stripUserAtHostPrefix removes "user@host:" from the start of a window title (e.g. "user@server: ~" -> "~").
func stripUserAtHostPrefix(title string) string {
	return strings.TrimSpace(userAtHostPrefix.ReplaceAllString(title, ""))
}

// getWindowTitle returns the display name for a window, truncated to fit within maxWidth.
// Returns empty string if title should be hidden or doesn't fit.
func getWindowTitle(window *terminal.Window, isRenaming bool, renameBuffer string, maxWidth int) string {
	windowName := ""
	if window.CustomName != "" {
		windowName = window.CustomName
	} else if window.Title != "" && !isDefaultTitle(window.Title, window.ID) {
		// Only show terminal-set title if it's not the default "Terminal <id>" format.
		// Strip common "user@host:" prefix from shell prompt titles.
		windowName = stripUserAtHostPrefix(window.Title)
	}

	if isRenaming {
		windowName = renameBuffer + "_"
	}

	if windowName == "" {
		return ""
	}

	maxNameLen := max(maxWidth-6, 0)
	nameWidth := ansi.StringWidth(windowName)
	if nameWidth > maxNameLen {
		if maxNameLen > 3 {
			// Truncate by runes to handle unicode properly
			runes := []rune(windowName)
			truncated := string(runes)
			for ansi.StringWidth(truncated) > maxNameLen-3 && len(runes) > 0 {
				runes = runes[:len(runes)-1]
				truncated = string(runes)
			}
			windowName = truncated + "..."
		} else {
			return ""
		}
	}
	return windowName
}

// renderTitleWithButtons renders a title badge on the left with buttons on the right of a border line.
// windowBg colors the gap between title and buttons when non-nil (matches window body background).
// titleFg is the text color for the window title.
func renderTitleWithButtons(windowName string, buttons string, width int, color color.Color, titleFg color.Color, isTop bool, windowBg color.Color) string {
	style := pool.GetStyle()
	defer pool.PutStyle(style)
	borderStyle := style.Foreground(color)
	nameStyle := style.Foreground(titleFg).Background(color)

	var borderLeft, borderChar, borderRight string
	if isTop {
		borderLeft = config.GetWindowBorderTopLeft()
		borderChar = config.GetWindowBorderTop()
		borderRight = config.GetWindowBorderTopRight()
	} else {
		borderLeft = config.GetWindowBorderBottomLeft()
		borderChar = config.GetWindowBorderBottom()
		borderRight = config.GetWindowBorderBottomRight()
	}

	leftCircle := borderStyle.Render(config.GetWindowPillLeft())
	nameText := nameStyle.Render(" " + windowName + " ")
	rightCircle := borderStyle.Render(config.GetWindowPillRight())
	nameBadge := leftCircle + nameText + rightCircle

	nameBadgeWidth := lipgloss.Width(nameBadge)
	buttonsWidth := lipgloss.Width(buttons)

	middlePadding := width - nameBadgeWidth - buttonsWidth
	if middlePadding < 0 {
		return RightString(buttons, width, color, windowBg)
	}

	gapStyle := borderStyle
	if windowBg != nil {
		gapStyle = borderStyle.Background(windowBg)
	}

	if config.HasSideBorders() {
		return borderStyle.Render(borderLeft) +
			nameBadge +
			gapStyle.Render(strings.Repeat(borderChar, middlePadding)) +
			buttons +
			borderStyle.Render(borderRight)
	}
	return nameBadge +
		gapStyle.Render(strings.Repeat(borderChar, middlePadding)) +
		buttons
}

// renderTitleBadge renders a centered title badge on a border line.
func renderTitleBadge(windowName string, width int, color color.Color, isTop bool) string {
	style := pool.GetStyle()
	defer pool.PutStyle(style)
	borderStyle := style.Foreground(color)
	nameStyle := baseButtonStyle.Background(color)

	var borderLeft, borderChar, borderRight string
	if isTop {
		borderLeft = config.GetWindowBorderTopLeft()
		borderChar = config.GetWindowBorderTop()
		borderRight = config.GetWindowBorderTopRight()
	} else {
		borderLeft = config.GetWindowBorderBottomLeft()
		borderChar = config.GetWindowBorderBottom()
		borderRight = config.GetWindowBorderBottomRight()
	}

	if windowName == "" {
		return borderStyle.Render(borderLeft + strings.Repeat(borderChar, width) + borderRight)
	}

	leftCircle := borderStyle.Render(config.GetWindowPillLeft())
	nameText := nameStyle.Render(" " + windowName + " ")
	rightCircle := borderStyle.Render(config.GetWindowPillRight())
	nameBadge := leftCircle + nameText + rightCircle

	badgeWidth := lipgloss.Width(nameBadge)
	totalPadding := width - badgeWidth

	if totalPadding < 0 {
		return borderStyle.Render(borderLeft + strings.Repeat(borderChar, width) + borderRight)
	}

	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	return borderStyle.Render(borderLeft+strings.Repeat(borderChar, leftPadding)) +
		nameBadge +
		borderStyle.Render(strings.Repeat(borderChar, rightPadding)+borderRight)
}
func addToBorder(content string, borderColor color.Color, titleFg color.Color, window *terminal.Window, isRenaming bool, renameBuffer string, isTiling bool) string {
	contentWidth := lipgloss.Width(content)
	width := max(contentWidth-config.ContentBorderWidth(), 0)
	titlePos := config.WindowTitlePosition

	style := pool.GetStyle()
	defer pool.PutStyle(style)

	// Window body background for coloring the title-bar gap and button pill (nil = use default)
	var windowBg color.Color
	if window.Terminal != nil {
		windowBg = window.Terminal.BackgroundColor()
	}

	// Build window buttons first so we know their width
	var buttons string
	var buttonsWidth int
	if config.HideWindowButtons {
		buttons = ""
		buttonsWidth = 0
	} else {
		buttonStyle := style.Foreground(titleFg).Background(borderColor)
		cross := buttonStyle.Render(config.GetWindowButtonClose())
		dash := buttonStyle.Render("  - ")

		if isTiling {
			buttons = makeRounded(dash+cross, borderColor, windowBg)
		} else {
			square := buttonStyle.Render(" □ ")
			buttons = makeRounded(dash+square+cross, borderColor, windowBg)
		}
		buttonsWidth = lipgloss.Width(buttons)
	}

	// Calculate available width for title based on position.
	// ContentBorderWidth() accounts for corner chars (╭/╮) that are present in side-border mode.
	var titleMaxWidth int
	if titlePos == "top" {
		titleMaxWidth = width - buttonsWidth - config.ContentBorderWidth()
	} else {
		titleMaxWidth = width
	}

	windowName := ""
	if titlePos != "hidden" {
		windowName = getWindowTitle(window, isRenaming, renameBuffer, titleMaxWidth)
	}

	borderStyle := style.Foreground(borderColor)

	// Build top border: resize handle (U+2921) on the left, then title/buttons
	resizeHandle := borderStyle.Render(config.GetWindowResizeHandle())
	var topBorder string
	if titlePos == "top" && windowName != "" {
		// Title on top with buttons on the right
		topBorder = resizeHandle + renderTitleWithButtons(windowName, buttons, width-lipgloss.Width(resizeHandle), borderColor, titleFg, true, windowBg)
	} else {
		// Normal top border with buttons on right
		topBorder = resizeHandle + RightString(buttons, width-lipgloss.Width(resizeHandle), borderColor, windowBg)
	}

	if !config.HasSideBorders() {
		return topBorder + "\n" + content
	}

	var bottomBorder string
	scrollIndicator := ""
	// Show scroll position when in copy mode with scroll offset
	if window.CopyMode != nil && window.CopyMode.Active && window.CopyMode.ScrollOffset > 0 {
		scrollbackLen := 0
		if window.Terminal != nil {
			scrollbackLen = window.Terminal.ScrollbackLen()
		}
		if scrollbackLen > 0 {
			scrollIndicator = fmt.Sprintf(" %d/%d ", window.CopyMode.ScrollOffset, scrollbackLen)
		}
	}

	if titlePos == "bottom" && windowName != "" {
		bottomBorder = renderTitleBadge(windowName, width, borderColor, false)
	} else if scrollIndicator != "" {
		indicatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true)
		indicator := indicatorStyle.Render(scrollIndicator)
		indicatorWidth := lipgloss.Width(indicator)
		lineWidth := max(width-indicatorWidth, 0)
		bottomBorder = borderStyle.Render(config.GetWindowBorderBottomLeft()+strings.Repeat(config.GetWindowBorderBottom(), lineWidth)) + indicator + borderStyle.Render(config.GetWindowBorderBottomRight())
	} else {
		bottomBorder = borderStyle.Render(config.GetWindowBorderBottomLeft() + strings.Repeat(config.GetWindowBorderBottom(), width) + config.GetWindowBorderBottomRight())
	}

	lines := strings.Split(content, "\n")

	if len(lines) > 0 {
		lines[len(lines)-1] = bottomBorder
	}
	return topBorder + "\n" + strings.Join(lines, "\n")
}

func styleToANSI(s lipgloss.Style) (prefix string, suffix string) {
	var te ansi.Style

	fg := s.GetForeground()
	bg := s.GetBackground()

	if _, ok := fg.(lipgloss.NoColor); !ok && fg != nil {
		te = te.ForegroundColor(ansi.Color(fg))
	}
	if _, ok := bg.(lipgloss.NoColor); !ok && bg != nil {
		te = te.BackgroundColor(ansi.Color(bg))
	}

	if s.GetBold() {
		te = te.Bold()
	}
	if s.GetItalic() {
		te = te.Italic(true)
	}
	if s.GetUnderline() {
		te = te.Underline(true)
	}
	if s.GetStrikethrough() {
		te = te.Strikethrough(true)
	}
	if s.GetBlink() {
		te = te.Blink(true)
	}
	if s.GetFaint() {
		te = te.Faint()
	}
	if s.GetReverse() {
		te = te.Reverse(true)
	}

	ansiStr := te.String()
	if ansiStr != "" {
		return ansiStr, "\x1b[0m"
	}
	return "", ""
}

func renderStyledText(style lipgloss.Style, text string) string {
	prefix, suffix := styleToANSI(style)
	if prefix == "" {
		return text
	}
	return prefix + text + suffix
}

func shouldApplyStyle(cell *uv.Cell) bool {
	if cell == nil {
		return false
	}
	return cell.Style.Fg != nil || cell.Style.Bg != nil || cell.Style.Attrs != 0
}

func buildOptimizedCellStyleCached(cell *uv.Cell, defaultBg color.Color) lipgloss.Style {
	return GetGlobalStyleCache().Get(cell, false, true, defaultBg)
}

func buildCellStyleCached(cell *uv.Cell, isCursor bool, defaultBg color.Color) lipgloss.Style {
	return GetGlobalStyleCache().Get(cell, isCursor, false, defaultBg)
}

func buildOptimizedCellStyleWithDefaultBg(cell *uv.Cell, defaultBg color.Color) lipgloss.Style {
	cellStyle := lipgloss.NewStyle()
	if cell == nil {
		return cellStyle
	}
	if cell.Style.Fg != nil {
		if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Foreground(ansiColor)
		} else if isColorSafe(cell.Style.Fg) {
			cellStyle = cellStyle.Foreground(cell.Style.Fg)
		}
	}
	bg := cell.Style.Bg
	if bg == nil && defaultBg != nil {
		bg = defaultBg
	}
	if bg != nil {
		if ansiColor, ok := bg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Background(ansiColor)
		} else if isColorSafe(bg) {
			cellStyle = cellStyle.Background(bg)
		}
	}
	return cellStyle
}

func buildCellStyleWithDefaultBg(cell *uv.Cell, isCursor bool, defaultBg color.Color) lipgloss.Style {
	cellStyle := lipgloss.NewStyle()
	if cell == nil {
		return cellStyle
	}
	bg := cell.Style.Bg
	if bg == nil && defaultBg != nil {
		bg = defaultBg
	}
	if isCursor {
		fg := lipgloss.Color("#FFFFFF")
		bgColor := lipgloss.Color("#000000")
		if cell.Style.Fg != nil {
			if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
				fg = ansiColor
			} else if isColorSafe(cell.Style.Fg) {
				fg = cell.Style.Fg
			}
		}
		if bg != nil {
			if ansiColor, ok := bg.(lipgloss.ANSIColor); ok {
				bgColor = ansiColor
			} else if isColorSafe(bg) {
				bgColor = bg
			}
		}
		return cellStyle.Background(fg).Foreground(bgColor)
	}
	if cell.Style.Fg != nil {
		if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Foreground(ansiColor)
		} else if isColorSafe(cell.Style.Fg) {
			cellStyle = cellStyle.Foreground(cell.Style.Fg)
		}
	}
	if bg != nil {
		if ansiColor, ok := bg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Background(ansiColor)
		} else if isColorSafe(bg) {
			cellStyle = cellStyle.Background(bg)
		}
	}
	if cell.Style.Attrs != 0 {
		attrs := cell.Style.Attrs
		if attrs&1 != 0 {
			cellStyle = cellStyle.Bold(true)
		}
		if attrs&2 != 0 {
			cellStyle = cellStyle.Faint(true)
		}
		if attrs&4 != 0 {
			cellStyle = cellStyle.Italic(true)
		}
		if attrs&32 != 0 {
			cellStyle = cellStyle.Reverse(true)
		}
		if attrs&128 != 0 {
			cellStyle = cellStyle.Strikethrough(true)
		}
	}
	return cellStyle
}

func buildOptimizedCellStyle(cell *uv.Cell) lipgloss.Style {
	cellStyle := lipgloss.NewStyle()

	if cell == nil {
		return cellStyle
	}

	if cell.Style.Fg != nil {
		if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Foreground(ansiColor)
		} else if isColorSafe(cell.Style.Fg) {
			cellStyle = cellStyle.Foreground(cell.Style.Fg)
		}
	}
	if cell.Style.Bg != nil {
		if ansiColor, ok := cell.Style.Bg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Background(ansiColor)
		} else if isColorSafe(cell.Style.Bg) {
			cellStyle = cellStyle.Background(cell.Style.Bg)
		}
	}

	return cellStyle
}

func isColorSafe(c color.Color) bool {
	if c == nil {
		return false
	}
	switch c.(type) {
	case lipgloss.ANSIColor, lipgloss.NoColor, lipgloss.RGBColor,
		color.RGBA, color.NRGBA, color.Gray, color.Gray16,
		color.RGBA64, color.CMYK, color.Alpha, color.Alpha16,
		color.YCbCr:
		return true
	default:
		// Unknown type  - attempt RGBA() and recover on panic
		safe := true
		func() {
			defer func() {
				if recover() != nil {
					safe = false
				}
			}()
			_, _, _, _ = c.RGBA()
		}()
		return safe
	}
}

func buildCellStyle(cell *uv.Cell, isCursor bool) lipgloss.Style {
	cellStyle := lipgloss.NewStyle()

	if cell == nil {
		return cellStyle
	}

	if isCursor {
		fg := lipgloss.Color("#FFFFFF")
		bg := lipgloss.Color("#000000")
		if cell.Style.Fg != nil {
			if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
				fg = ansiColor
			} else if isColorSafe(cell.Style.Fg) {
				fg = cell.Style.Fg
			}
		}
		if cell.Style.Bg != nil {
			if ansiColor, ok := cell.Style.Bg.(lipgloss.ANSIColor); ok {
				bg = ansiColor
			} else if isColorSafe(cell.Style.Bg) {
				bg = cell.Style.Bg
			}
		}
		return cellStyle.Background(fg).Foreground(bg)
	}

	if cell.Style.Fg != nil {
		if ansiColor, ok := cell.Style.Fg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Foreground(ansiColor)
		} else if isColorSafe(cell.Style.Fg) {
			cellStyle = cellStyle.Foreground(cell.Style.Fg)
		}
	}
	if cell.Style.Bg != nil {
		if ansiColor, ok := cell.Style.Bg.(lipgloss.ANSIColor); ok {
			cellStyle = cellStyle.Background(ansiColor)
		} else if isColorSafe(cell.Style.Bg) {
			cellStyle = cellStyle.Background(cell.Style.Bg)
		}
	}

	if cell.Style.Attrs != 0 {
		attrs := cell.Style.Attrs
		if attrs&1 != 0 {
			cellStyle = cellStyle.Bold(true)
		}
		if attrs&2 != 0 {
			cellStyle = cellStyle.Faint(true)
		}
		if attrs&4 != 0 {
			cellStyle = cellStyle.Italic(true)
		}
		if attrs&32 != 0 {
			cellStyle = cellStyle.Reverse(true)
		}
		if attrs&128 != 0 {
			cellStyle = cellStyle.Strikethrough(true)
		}
	}

	return cellStyle
}

func clipWindowContent(content string, x, y, viewportWidth, viewportHeight int) (string, int, int) {
	lines := strings.Split(content, "\n")
	windowHeight := len(lines)

	windowWidth := 0
	if len(lines) > 0 {
		windowWidth = ansi.StringWidth(lines[0])
	}

	if x+windowWidth <= 0 || x >= viewportWidth || y+windowHeight <= 0 || y >= viewportHeight {
		return "", max(x, 0), max(y, 0)
	}

	clipTop := 0
	clipLeft := 0
	finalX := x
	finalY := y

	if y < 0 {
		clipTop = -y
		finalY = 0
	}

	if x < 0 {
		clipLeft = -x
		finalX = 0
	}

	if clipTop >= len(lines) {
		return "", finalX, finalY
	}
	visibleLines := lines[clipTop:]

	maxVisibleLines := viewportHeight - finalY
	if maxVisibleLines < len(visibleLines) {
		visibleLines = visibleLines[:maxVisibleLines]
	}

	if clipLeft > 0 || finalX+windowWidth > viewportWidth {
		maxWidth := viewportWidth - finalX
		clippedLines := make([]string, len(visibleLines))

		for lineIdx, line := range visibleLines {
			lineWidth := ansi.StringWidth(line)

			if clipLeft >= lineWidth {
				clippedLines[lineIdx] = ""
				continue
			}

			tempLine := line
			if lineWidth > maxWidth+clipLeft {
				tempLine = ansi.Truncate(line, maxWidth+clipLeft, "")
			}

			if clipLeft > 0 {
				result := strings.Builder{}
				pos := 0
				skipCount := clipLeft

				runes := []rune(tempLine)
				runeIdx := 0
				for runeIdx < len(runes) {
					if runes[runeIdx] == '\x1b' {
						seqStart := runeIdx
						runeIdx++

						if runeIdx < len(runes) && runes[runeIdx] == '[' {
							runeIdx++
							for runeIdx < len(runes) && (runes[runeIdx] < 0x40 || runes[runeIdx] > 0x7E) {
								runeIdx++
							}
							if runeIdx < len(runes) {
								runeIdx++
							}
						} else if runeIdx < len(runes) && runes[runeIdx] == ']' {
							runeIdx++
							for runeIdx < len(runes) {
								if runes[runeIdx] == '\x07' || (runes[runeIdx] == '\x1b' && runeIdx+1 < len(runes) && runes[runeIdx+1] == '\\') {
									runeIdx++
									if runeIdx < len(runes) && runes[runeIdx-1] == '\x1b' {
										runeIdx++
									}
									break
								}
								runeIdx++
							}
						}

						// Always include escape sequences  - they set terminal state (colors, styles)
						result.WriteString(string(runes[seqStart:runeIdx]))
						continue
					}

					if pos >= skipCount {
						result.WriteRune(runes[runeIdx])
					}
					pos++
					runeIdx++
				}

				clippedLines[lineIdx] = result.String() + "\x1b[0m"
			} else {
				clippedLines[lineIdx] = tempLine
				if lineWidth > maxWidth {
					clippedLines[lineIdx] += "\x1b[0m"
				}
			}
		}

		return strings.Join(clippedLines, "\n"), finalX, finalY
	}

	return strings.Join(visibleLines, "\n"), finalX, finalY
}

func (m *OS) isPositionInSelection(window *terminal.Window, x, y int) bool {
	if !window.IsSelecting && window.SelectedText == "" {
		return false
	}

	startX, startY := window.SelectionStart.X, window.SelectionStart.Y
	endX, endY := window.SelectionEnd.X, window.SelectionEnd.Y

	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	if y < startY || y > endY {
		return false
	}
	if y == startY && y == endY {
		return x >= startX && x <= endX
	} else if y == startY {
		return x >= startX
	} else if y == endY {
		return x <= endX
	} else {
		return true
	}
}
