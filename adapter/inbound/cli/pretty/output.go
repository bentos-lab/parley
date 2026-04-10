package pretty

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
)

var (
	accentColor = lipgloss.Color("#3B82F6")
	borderColor = lipgloss.Color("#60A5FA")
	mutedColor  = lipgloss.Color("#94A3B8")
)

var (
	titleStyle  = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	bulletStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
)

func New() *Output {
	return &Output{
		agentThemes: make(map[string]int),
		lastTheme:   -1,
	}
}

func (o *Output) DebateHeader(writer io.Writer, summary cli.DebateSummary, agents []cli.AgentRow) error {
	header := formatHeader(summary)
	topicSection := HeaderCard("Topic", mutedStyle.Render(summary.Topic))
	agentRows := make([][]string, 0, len(agents))
	for _, agent := range agents {
		agentRows = append(agentRows, []string{
			agent.ID,
			agent.Name,
			agent.Stance,
			agent.Voice,
		})
	}
	agentsSection := ""
	if len(agentRows) == 0 {
		agentsSection = HeaderCard("Agents", BulletList([]string{"(none)"}))
	} else {
		agentsSection = HeaderCard("Agents", AgentsTable(agentRows))
	}
	if _, err := fmt.Fprintln(writer, header+"\n\n"+topicSection+"\n\n"+agentsSection); err != nil {
		return err
	}
	return nil
}

func (o *Output) DebateBasics(writer io.Writer, basics cli.DebateBasics) error {
	header := formatBasicsHeader(basics)
	topicSection := HeaderCard("Topic", mutedStyle.Render(basics.Topic))
	if _, err := fmt.Fprintln(writer, header+"\n\n"+topicSection); err != nil {
		return err
	}
	return nil
}

func (o *Output) DebateName(writer io.Writer, name string) error {
	width := contentWidth()
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	line := center.Render(mutedStyle.Render(name))
	if _, err := fmt.Fprintln(writer, line); err != nil {
		return err
	}
	return nil
}

func (o *Output) DebateAgents(writer io.Writer, agents []cli.AgentRow) error {
	agentRows := make([][]string, 0, len(agents))
	for _, agent := range agents {
		agentRows = append(agentRows, []string{
			agent.ID,
			agent.Name,
			agent.Stance,
			agent.Voice,
		})
	}
	agentsSection := ""
	if len(agentRows) == 0 {
		agentsSection = HeaderCard("Agents", BulletList([]string{"(none)"}))
	} else {
		agentsSection = HeaderCard("Agents", AgentsTable(agentRows))
	}
	if _, err := fmt.Fprintln(writer, agentsSection); err != nil {
		return err
	}
	return nil
}

func (o *Output) DebateRound(writer io.Writer, roundNumber int, agentName string, message string, summary string, weakness string, newPoint string, rebuttal string) error {
	_ = weakness
	_ = newPoint
	_ = rebuttal
	themeIndex := o.themeIndexForAgent(agentName)
	theme := roundThemes[themeIndex]
	formatted := formatRoundMessage(message)
	content := BulletList([]string{
		fmt.Sprintf("%s: %s", renderRoundWithStyle("Voice", true, false), formatted),
		"",
		fmt.Sprintf("%s: %s", renderRoundWithStyle("Summary", true, false), summary),
	})
	output := RoundCardWithTheme(roundNumber, agentName, content, theme)
	o.lastAgent = agentName
	o.lastTheme = themeIndex
	if _, err := fmt.Fprintln(writer, output); err != nil {
		return err
	}
	return nil
}

// DebateSummary renders the debate summary after rounds complete.
// Parameters: writer is the output destination, summary is the summary payload with agents metadata.
// Returns: an error if writing fails.
func (o *Output) DebateSummary(writer io.Writer, summary cli.DebateSummaryOutput) error {
	content := buildSummaryContent(summary)
	card := HeaderCard("Summary", content)
	if _, err := fmt.Fprintln(writer, card); err != nil {
		return err
	}
	return nil
}

