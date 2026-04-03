package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/shared/audio"
	"github.com/bentos-lab/parley/wiring"
)

const (
	defaultNumAgents = 2
	defaultNumRounds = 10
)

type Listener struct {
	client       *whatsmeow.Client
	usecases     *wiring.Usecases
	cfg          config.Config
	history      *historyStore
	handlerID    uint32
	parseUsecase *core.ParseParleyCommandUsecase
}

// NewListener builds a WhatsApp listener if a session already exists in the cache.
// Parameters: ctx controls cancellation, usecases provides the debate usecases, cfg drives provider settings.
// Returns: listener ready to start and any error encountered while preparing the WhatsApp client.
func NewListener(ctx context.Context, usecases *wiring.Usecases, cfg config.Config) (*Listener, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("whatsapp session missing: %w", err)
	}
	container, err := newSessionContainer(path)
	if err != nil {
		return nil, fmt.Errorf("open whatsapp store: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("read device: %w", err)
	}
	client := whatsmeow.NewClient(device, nil)
	if usecases == nil {
		return nil, errors.New("usecases are required")
	}
	if usecases.ParseParleyCommand == nil {
		return nil, errors.New("parse usecase is required")
	}
	history, err := newHistoryStore()
	if err != nil {
		return nil, fmt.Errorf("init history store: %w", err)
	}
	return &Listener{
		client:       client,
		usecases:     usecases,
		cfg:          cfg,
		history:      history,
		parseUsecase: usecases.ParseParleyCommand,
	}, nil
}

// Start connects the WhatsApp client and registers an event handler for incoming messages.
// Parameters: ctx controls the listener lifecycle.
// Returns: nothing.
func (l *Listener) Start(ctx context.Context) {
	if err := l.client.Connect(); err != nil {
		log.Printf("whatsapp listener: connect error: %v", err)
		return
	}
	if !l.client.WaitForConnection(30 * time.Second) {
		log.Print("whatsapp listener: failed to wait for connection")
		return
	}
	logWhatsAppInfoLine("[WHATSAPP]", "Listening for WhatsApp events")
	l.handlerID = l.client.AddEventHandler(l.handleEvent)
	go func() {
		<-ctx.Done()
		l.client.RemoveEventHandler(l.handlerID)
		l.client.Disconnect()
	}()
}

// handleEvent filters WhatsApp events and routes qualifying message events to chatLoop.
// Parameters: evt is the raw event published by the WhatsApp client.
// Returns: nothing.
func (l *Listener) handleEvent(evt any) {
	msgEvt, ok := evt.(*events.Message)
	if !ok {
		return
	}
	if !msgEvt.Info.IsFromMe {
		return
	}
	if msgEvt.Info.IsGroup {
		return
	}
	// Only react to commands that are posted back to the paired account's own chat (Saved Messages).
	if msgEvt.Info.Chat != msgEvt.Info.Sender {
		return
	}
	text, isAudio := extractText(msgEvt.Message)
	if text == "" && !isAudio {
		return
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}
	l.chatLoop(msgEvt, trimmed, isAudio)
}

// chatLoop parses `/parley` votes, records chat history, triggers the parse usecase, and runs the resulting command.
// Parameters: evt is the message event, trimmed is the cleaned text, isAudio indicates whether the message is audio-only.
// Returns: nothing.
func (l *Listener) chatLoop(evt *events.Message, trimmed string, isAudio bool) {
	if !strings.HasPrefix(trimmed, "/parley") {
		return
	}
	start := time.Now()
	commandName := "[unknown]"
	l.logWhatsAppEvent(trimmed)
	chatID := evt.Info.Chat.String()
	if !isAudio {
		l.recordHistory(chatID, core.ParleyCommandHistoryMessage{
			Role:    "user",
			Content: trimmed,
		})
	}
	input := core.ParseParleyCommandInput{
		Message:          trimmed,
		History:          l.history.snapshot(chatID),
		DefaultNumAgents: defaultNumAgents,
		DefaultNumRounds: defaultNumRounds,
	}
	output, err := l.parseUsecase.Execute(context.Background(), input)
	if err != nil {
		l.logError("parse command", err)
		l.sendText(evt.Info.Chat, "[parley] Unable to understand your command right now.")
		l.logWhatsAppCommandCompletion(commandName, false, time.Since(start))
		return
	}
	l.logWhatsAppCommand(output.Command)
	if cmd := output.Command.Command; cmd != "" {
		commandName = string(cmd)
	}
	responseText := l.executeCommand(context.Background(), evt.Info.Chat, output.Command)
	if responseText != "" {
		l.sendText(evt.Info.Chat, responseText)
		l.recordHistory(chatID, core.ParleyCommandHistoryMessage{
			Role:    "assistant",
			Content: responseText,
		})
	}
	l.logWhatsAppCommandCompletion(commandName, true, time.Since(start))
}

