package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chxcodepro/qilin-manager-tui/internal/system"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#0F172A")).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#CBD5E1"))

	activeTabStyle = tabStyle.
			Bold(true).
			Foreground(lipgloss.Color("#0F172A")).
			Background(lipgloss.Color("#F59E0B"))

	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(1, 2)

	cardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#475569")).
			Padding(1, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#FCD34D"))

	selectedCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#F59E0B")).
				Bold(true)

	editingCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(lipgloss.Color("#2563EB")).
				Bold(true)

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	yellowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24"))

	redStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444"))
)

func (m model) viewHeader() string {
	tabs := []string{
		m.renderTab(sectionOverview, "系统"),
		m.renderTab(sectionDisk, "磁盘"),
		m.renderTab(sectionPerf, "资源"),
		m.renderTab(sectionPackage, "软件"),
		m.renderTab(sectionMaintain, "维护"),
	}

	right := "版本 " + m.version
	if !m.snapshot.GeneratedAt.IsZero() {
		right += " | 更新于 " + m.snapshot.GeneratedAt.Format("15:04:05")
	}
	if m.loading {
		right += " | 刷新中"
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Top, titleStyle.Render("银河麒麟 TUI 管理面板"), "  ", strings.Join(tabs, " "))
	return panelStyle.Width(max(m.width-4, 60)).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerLine,
			labelStyle.Render(right),
		),
	)
}

func (m model) renderTab(target section, title string) string {
	if m.active == target {
		return activeTabStyle.Render(title)
	}
	return tabStyle.Render(title)
}

func (m model) viewBody() string {
	width := max(m.width-4, 60)
	bodyWidth := width - 6
	switch m.active {
	case sectionOverview:
		return panelStyle.Width(width).Render(m.viewOverview(bodyWidth))
	case sectionDisk:
		return panelStyle.Width(width).Render(m.viewDisk(bodyWidth))
	case sectionPerf:
		return panelStyle.Width(width).Render(m.viewPerf(bodyWidth))
	case sectionPackage:
		return panelStyle.Width(width).Render(m.viewPackages(bodyWidth))
	case sectionMaintain:
		return panelStyle.Width(width).Render(m.viewMaintain(bodyWidth))
	default:
		return panelStyle.Width(width).Render("未知页面")
	}
}

func (m model) viewOverview(width int) string {
	sysCard := renderInfoCard("系统概览", m.snapshot.System.Items, width/3-2)

	tableW := width - width/3 - 2
	table := cardStyle.Width(tableW).Render(
		highlightStyle.Render("网卡配置表") + "\n" +
			labelStyle.Render("Enter 编辑 | Ctrl+S 保存当前行") + "\n" +
			renderNetworkTable(m, tableW-6),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, sysCard, "  ", table)
}

func (m model) viewDisk(width int) string {
	mountCard := cardStyle.Width(width/3 - 2).Render(
		highlightStyle.Render("挂载信息") + "\n" +
			renderList(m.snapshot.Disk.Filesystems, "暂无数据"),
	)

	entryW := width - width/3 - 2
	pathLine := labelStyle.Render("路径: ") + valueStyle.Render(m.snapshot.Disk.Target) +
		"  " + labelStyle.Render("上级: ") + valueStyle.Render(firstText(m.snapshot.Disk.Parent, "/"))
	entryCard := cardStyle.Width(entryW).Render(
		highlightStyle.Render("子项占用") + "\n" +
			pathLine + "\n" +
			labelStyle.Render("↑/↓ 选择 | Enter 进入 | Backspace 返回") + "\n" +
			renderDiskEntries(m.snapshot.Disk.Entries, m.diskCursor, entryW-20),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, mountCard, "  ", entryCard)
}

func (m model) viewPerf(width int) string {
	pairs := make([]string, 0, len(m.snapshot.Perf.Summary))
	for _, item := range m.snapshot.Perf.Summary {
		pairs = append(pairs, labelStyle.Render(item.Label+":")+valueStyle.Render(" "+item.Value))
	}

	return cardStyle.Width(width).Render(
		highlightStyle.Render("系统资源") + "\n" +
			strings.Join(pairs, "  ") + "\n" +
			renderProcessTable(m.snapshot.Perf.Top, m.perfCursor, width-6),
	)
}