func (o *Output) DebateResult(writer io.Writer, file string, id string) error {
	bullets := []string{fmt.Sprintf("File: %s", file)}
	if id != "" {
		bullets = append(bullets,
			fmt.Sprintf("ID: %s", id),
		)
	}
	if _, err := fmt.Fprintln(writer, StatusLine("Result", bullets)); err != nil {
		return err
	}
	return nil
}

// DebateDetails renders the full debate payload for the get command.
// Parameters: writer is the output destination, details contains full debate information.
// Returns: an error if writing fails.
func (o *Output) DebateDetails(writer io.Writer, details cli.DebateDetailsOutput) error {
	header := formatHeader(cli.DebateSummary{
		Name:        details.Name,
		Topic:       details.Topic,
		TTSProvider: details.Header.TTSProvider,
		AppName:     details.Header.AppName,
		AgentsCount: details.Header.AgentsCount,
		LLMProvider: details.Header.LLMProvider,
		LLMModel:    details.Header.LLMModel,
		TTSModel:    details.Header.TTSModel,
	})
	identifier := HeaderCard("ID", mutedStyle.Render(details.Header.ID))
	topic := HeaderCard("Topic", mutedStyle.Render(details.Topic))

	agentRows := make([][]string, 0, len(details.Agents))
	for _, agent := range details.Agents {
		agentRows = append(agentRows, []string{agent.ID, agent.Name, agent.Stance, agent.Voice})
	}
	agents := HeaderCard("Agents", AgentsTable(agentRows))
	if len(agentRows) == 0 {
		agents = HeaderCard("Agents", BulletList([]string{"(none)"}))
	}

	roundSections := make([]string, 0, len(details.Rounds))
	for _, round := range details.Rounds {
		content := BulletList([]string{
			fmt.Sprintf("%s: %s", renderRoundWithStyle("Agent", true, false), round.AgentName),
			fmt.Sprintf("%s: %s", renderRoundWithStyle("Voice", true, false), formatRoundMessage(round.Message)),
			fmt.Sprintf("%s: %s", renderRoundWithStyle("Summary", true, false), round.Summary),
			fmt.Sprintf("%s: %s", renderRoundWithStyle("Weakness", true, false), round.Weakness),
			fmt.Sprintf("%s: %s", renderRoundWithStyle("New Point", true, false), round.NewPoint),
			fmt.Sprintf("%s: %s", renderRoundWithStyle("Rebuttal", true, false), round.Rebuttal),
		})
		roundSections = append(roundSections, RoundCardWithTheme(round.Number, round.AgentName, content, roundThemes[round.Number%len(roundThemes)]))
	}
	if len(roundSections) == 0 {
		roundSections = append(roundSections, HeaderCard("Rounds", BulletList([]string{"(none)"})))
	}

	summary := buildSummaryContent(cli.DebateSummaryOutput{
		Summary: details.Summary,
		Agents:  details.Agents,
	})
	summaryCard := HeaderCard("Summary", summary)

	all := []string{header, identifier, topic, agents}
	all = append(all, roundSections...)
	all = append(all, summaryCard)
	_, err := fmt.Fprintln(writer, strings.Join(all, "\n\n"))
	return err
}

func (o *Output) ListDebates(writer io.Writer, ids []string) error {
	if len(ids) == 0 {
		if _, err := fmt.Fprintln(writer, "No debates found."); err != nil {
			return err
		}
		return nil
	}
	for _, id := range ids {
		if _, err := fmt.Fprintln(writer, id); err != nil {
			return err
		}
	}
	return nil
}

func (o *Output) InstallGuide(writer io.Writer, title string, guide string) error {
	if _, err := fmt.Fprintln(writer, InstallGuide(title, guide)); err != nil {
		return err
	}
	return nil
}

