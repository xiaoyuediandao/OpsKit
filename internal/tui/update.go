package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"opskit/internal/achievement"
	"opskit/internal/ai"
	"opskit/internal/health"
	"opskit/internal/quest"
	"opskit/internal/state"
	"opskit/internal/tools"
)

const maxAgentLoopDepth = 10

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tickMsg:
		m.tickCount++
		return m, tea.Every(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.mode == ModeSetup {
		return m.updateSetup(msg)
	}
	return m.updateChat(msg)
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	headerH := 3
	footerH := 3
	viewportH := m.height - headerH - footerH
	if viewportH < 1 {
		viewportH = 1
	}

	viewportW := m.width
	if m.showTaskPanel {
		viewportW = m.width - taskPanelWidth - 1
	}

	if !m.ready {
		m.viewport = viewport.New(viewportW, viewportH)
		m.viewport.SetContent(m.renderMessages())
		m.ready = true
		initMarkdown(viewportW - 4)
		// Focus input now that terminal is initialized — avoids capturing
		// terminal color query responses (e.g. rgb:ffff/ffff/ffff) as input.
		if m.mode == ModeChat {
			m.input.Reset()
			m.input.Focus()
		}
	} else {
		m.viewport.Width = viewportW
		m.viewport.Height = viewportH
		initMarkdown(viewportW - 4)
		m.viewport.SetContent(m.renderMessages())
	}

	inputWidth := m.width - 6
	if m.showTaskPanel {
		inputWidth = m.width - taskPanelWidth - 7
	}
	if inputWidth < 10 {
		inputWidth = 10
	}
	m.input.Width = inputWidth

	return m, nil
}

func (m Model) updateChat(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// If there's a pending tool call waiting for confirmation
		if m.pendingToolCall != nil {
			switch msg.Type {
			case tea.KeyEnter:
				return m.executePendingTool()
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				// Cancel pending tool call
				m.pendingToolCall = nil
				m.pendingToolCalls = nil
				m.agentThinking = false
				m.loading = false
				m.messages = append(m.messages, ChatMessage{
					Role:      "assistant",
					Content:   "好的，我取消了这个操作。还有什么我可以帮你的吗？",
					Timestamp: time.Now(),
				})
				m.refreshViewport()
				m.input.Focus()
				return m, nil
			}
			return m, nil
		}

		if m.loading || m.agentThinking {
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleChatEnter()

		case tea.KeyRunes:
			if string(msg.Runes) == "t" || string(msg.Runes) == "T" {
				if !m.input.Focused() {
					return m.toggleTaskPanel()
				}
			}
		}

	case streamChunkMsg:
		m.streamBuf.WriteString(msg.chunk)
		// Update the last assistant message in real time
		m.updateStreamingMessage(m.streamBuf.String())
		m.refreshViewport()
		return m, streamNextChunk(m.chunkCh)

	case streamDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("流式响应出错: %v", msg.err),
				Timestamp: time.Now(),
			})
		} else {
			// Finalize the streaming message
			content := m.streamBuf.String()
			m.finalizeStreamingMessage(content)
			m.lobster.AddXP(10)
			m.saveFullState()
		}
		m.streamBuf.Reset()
		m.refreshViewport()
		m.input.Focus()
		return m, nil

	case agentThinkDoneMsg:
		return m.handleAgentThinkDone(msg)

	case agentToolDoneMsg:
		return m.handleAgentToolDone(msg)

	case healthTickMsg:
		hp := health.CalculateHP(msg.report, m.lobster.MaxHP)
		m.lobster.HP = hp
		m.lobster.UpdateStatusFromHP()
		m.saveFullState()
		m.refreshViewport()
		return m, m.scheduleHealthTick()
	}

	// Update input and viewport
	var cmds []tea.Cmd
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) handleChatEnter() (Model, tea.Cmd) {
	userText := strings.TrimSpace(m.input.Value())
	if userText == "" {
		return m, nil
	}

	m.input.Reset()
	m.input.Blur()

	// Add user message
	m.messages = append(m.messages, ChatMessage{
		Role:      "user",
		Content:   userText,
		Timestamp: time.Now(),
	})

	// Add to agent history
	m.agentHistory = append(m.agentHistory, ai.AgentMessage{
		Role:    "user",
		Content: userText,
	})

	m.agentThinking = true
	m.loading = true
	m.agentLoopDepth = 0

	m.refreshViewport()

	return m, m.doAgentThink()
}

func (m Model) doAgentThink() tea.Cmd {
	history := m.agentHistory
	client := m.aiClient
	return func() tea.Msg {
		resp, err := client.ChatWithTools(history, tools.AllToolDefinitions())
		return agentThinkDoneMsg{resp: resp, err: err}
	}
}