// recordHistory attempts to persist a single history entry and logs any failure without disrupting execution.
// Parameters: chat identifies the WhatsApp chat and entry captures the role plus content to save.
// Returns: nothing.
func (l *Listener) recordHistory(chat string, entry core.ParleyCommandHistoryMessage) {
	if err := l.history.add(chat, entry); err != nil {
		l.logError("persist history entry", err)
	}
}

// executeCommand routes the parsed command to the appropriate handler.
// Parameters: ctx is the invocation context, chat identifies the recipient, cmd is the parsed command.
// Returns: response text that the assistant should send (if any).
func (l *Listener) executeCommand(ctx context.Context, chat types.JID, cmd core.ParleyCommand) string {
	switch cmd.Command {
	case core.ParleyCommandCreate:
		return l.handleCreate(ctx, chat, cmd)
	case core.ParleyCommandResume:
		return l.handleResume(ctx, chat, cmd)
	case core.ParleyCommandList:
		return l.handleList()
	case core.ParleyCommandDelete:
		return l.handleDelete(cmd)
	case core.ParleyCommandAudio:
		return l.handleAudio(ctx, chat, cmd)
	default:
		return "[parley] Unsupported request"
	}
}

// handleCreate executes the create command, runs rounds, and sends the resulting audio.
// Parameters: ctx is the context, chat identifies the receiving WhatsApp chat, cmd carries the parsed command details.
// Returns: reply text describing success or errors.
func (l *Listener) handleCreate(ctx context.Context, chat types.JID, cmd core.ParleyCommand) string {
	topic := strings.TrimSpace(cmd.Topic)
	if topic == "" {
		return "[parley] Please provide a topic to start the debate."
	}
	nameOutput, err := l.usecases.GenerateDebateName.Execute(ctx, core.GenerateDebateNameInput{Topic: topic})
	if err != nil {
		l.logError("generate debate name", err)
		return "[parley] Unable to start the debate right now."
	}
	agents, err := l.buildAgents(ctx, cmd)
	if err != nil {
		l.logError("build agents", err)
		return "[parley] Unable to start the debate right now."
	}
	voicesOutput, err := l.usecases.AssignDebateVoices.Execute(ctx, core.AssignDebateVoicesInput{
		Agents:      agents,
		TTSProvider: l.cfg.TTSProvider,
	})
	if err != nil {
		l.logError("assign voices", err)
		return "[parley] Unable to start the debate right now."
	}
	createOutput, err := l.usecases.CreateDebate.Execute(ctx, core.CreateDebateInput{
		Name:        nameOutput.Name,
		Topic:       topic,
		Agents:      voicesOutput.Agents,
		TTSProvider: l.cfg.TTSProvider,
	})
	if err != nil {
		l.logError("create debate", err)
		return "[parley] Unable to start the debate right now."
	}
	filename := createOutput.Filename
	id := debate.IDFromFilename(filename)
	for i := 0; i < cmd.NumRounds; i++ {
		if _, err := l.usecases.CreateRound.Execute(ctx, core.CreateRoundInput{Filename: filename}); err != nil {
			l.logError(fmt.Sprintf("create round %d for debate %s", i+1, id), err)
			return "[parley] Unable to complete the debate rounds right now."
		}
	}
	if err := l.sendDebateAudio(ctx, chat, id); err != nil {
		return fmt.Sprintf("[parley] Debate %s created but the audio delivery is pending. %s", id, formatDebateID(id))
	}
	return fmt.Sprintf("[parley] Debate %s is ready and audio has been sent. %s", id, formatDebateID(id))
}