// HeaderCard renders a titled, framed card with the provided content.
func HeaderCard(title string, content string) string {
	body := content
	if title != "" {
		sizing := currentSizing()
		separator := mutedStyle.Render(strings.Repeat("─", sizing.maxWidth-6))
		centeredTitle := lipgloss.NewStyle().Width(contentWidth()).Align(lipgloss.Center).Render(titleStyle.Render(title))
		body = centeredTitle + "\n" + separator + "\n" + content
	}
	return frame(body)
}

// InfoTable renders a two-column table with a styled label column.
func InfoTable(rows [][2]string) string {
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, fmt.Sprintf("%s: %s", row[0], row[1]))
	}
	return BulletList(items)
}

func formatHeader(summary cli.DebateSummary) string {
	width := contentWidth()
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	title := center.Render(titleStyle.Render(summary.AppName))
	lines := []string{
		center.Render(mutedStyle.Render(summary.Name)),
		center.Render(mutedStyle.Render(fmt.Sprintf("Thinking: %s • %s", summary.LLMProvider, summary.LLMModel))),
		center.Render(mutedStyle.Render(fmt.Sprintf("Speaking: %s • %s", summary.TTSProvider, summary.TTSModel))),
		center.Render(mutedStyle.Render(fmt.Sprintf("Agents: %d", summary.AgentsCount))),
	}
	return title + "\n" + strings.Join(lines, "\n")
}

func formatBasicsHeader(summary cli.DebateBasics) string {
	width := contentWidth()
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	title := center.Render("=== " + titleStyle.Render(summary.AppName) + " ===")
	lines := []string{
		center.Render(mutedStyle.Render(fmt.Sprintf("Thinking: %s • %s", summary.LLMProvider, summary.LLMModel))),
		center.Render(mutedStyle.Render(fmt.Sprintf("Speaking: %s • %s", summary.TTSProvider, summary.TTSModel))),
	}
	return title + "\n" + strings.Join(lines, "\n")
}

// buildSummaryContent builds the summary card content for pretty output.
// Parameters: summary is the summary payload with agents metadata.
// Returns: the formatted summary content string.
func buildSummaryContent(summary cli.DebateSummaryOutput) string {
	sections := make([]string, 0, len(summary.Agents)+1)
	for index, agent := range summary.Agents {
		points := []string{"(no points)"}
		if index < len(summary.Summary.Agents) && len(summary.Summary.Agents[index]) > 0 {
			points = summary.Summary.Agents[index]
		}
		header := fmt.Sprintf("%s (%s)", agent.Name, agent.Stance)
		section := titleStyle.Render(header) + "\n" + BulletList(points)
		sections = append(sections, section)
	}
	conclusion := summary.Summary.Conclusion
	if conclusion == "" {
		conclusion = "(none)"
	}
	sections = append(sections, fmt.Sprintf("%s %s", titleStyle.Render("Conclusion:"), conclusion))
	return strings.Join(sections, "\n\n")
}

// AgentsTable renders a table for debate agents.
func AgentsTable(agents [][]string) string {
	blocks := make([]string, 0, len(agents))
	for _, agent := range agents {
		items := []string{
			fmt.Sprintf("Name: %s", safeCell(agent, 1)),
			fmt.Sprintf("Stance: %s", safeCell(agent, 2)),
			fmt.Sprintf("Voice: %s", safeCell(agent, 3)),
		}
		blocks = append(blocks, BulletList(items))
	}
	return strings.Join(blocks, "\n\n")
}

// RoundCardWithTheme renders a framed round block using a themed title and border.
func RoundCardWithTheme(roundNumber int, agentName string, message string, theme roundTheme) string {
	title := fmt.Sprintf("Round %d — %s", roundNumber, agentName)
	return HeaderCardWithTheme(title, message, theme)
}