func (m model) viewPackages(width int) string {
	contentW := width - 6

	visibleApps := m.visibleApps()

	nameW, pkgW := 0, 0
	for _, app := range visibleApps {
		if w := displayWidth(app.Name); w > nameW {
			nameW = w
		}
		if w := displayWidth(app.Package); w > pkgW {
			pkgW = w
		}
	}
	nameW += 2
	pkgW += 1
	statusColW := 14
	lineOverhead := 4 + 1 + 1 + 1
	descW := contentW - nameW - pkgW - statusColW - lineOverhead
	if descW < 6 {
		descW = 6
	}

	appLines := make([]string, 0, len(visibleApps)+3)

	if m.searchMode {
		appLines = append(appLines, highlightStyle.Render("搜索: ")+m.searchInput+"_")
	} else if m.showSearch {
		appLines = append(appLines, highlightStyle.Render("搜索结果")+" (Esc 返回默认列表)")
	}

	appLines = append(appLines, labelStyle.Render("    "+padRight("名称", nameW)+" "+padRight("包名", pkgW)+" "+padRight("状态", statusColW)+" 说明"))
	for idx, app := range visibleApps {
		selected := " "
		if m.selectedApps[app.Package] {
			selected = "x"
		}

		desc := truncateText(app.Description, descW)

		if idx == m.appCursor {
			statusText := "未安装"
			if app.Upgradable {
				statusText = "↑" + truncateText(app.InstalledVer, statusColW-2)
			} else if app.Installed {
				statusText = "✓" + truncateText(app.InstalledVer, statusColW-2)
			}
			line := "[" + selected + "] " + padRight(app.Name, nameW) + " " + padRight(app.Package, pkgW) + " " + padRight(statusText, statusColW) + " " + desc
			appLines = append(appLines, selectedRowStyle.Render(line))
		} else {
			var statusCell string
			if app.Upgradable {
				statusCell = yellowStyle.Width(statusColW).Render("↑" + truncateText(app.InstalledVer, statusColW-2))
			} else if app.Installed {
				statusCell = greenStyle.Width(statusColW).Render("✓" + truncateText(app.InstalledVer, statusColW-2))
			} else {
				statusCell = dimStyle.Width(statusColW).Render("未安装")
			}
			line := "[" + selected + "] " + padRight(app.Name, nameW) + " " + padRight(app.Package, pkgW) + " " + statusCell + " " + desc
			appLines = append(appLines, line)
		}
	}

	return cardStyle.Width(width).Render(
		highlightStyle.Render("软件管理") + "  " +
			highlightStyle.Render("[i]") + " 安装  " +
			highlightStyle.Render("[d]") + " 卸载  " +
			highlightStyle.Render("[/]") + " 搜索\n" +
			renderList(appLines, "暂无软件"),
	)
}

func (m model) viewMaintain(width int) string {
	statusLine := labelStyle.Render("源状态:") +
		" apt " + boolText(m.snapshot.Packages.AptReady) +
		" | sudo " + boolText(m.snapshot.Packages.SudoReady) +
		" | 备份源 " + boolText(m.snapshot.Packages.BackupExists)

	actions := m.maintainActions()
	lines := make([]string, 0, len(actions)+2)
	lines = append(lines, statusLine, "")
	for idx, a := range actions {
		prefix := "  "
		if idx == m.maintainCursor {
			line := fmt.Sprintf("▸ %s    %s", a.Title, a.Preview)
			lines = append(lines, selectedRowStyle.Render(line))
		} else {
			lines = append(lines, prefix+highlightStyle.Render(a.Title)+"    "+labelStyle.Render(a.Preview))
		}
	}

	return cardStyle.Width(width).Render(
		highlightStyle.Render("系统维护") + "  " + labelStyle.Render("↑/↓ 选择 | Enter 执行") + "\n" +
			strings.Join(lines, "\n"),
	)
}

func (m model) viewFooter() string {
	width := max(m.width-4, 60)
	lines := []string{"状态: " + m.status}
	if m.showHelp {
		lines = append(lines, "全局: Tab/Shift+Tab 切页 | r 刷新 | ` 控制台 | ? 帮助 | q 退出")
		switch m.active {
		case sectionOverview:
			lines = append(lines, "系统页: ↑/↓ 选行 | Enter 编辑网卡")
		case sectionDisk:
			lines = append(lines, "磁盘页: ↑/↓ 选项 | Enter 进入目录 | Backspace 返回")
		case sectionPerf:
			lines = append(lines, "资源页: ↑/↓ 选进程 | x 终止进程")
		case sectionPackage:
			lines = append(lines, "软件页: ↑/↓ 选中 | 空格勾选 | / 搜索 | i 安装 | d 卸载 | Esc 返回列表")
		case sectionMaintain:
			lines = append(lines, "维护页: ↑/↓ 选择 | Enter 执行维护命令")
		}
	}
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m model) viewConfirmDialog() string {
	if m.confirming == nil {
		return ""
	}
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		highlightStyle.Render("请确认"),
		m.confirming.action.Title,
		labelStyle.Render(m.confirming.action.Confirm),
		"",
		"按 y 或 Enter 确认，按 n 或 Esc 取消",
	)
	return cardStyle.
		Width(min(max(m.width/2, 40), 80)).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Render(content)
}