// handleResume appends rounds to the existing debate and re-sends regenerated audio.
// Parameters: ctx is the context, chat identifies the recipient, cmd contains the debate ID and round count.
// Returns: response text summarizing the operation.
func (l *Listener) handleResume(ctx context.Context, chat types.JID, cmd core.ParleyCommand) string {
	if cmd.DebateID == "" {
		return "[parley] Please provide a debate ID to resume."
	}
	filename := debate.FilenameFromID(cmd.DebateID)
	if _, err := l.usecases.LoadDebate.Execute(core.LoadDebateInput{Filename: filename}); err != nil {
		l.logError(fmt.Sprintf("load debate %s", cmd.DebateID), err)
		return "[parley] Unable to resume the debate right now."
	}
	for i := 0; i < cmd.NumRounds; i++ {
		if _, err := l.usecases.CreateRound.Execute(ctx, core.CreateRoundInput{Filename: filename}); err != nil {
			l.logError(fmt.Sprintf("create round %d for debate %s", i+1, cmd.DebateID), err)
			return "[parley] Unable to resume the debate right now."
		}
	}
	if err := l.sendDebateAudio(ctx, chat, cmd.DebateID); err != nil {
		return fmt.Sprintf("[parley] Debate %s resumed but the audio delivery is pending. %s", cmd.DebateID, formatDebateID(cmd.DebateID))
	}
	return fmt.Sprintf("[parley] Debate %s resumed and audio has been sent. %s", cmd.DebateID, formatDebateID(cmd.DebateID))
}

// handleList replies with the IDs of currently stored debates.
// Parameters: ctx is unused, chat identifies the recipient, cmd carries parsing metadata.
// Returns: text listing debate IDs or an error notice.
func (l *Listener) handleList() string {
	result, err := l.usecases.ListDebates.Execute()
	if err != nil {
		l.logError("list debates", err)
		return "[parley] Unable to list debates right now."
	}
	if len(result.Items) == 0 {
		return "[parley] No debates available."
	}
	var ids []string
	for _, item := range result.Items[:10] { // maximum 10 items sent to user
		ids = append(ids, item.ID)
	}
	return fmt.Sprintf("[parley] Debates:\n%s", strings.Join(ids, "\n"))
}

// handleDelete removes the requested debate by ID.
// Parameters: ctx is unused, chat identifies the requester, cmd contains the debate ID to delete.
// Returns: confirmation text or an error.
func (l *Listener) handleDelete(cmd core.ParleyCommand) string {
	if cmd.DebateID == "" {
		return "[parley] Please provide a debate ID to delete."
	}
	filename := debate.FilenameFromID(cmd.DebateID)
	if err := l.usecases.DeleteDebate.Execute(core.DeleteDebateInput{Filename: filename}); err != nil {
		l.logError(fmt.Sprintf("delete debate %s", cmd.DebateID), err)
		return "[parley] Unable to delete the requested debate right now."
	}
	return fmt.Sprintf("[parley] Debate %s deletion confirmed. %s", cmd.DebateID, formatDebateID(cmd.DebateID))
}

// handleAudio streams the audio file for the provided debate ID to the WhatsApp chat.
// Parameters: ctx provides cancellation, chat identifies the recipient, cmd carries the target debate ID.
// Returns: user-facing status text.
func (l *Listener) handleAudio(ctx context.Context, chat types.JID, cmd core.ParleyCommand) string {
	if cmd.DebateID == "" {
		return "[parley] Please provide a debate ID for audio."
	}
	if err := l.sendDebateAudio(ctx, chat, cmd.DebateID); err != nil {
		return "[parley] Unable to generate audio for that debate right now."
	}
	return fmt.Sprintf("[parley] Audio for debate %s has been sent.", formatDebateID(cmd.DebateID))
}