// BulletList renders a list of bullet points with wrapped text.
func BulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	prefix := bulletStyle.Render("• ")
	lines := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			lines = append(lines, prefix+item)
		} else {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

// StatusLine renders a compact status card with bullets.
func StatusLine(title string, bullets []string) string {
	return HeaderCard(title, BulletList(bullets))
}

// InstallGuide renders a framed code block for install instructions.
func InstallGuide(title string, guide string) string {
	code := lipgloss.NewStyle().Foreground(mutedColor).Render(guide)
	return HeaderCard(title, code)
}

func frame(body string) string {
	width := contentWidth()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)
	return style.Render(body)
}

func frameWithTheme(body string, theme roundTheme) string {
	width := contentWidth()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.border).
		Padding(0, 1).
		Width(width)
	return style.Render(body)
}

func contentWidth() int {
	sizing := currentSizing()
	inner := max(sizing.maxWidth-4, sizing.minWrapWidth)
	if inner < sizing.minContentWidth && sizing.maxWidth >= sizing.minContentWidth+4 {
		inner = sizing.minContentWidth
	}
	if inner < sizing.minWrapWidth {
		inner = sizing.minWrapWidth
	}
	return inner
}

func safeCell(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

type roundTheme struct {
	accent lipgloss.Color
	border lipgloss.Color
}

var roundThemes = []roundTheme{
	{accent: lipgloss.Color("#10b416"), border: lipgloss.Color("#10b416")},
	{accent: lipgloss.Color("#f59e0b"), border: lipgloss.Color("#f59e0b")},
	{accent: lipgloss.Color("#98e4cb"), border: lipgloss.Color("#98e4cb")},
	{accent: lipgloss.Color("#eb2727"), border: lipgloss.Color("#eb2727")},
	{accent: lipgloss.Color("#a290cd"), border: lipgloss.Color("#a290cd")},
}

// HeaderCardWithTheme renders a titled, framed card with themed title/border.
func HeaderCardWithTheme(title string, content string, theme roundTheme) string {
	body := content
	if title != "" {
		sizing := currentSizing()
		separator := mutedStyle.Render(strings.Repeat("─", sizing.maxWidth-6))
		titleStyle := lipgloss.NewStyle().Foreground(theme.accent).Bold(true)
		centeredTitle := lipgloss.NewStyle().Width(contentWidth()).Align(lipgloss.Center).Render(titleStyle.Render(title))
		body = centeredTitle + "\n" + separator + "\n" + content
	}
	return frameWithTheme(body, theme)
}

type Output struct {
	agentThemes map[string]int
	lastAgent   string
	lastTheme   int
	nextTheme   int
}

func (o *Output) themeIndexForAgent(name string) int {
	if index, ok := o.agentThemes[name]; ok {
		return index
	}
	if len(roundThemes) == 0 {
		return 0
	}
	index := o.nextTheme % len(roundThemes)
	if name != o.lastAgent && len(roundThemes) > 1 && index == o.lastTheme {
		index = (index + 1) % len(roundThemes)
	}
	o.agentThemes[name] = index
	o.nextTheme = (index + 1) % len(roundThemes)
	return index
}

type sizing struct {
	maxWidth        int
	minContentWidth int
	minWrapWidth    int
}

func currentSizing() sizing {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || width <= 0 {
		width = 80
	}
	maxWidth := int(float64(width) * 0.9)
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > width {
		maxWidth = width
	}
	minContentWidth := max(int(float64(width)*0.5), 30)
	if minContentWidth > maxWidth-4 {
		minContentWidth = maxWidth - 4
	}
	minWrapWidth := int(float64(width) * 0.2)
	if minWrapWidth < 10 {
		minWrapWidth = 10
	}
	if minWrapWidth > minContentWidth {
		minWrapWidth = minContentWidth
	}
	if minWrapWidth < 5 {
		minWrapWidth = 5
	}
	return sizing{
		maxWidth:        maxWidth,
		minContentWidth: minContentWidth,
		minWrapWidth:    minWrapWidth,
	}
}