func (m model) viewNetworkDialog() string {
	draft := m.currentNetworkDraft()
	if draft == nil {
		return ""
	}

	labels := []string{"模式", "IP 地址", "子网掩码", "网关", "DNS"}

	lines := make([]string, 0, len(labels)+5)
	lines = append(lines, highlightStyle.Render("编辑网卡: "+draft.Device))
	if draft.Connection != "" {
		lines = append(lines, labelStyle.Render("连接: "+draft.Connection))
	}
	lines = append(lines, "")

	for i, label := range labels {
		value := m.netDialog.Values[i]
		if i == netFieldMode {
			if value == "DHCP" {
				value = greenStyle.Render("DHCP") + " / " + dimStyle.Render("静态")
			} else {
				value = dimStyle.Render("DHCP") + " / " + greenStyle.Render("静态")
			}
		}

		labelText := padRight(label, 8)
		if i == m.netDialog.Field {
			if i != netFieldMode {
				value = value + "_"
			}
			lines = append(lines, selectedCellStyle.Render(labelText+"  "+value))
		} else {
			lines = append(lines, labelStyle.Render(labelText)+"  "+valueStyle.Render(value))
		}
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("↑/↓ 切换 | 空格切换模式 | Ctrl+S 保存 | Esc 取消"))

	return cardStyle.
		Width(min(max(m.width/2, 45), 70)).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func renderInfoCard(title string, items []system.InfoItem, width int) string {
	lines := make([]string, 0, len(items)+1)
	lines = append(lines, highlightStyle.Render(title))
	maxLW := 0
	for _, item := range items {
		if w := displayWidth(item.Label); w > maxLW {
			maxLW = w
		}
	}
	for _, item := range items {
		label := padRight(item.Label, maxLW)
		lines = append(lines, labelStyle.Render(label+":")+" "+valueStyle.Render(item.Value))
	}
	return cardStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func renderProcessTable(items []system.ProcessItem, cursor int, availWidth int) string {
	pidW := 7
	cpuW := 7
	memW := 7
	nameW := availWidth - pidW - cpuW - memW - 3
	if nameW < 12 {
		nameW = 12
	}
	if nameW > 30 {
		nameW = 30
	}

	lines := []string{
		labelStyle.Render(padRight("PID", pidW) + " " + padRight("进程", nameW) + " " + padRight("CPU%", cpuW) + " " + "MEM%"),
	}
	for idx, item := range items {
		line := padRight(item.PID, pidW) + " " + padRight(truncateText(item.Name, nameW), nameW) + " " + padRight(item.CPU, cpuW) + " " + item.Memory
		if idx == cursor {
			line = selectedRowStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderPackageState(state system.PackageSection) string {
	type kv struct{ label, value string }
	items := []kv{
		{"apt 可用", boolText(state.AptReady)},
		{"sudo 可用", boolText(state.SudoReady)},
		{"备份源存在", boolText(state.BackupExists)},
	}
	maxLW := 0
	for _, item := range items {
		if w := displayWidth(item.label); w > maxLW {
			maxLW = w
		}
	}
	lines := make([]string, 0, len(items)+3)
	for _, item := range items {
		lines = append(lines, padRight(item.label, maxLW)+": "+item.value)
	}
	lines = append(lines, "", "当前 sources.list 预览:")
	lines = append(lines, state.SourceLines...)
	return strings.Join(lines, "\n")
}

func renderNetworkTable(m model, availWidth int) string {
	headers := []string{"网卡", "状态", "模式", "IP", "掩码", "网关", "DNS", "连接"}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, draft := range m.networkDrafts {
		values := []string{draft.Device, draft.State, draft.Mode, draft.Address, draft.Mask, draft.Gateway, draft.DNS, draft.Connection}
		for i, v := range values {
			if w := displayWidth(v); w > widths[i] {
				widths[i] = w
			}
		}
	}
	for i := range widths {
		widths[i] += 2
	}

	separators := len(widths) - 1
	total := separators
	for _, w := range widths {
		total += w
	}
	if availWidth > 0 && total > availWidth {
		usable := availWidth - separators
		contentTotal := total - separators
		for i := range widths {
			widths[i] = max(widths[i]*usable/contentTotal, 4)
		}
	} else if availWidth > 0 && total < availWidth {
		extra := availWidth - total
		for extra > 0 {
			for i := range widths {
				if extra <= 0 {
					break
				}
				widths[i]++
				extra--
			}
		}
	}

	lines := []string{renderTableRow(headers, widths, -1, -1, true)}
	if len(m.networkDrafts) == 0 {
		lines = append(lines, "暂无网卡")
		return strings.Join(lines, "\n")
	}

	for rowIdx, draft := range m.networkDrafts {
		values := []string{
			draft.Device,
			draft.State,
			draft.Mode,
			draft.Address,
			draft.Mask,
			draft.Gateway,
			draft.DNS,
			draft.Connection,
		}
		line := renderTableRow(values, widths, rowIdx, m.networkCursor, false)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func renderTableRow(values []string, widths []int, rowIdx int, cursorRow int, header bool) string {
	cells := make([]string, 0, len(values))
	for idx, value := range values {
		text := truncateText(firstText(value, "-"), widths[idx])
		style := lipgloss.NewStyle().Width(widths[idx])
		if header {
			style = style.Bold(true)
		}
		if rowIdx == cursorRow {
			style = selectedRowStyle.Width(widths[idx])
		}
		cells = append(cells, style.Render(text))
	}
	return strings.Join(cells, " ")
}

func renderDiskEntries(entries []system.DiskEntry, cursor int, nameWidth int) string {
	if len(entries) == 0 {
		return "当前目录没有子项，或需要更高权限"
	}
	lines := []string{padRight("大小", 8) + " " + padRight("类型", 4) + " " + "名称"}
	for idx, entry := range entries {
		kind := "文件"
		if entry.IsDir {
			kind = "目录"
		}
		line := padRight(entry.Size, 8) + " " + padRight(kind, 4) + " " + truncateText(entry.Name, nameWidth)
		if idx == cursor {
			line = selectedRowStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func truncateText(value string, width int) string {
	value = strings.TrimSpace(value)
	if width <= 0 {
		return ""
	}
	if displayWidth(value) <= width {
		return value
	}
	if width <= 1 {
		return "…"
	}
	var b strings.Builder
	used := 0
	for _, r := range value {
		rw := 1
		if isWide(r) {
			rw = 2
		}
		if used+rw > width-1 {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	b.WriteRune('…')
	return b.String()
}

func firstText(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isWide(r rune) bool {
	return (r >= 0x2E80 && r <= 0x9FFF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE4F) ||
		(r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6)
}

func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func padRight(s string, width int) string {
	gap := width - displayWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

func boolText(v bool) string {
	if v {
		return "是"
	}
	return "否"
}

func (m model) networkSaveModeText() string {
	if m.snapshot.Network.NMCLIAvailable {
		return "nmcli 持久化保存"
	}
	return "只读，缺少 nmcli"
}

func (m model) viewConsole(width int) string {
	if len(m.consoleLogs) == 0 {
		return ""
	}

	title := highlightStyle.Render("控制台") + "  " + dimStyle.Render("` 展开/收起")

	if m.consoleExpanded && len(m.consoleLogs) > 0 {
		idx := len(m.consoleLogs) - 1
		if m.consoleCursor >= 0 && m.consoleCursor < len(m.consoleLogs) {
			idx = m.consoleCursor
		}
		entry := m.consoleLogs[idx]
		statusText := yellowStyle.Render("执行中...")
		if entry.Done {
			if entry.Err != nil {
				statusText = redStyle.Render("✗ 失败: " + entry.Err.Error())
			} else {
				statusText = greenStyle.Render("✓ 完成")
			}
		}
		header := fmt.Sprintf("%s  %s  %s", entry.Time.Format("15:04:05"), entry.Title, statusText)
		output := entry.Output
		if output == "" {
			output = dimStyle.Render("(无输出)")
		}
		lines := []string{title, header, labelStyle.Render(entry.Command), output}
		return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
	}

	// 收起态：最近 3 条
	start := len(m.consoleLogs) - 3
	if start < 0 {
		start = 0
	}
	lines := []string{title}
	for _, entry := range m.consoleLogs[start:] {
		statusText := yellowStyle.Render("执行中...")
		if entry.Done {
			if entry.Err != nil {
				statusText = redStyle.Render("✗ 失败")
			} else {
				statusText = greenStyle.Render("✓ 完成")
			}
		}
		lines = append(lines, fmt.Sprintf("%s  %-20s %s", entry.Time.Format("15:04:05"), entry.Title, statusText))
	}
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}