// buildAgents returns explicitly configured agents or generates them as needed.
// Parameters: ctx is the calling context, cmd captures the command details (topic, agent hints, count).
// Returns: resolved agent list or an error.
func (l *Listener) buildAgents(ctx context.Context, cmd core.ParleyCommand) ([]debate.DebateAgent, error) {
	if len(cmd.Agents) > 0 {
		agents := make([]debate.DebateAgent, len(cmd.Agents))
		for i, agent := range cmd.Agents {
			agents[i] = debate.DebateAgent{Name: agent.Name, Stance: agent.Stance}
		}
		return agents, nil
	}
	if cmd.NumAgents <= 0 {
		cmd.NumAgents = defaultNumAgents
	}
	agentsOutput, err := l.usecases.GenerateDebateAgents.Execute(ctx, core.GenerateAgentsInput{
		Topic: cmd.Topic,
		Count: cmd.NumAgents,
	})
	if err != nil {
		return nil, err
	}
	return agentsOutput.Agents, nil
}

// sendDebateAudio generates audio for the debate and streams it to the chat after announcing the file.
// Parameters: ctx controls cancellation, chat identifies the recipient, debateID selects the debate.
// Returns: any error that occurred while generating or sending the audio.
func (l *Listener) sendDebateAudio(ctx context.Context, chat types.JID, debateID string) error {
	filename := debate.FilenameFromID(debateID)
	audioOutput, err := l.usecases.GenerateAudio.Execute(ctx, core.GenerateAudioInput{Filename: filename})
	if err != nil {
		l.logError(fmt.Sprintf("generate audio for debate %s", debateID), err)
		return err
	}
	convertedPath, duration, cleanup, convErr := audio.ConvertWAVToOpus(ctx, audioOutput.Path)
	if convErr == nil {
		defer cleanup()
		if err := l.sendVoiceNote(chat, convertedPath, duration); err != nil {
			l.logError(fmt.Sprintf("send voice note for debate %s", debateID), err)
			return err
		}
		return nil
	}
	l.logError(fmt.Sprintf("convert audio for debate %s", debateID), convErr)
	fallback := fmt.Sprintf("[parley] Debate %s audio conversion failed; sending WAV document instead. %s", debateID, formatDebateID(debateID))
	if err := l.sendText(chat, fallback); err != nil {
		l.logError(fmt.Sprintf("announce conversion failure for debate %s", debateID), err)
		return err
	}
	if err := l.sendWavDocument(chat, audioOutput.Path); err != nil {
		l.logError(fmt.Sprintf("send audio file for debate %s", debateID), err)
		return err
	}
	return nil
}

// sendText posts a plain text message to the given chat.
// Parameters: chat is the WhatsApp identifier, body is the message payload.
// Returns: any error from the WhatsApp client.
func (l *Listener) sendText(chat types.JID, body string) error {
	if chat.IsEmpty() {
		err := errors.New("chat jid is empty")
		l.logError("send text", err)
		return err
	}
	_, err := l.client.SendMessage(context.Background(), chat, &waE2E.Message{
		Conversation: proto.String(body),
	})
	if err != nil {
		l.logError(fmt.Sprintf("send text to %s", chat.String()), err)
	}
	return err
}

// sendVoiceNote uploads audio bytes as a WhatsApp voice note.
// Parameters: chat identifies the recipient, path is the converted audio file, durationSeconds estimates the clip length.
// Returns: any error encountered while uploading or sending the voice note.
func (l *Listener) sendVoiceNote(chat types.JID, path string, durationSeconds uint32) error {
	data, err := os.ReadFile(path)
	if err != nil {
		l.logError(fmt.Sprintf("read audio file %s", path), err)
		return err
	}
	resp, err := l.client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		l.logError(fmt.Sprintf("upload audio file %s", path), err)
		return err
	}
	msg := &waE2E.AudioMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(resp.FileLength),
		Seconds:       proto.Uint32(durationSeconds),
		Mimetype:      proto.String("audio/ogg; codecs=opus"),
		PTT:           proto.Bool(true),
	}
	_, err = l.client.SendMessage(context.Background(), chat, &waE2E.Message{
		AudioMessage: msg,
	})
	if err != nil {
		l.logError(fmt.Sprintf("send audio to %s", chat.String()), err)
	}
	return err
}