func (m Model) handleAgentThinkDone(msg agentThinkDoneMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.loading = false
		m.agentThinking = false
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   fmt.Sprintf("哦不，我脑子有点转不过来了 😵\n\n错误: %v", msg.err),
			Timestamp: time.Now(),
		})
		m.refreshViewport()
		m.input.Focus()
		return m, nil
	}

	resp := msg.resp

	// No tool calls -> final response
	if len(resp.ToolCalls) == 0 {
		m.loading = false
		m.agentThinking = false
		content := resp.Content
		if content == "" {
			content = "（我思考了一下，但没有什么要说的...）"
		}
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   content,
			Timestamp: time.Now(),
		})
		// Add to agent history
		m.agentHistory = append(m.agentHistory, ai.AgentMessage{
			Role:    "assistant",
			Content: content,
		})
		m.lobster.AddXP(15)
		m.lobster.Status = "idle"
		m.saveFullState()
		m.refreshViewport()
		m.input.Focus()
		return m, nil
	}

	// Check loop depth
	if m.agentLoopDepth >= maxAgentLoopDepth {
		m.loading = false
		m.agentThinking = false
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   "我已经执行了很多步骤了，先停下来休息一下 🦞💤\n\n如果还需要继续，请告诉我！",
			Timestamp: time.Now(),
		})
		// Keep last few messages for context instead of clearing everything
		if len(m.agentHistory) > 4 {
			m.agentHistory = m.agentHistory[len(m.agentHistory)-4:]
		}
		m.refreshViewport()
		m.input.Focus()
		return m, nil
	}

	// Has tool calls - show preview and wait for user confirmation
	toolCall := resp.ToolCalls[0]
	remaining := resp.ToolCalls[1:]

	// Add assistant message with tool calls to history
	m.agentHistory = append(m.agentHistory, ai.AgentMessage{
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
	})

	// Show thinking text if any
	if resp.Content != "" {
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   resp.Content,
			Timestamp: time.Now(),
		})
	}

	// Show tool preview
	preview := formatToolPreview(toolCall)
	m.messages = append(m.messages, ChatMessage{
		Role:      "tool_preview",
		Content:   preview,
		ToolName:  toolCall.Function.Name,
		Timestamp: time.Now(),
	})

	m.pendingToolCall = &toolCall
	m.pendingToolCalls = remaining
	m.loading = false
	m.agentLoopDepth++

	m.refreshViewport()
	return m, nil
}

func (m Model) executePendingTool() (Model, tea.Cmd) {
	if m.pendingToolCall == nil {
		return m, nil
	}

	tc := *m.pendingToolCall
	m.pendingToolCall = nil
	m.executing = true
	m.loading = true
	m.lobster.Status = "active"

	// Add executing status message
	m.messages = append(m.messages, ChatMessage{
		Role:      "system",
		Content:   fmt.Sprintf("正在执行: %s...", tc.Function.Name),
		Timestamp: time.Now(),
	})
	m.refreshViewport()

	return m, func() tea.Msg {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return agentToolDoneMsg{
				toolCallID: tc.ID,
				toolName:   tc.Function.Name,
				result:     tools.ToolResult{Error: fmt.Sprintf("参数解析失败: %v", err), IsErr: true},
			}
		}
		result := tools.Execute(tc.Function.Name, args)
		return agentToolDoneMsg{
			toolCallID: tc.ID,
			toolName:   tc.Function.Name,
			result:     result,
			args:       args,
		}
	}
}

func (m Model) handleAgentToolDone(msg agentToolDoneMsg) (Model, tea.Cmd) {
	m.executing = false
	prevHP := m.lobster.HP

	// Build result content
	var resultContent string
	if msg.result.IsErr {
		resultContent = fmt.Sprintf("执行出错:\n%s", msg.result.Error)
		if msg.result.Output != "" {
			resultContent += "\n输出:\n" + msg.result.Output
		}
		m.lobster.HP -= 5
		if m.lobster.HP < 0 {
			m.lobster.HP = 0
		}
		m.lobster.Status = "sick"
	} else {
		resultContent = msg.result.Output
		if resultContent == "" {
			resultContent = "(命令执行成功，无输出)"
		}
		m.lobster.AddXP(20)
		m.lobster.Status = "active"
	}

	// Show result in chat
	m.messages = append(m.messages, ChatMessage{
		Role:      "tool_result",
		Content:   resultContent,
		ToolName:  msg.toolName,
		IsError:   msg.result.IsErr,
		Timestamp: time.Now(),
	})

	// Add tool result to agent history
	m.agentHistory = append(m.agentHistory, ai.AgentMessage{
		Role:       "tool",
		Content:    resultContent,
		ToolCallID: msg.toolCallID,
		Name:       msg.toolName,
	})

	// Quest completion and achievement checks on success
	if !msg.result.IsErr {
		m.toolUseCount++

		// Check if this tool execution completes a quest
		completedQuestID := quest.CheckCompletion(msg.toolName, msg.args, msg.result.Output, m.questStates)
		if completedQuestID != "" {
			q := quest.CompleteQuest(completedQuestID, m.questStates)
			if q != nil {
				m.lobster.AddXP(q.XPReward)
				m.messages = append(m.messages, ChatMessage{
					Role:      "system",
					Content:   fmt.Sprintf("🎉 任务完成：%s [%s] %s +%d XP", q.ID, q.Title, q.Description, q.XPReward),
					Timestamp: time.Now(),
				})
				if q.StageReward >= 0 && q.StageReward > m.lobster.Stage {
					oldEmoji := m.lobster.StageEmoji()
					oldName := m.lobster.StageName()
					m.lobster.Stage = q.StageReward
					m.messages = append(m.messages, ChatMessage{
						Role:      "system",
						Content:   fmt.Sprintf("龙虾状态变化：%s %s → %s %s", oldEmoji, oldName, m.lobster.StageEmoji(), m.lobster.StageName()),
						Timestamp: time.Now(),
					})
				}
				m.tasks = questsToTasks(m.questStates)
			}
		}

		// Achievement check
		newAchievements := achievement.Check("tool_use", m.questStates, m.achievements, m.toolUseCount, m.logViewCount)
		for _, a := range newAchievements {
			m.achievements = append(m.achievements, a.ID)
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("🏆 成就解锁：%s「%s」--- %s", a.Icon, a.Name, a.Desc),
				Timestamp: time.Now(),
			})
		}
	}

	// Check revive achievement
	if prevHP <= 0 && m.lobster.HP > 0 {
		reviveAchievements := achievement.Check("revive", m.questStates, m.achievements, m.toolUseCount, m.logViewCount)
		for _, a := range reviveAchievements {
			m.achievements = append(m.achievements, a.ID)
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("🏆 成就解锁：%s「%s」--- %s", a.Icon, a.Name, a.Desc),
				Timestamp: time.Now(),
			})
		}
	}

	m.saveFullState()
	m.refreshViewport()

	// If there are more pending tool calls, show the next one
	if len(m.pendingToolCalls) > 0 {
		next := m.pendingToolCalls[0]
		m.pendingToolCalls = m.pendingToolCalls[1:]

		preview := formatToolPreview(next)
		m.messages = append(m.messages, ChatMessage{
			Role:      "tool_preview",
			Content:   preview,
			ToolName:  next.Function.Name,
			Timestamp: time.Now(),
		})
		m.pendingToolCall = &next
		m.loading = false
		m.agentThinking = false
		m.refreshViewport()
		return m, nil
	}

	// Continue agent loop
	m.agentThinking = true
	m.loading = true
	return m, m.doAgentThink()
}

func (m Model) toggleTaskPanel() (Model, tea.Cmd) {
	m.showTaskPanel = !m.showTaskPanel
	viewportW := m.width
	if m.showTaskPanel {
		viewportW = m.width - taskPanelWidth - 1
	}
	viewportH := m.height - 3 - 3
	if viewportH < 1 {
		viewportH = 1
	}
	m.viewport.Width = viewportW
	m.viewport.Height = viewportH
	inputWidth := m.width - 6
	if m.showTaskPanel {
		inputWidth = m.width - taskPanelWidth - 7
	}
	if inputWidth < 10 {
		inputWidth = 10
	}
	m.input.Width = inputWidth
	m.refreshViewport()
	return m, nil
}

func (m *Model) refreshViewport() {
	content := m.renderMessages()
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) saveFullState() {
	st := m.lobster.ToState()
	st.Quests = m.questStates
	st.Achievements = m.achievements
	st.ToolUseCount = m.toolUseCount
	st.LogViewCount = m.logViewCount
	_ = state.Save(st)
}

func (m Model) scheduleHealthTick() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(60 * time.Second)
		report := health.RunProbes()
		return healthTickMsg{report: report}
	}
}

func (m *Model) updateStreamingMessage(content string) {
	// Find or update the last streaming assistant message
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant_streaming" {
			m.messages[i].Content = content
			return
		}
	}
	// Add new streaming message
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant_streaming",
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (m *Model) finalizeStreamingMessage(content string) {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant_streaming" {
			m.messages[i].Role = "assistant"
			m.messages[i].Content = content
			return
		}
	}
	// Fallback
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	})
}

// streamNextChunk returns a Cmd that waits for the next chunk from the channel.
func streamNextChunk(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return streamDoneMsg{}
		}
		return streamChunkMsg{chunk: chunk}
	}
}

// formatToolPreview formats a tool call for display.
func formatToolPreview(tc ai.ToolCall) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return fmt.Sprintf("工具调用: **%s**\n\n⚠ 参数解析失败: %v\n\n原始参数: `%s`\n\n按 **Esc** 取消", tc.Function.Name, err, tc.Function.Arguments)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("工具调用: **%s**\n\n", tc.Function.Name))

	for k, v := range args {
		switch val := v.(type) {
		case string:
			if len(val) > 200 {
				val = val[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("• %s: `%s`\n", k, val))
		default:
			sb.WriteString(fmt.Sprintf("• %s: %v\n", k, v))
		}
	}

	sb.WriteString("\n按 **Enter** 确认执行，按 **Esc** 取消")
	return sb.String()
}