// sendWavDocument uploads a WAV file as a document when voice note conversion fails.
// Parameters: chat identifies the recipient, path is the local WAV file path.
// Returns: any error encountered when uploading or sending the document.
func (l *Listener) sendWavDocument(chat types.JID, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		l.logError(fmt.Sprintf("read audio file %s", path), err)
		return err
	}
	resp, err := l.client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		l.logError(fmt.Sprintf("upload audio file %s", path), err)
		return err
	}
	doc := &waE2E.DocumentMessage{
		FileName:      proto.String(filepath.Base(path)),
		Mimetype:      proto.String("audio/wav"),
		FileLength:    proto.Uint64(resp.FileLength),
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
	}
	_, err = l.client.SendMessage(context.Background(), chat, &waE2E.Message{
		DocumentMessage: doc,
	})
	if err != nil {
		l.logError(fmt.Sprintf("send audio to %s", chat.String()), err)
	}
	return err
}

// extractText returns the textual content of a WhatsApp message and whether it was audio-only.
// Parameters: msg is the message payload extracted from the event.
// Returns: the trimmed text and a flag indicating that the message is audio.
func extractText(msg *waE2E.Message) (string, bool) {
	if msg == nil {
		return "", false
	}
	if conv := strings.TrimSpace(msg.GetConversation()); conv != "" {
		return conv, false
	}
	if ext := strings.TrimSpace(msg.GetExtendedTextMessage().GetText()); ext != "" {
		return ext, false
	}
	if msg.GetAudioMessage() != nil {
		return "", true
	}
	return "", false
}

func (l *Listener) logError(context string, err error) {
	if err == nil {
		return
	}
	logWhatsAppError(context, err)
}

func logWhatsAppError(context string, err error) {
	fmt.Fprintf(
		os.Stdout,
		"%s ▸ %s %s %v\n",
		formatLogTime(time.Now()),
		colorizeErrorLabel("ERR"),
		context,
		err,
	)
}

func (l *Listener) logWhatsAppEvent(message string) {
	logWhatsAppInfoLine("[WHATSAPP-REQ]", fmt.Sprintf("%q", message))
}

func (l *Listener) logWhatsAppCommand(command core.ParleyCommand) {
	c, _ := json.Marshal(command)
	logWhatsAppInfoLine("[WHATSAPP-CMD]", fmt.Sprintf("command=%q", string(c)))
}

func (l *Listener) logWhatsAppCommandCompletion(name string, success bool, duration time.Duration) {
	if name == "" {
		name = "[unknown]"
	}
	status := "failure"
	if success {
		status = "success"
	}
	logWhatsAppInfoLine("[WHATSAPP-RES]", fmt.Sprintf("command=%s status=%s duration=%s", name, status, duration))
}

func formatDebateID(id string) string {
	return fmt.Sprintf("Debate id=[%s]", id)
}

func logWhatsAppInfoLine(label, message string) {
	fmt.Fprintf(os.Stdout, "%s ▸ %s %s\n", formatLogTime(time.Now()), colorizeInfoLabel(label), message)
}

func colorizeInfoLabel(label string) string {
	return fmt.Sprintf("%s%s%s%s", ansiBg256(4), ansiFg256(0), label, ansiReset)
}

func formatLogTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

const ansiReset = "\x1b[0m"

func colorizeErrorLabel(label string) string {
	return fmt.Sprintf("%s%s%s%s", ansiBg256(9), ansiFg256(0), label, ansiReset)
}

func ansiFg256(color int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", color)
}

func ansiBg256(color int) string {
	return fmt.Sprintf("\x1b[48;5;%dm", color)
}
