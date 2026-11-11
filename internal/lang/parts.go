package lang

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// part is a piece of a message the app assembles at runtime, as in
// "Added “" + title + "” to “" + playlist, or "Could not " + err.
type part struct {
	en, zh string
	// anywhere says the piece may be translated wherever it appears; a
	// piece that opens with a word may only start a message, because the
	// same words inside a song's title are that song's name.
	anywhere bool
}

// parts is what messages built at runtime are translated by, longest first so
// that "Added “" wins over "Added ".
var parts = build()

func build() []part {
	table := []part{
		{en: "Could not remove dislike from “", zh: "无法取消“"},
		{en: "Remove dislike from ", zh: "取消不喜欢 "},
		{en: "Could not unsubscribe from ", zh: "无法取消订阅 "},
		{en: "Could not subscribe to ", zh: "无法订阅 "},
		{en: "Could not download “", zh: "无法下载“"},
		{en: "Could not create “", zh: "无法创建“"},
		{en: "Could not switch to “", zh: "无法切换到“"},
		{en: "Could not unlike “", zh: "无法取消喜欢“"},
		{en: "Could not dislike “", zh: "无法标记为不喜欢“"},
		{en: "Could not like “", zh: "无法喜欢“"},
		{en: "Could not remove “", zh: "无法移除“"},
		{en: "Could not add “", zh: "无法添加“"},
		{en: "Could not open “", zh: "无法打开“"},
		{en: "Could not restore your YouTube Music session: ", zh: "无法恢复你的 YouTube Music 会话："},
		{en: " is not in the system credential store", zh: " 不在系统凭据存储中"},
		{en: "Could not save ", zh: "无法保存 "},
		{en: "Could not remove ", zh: "无法移除 "},
		{en: "Could not delete ", zh: "无法删除 "},
		{en: "Unsubscribed from ", zh: "已取消订阅 "},
		{en: "Subscribed to ", zh: "已订阅 "},
		{en: "Importing from ", zh: "从 "},
		{en: " did not work. Paste the Cookie header instead.", zh: " 导入失败，请改为粘贴 Cookie 请求头。"},
		{en: ". Meiro remembers your choice.", zh: "。Meiro 会记住你的选择。"},
		{en: "”: YouTube did not name that channel.", zh: "”：YouTube 没有给出该频道的名称。"},
		{en: "” is already in “", zh: "”已经在“"},
		{en: "” from playlist: ", zh: "” 从播放列表中移除："},
		{en: "” from playlist", zh: "” 从播放列表中移除"},
		{en: "” to your library", zh: "”保存到媒体库"},
		{en: " from your library", zh: " 从媒体库移除"},
		{en: " to your library", zh: " 保存到媒体库"},
		{en: "” to “", zh: "”到“"},
		{en: "”: ", zh: "”： "},
		{en: "Added “", zh: "已添加“"},
		{en: "Created “", zh: "已创建“"},
		{en: "Deleted “", zh: "已删除“"},
		{en: "Removed “", zh: "已移除“"},
		{en: "It will hold “", zh: "它将包含“"},
		{en: "Skipped ", zh: "已跳过 "},
		{en: ": YouTube marks it as made with AI", zh: "：YouTube 将其标记为 AI 生成"},
		{en: " will play next", zh: " 将在接下来播放"},
		{en: " to the queue", zh: " 到队列"},
		{en: " from the queue", zh: " 从队列中移除"},
		{en: "Playing ", zh: "正在播放 "},
		{en: "Downloading ", zh: "正在下载 "},
		{en: "Added ", zh: "已添加 "},
		{en: "Saved ", zh: "已保存 "},
		{en: "Removed ", zh: "已移除 "},
		{en: "Add to ", zh: "添加到 "},
		{en: "More options for ", zh: "更多选项："},
		{en: "Go to album: ", zh: "前往专辑："},
		{en: "Go to artist: ", zh: "前往歌手："},
		{en: "Playback speed: ", zh: "播放速度："},
		{en: "Play ", zh: "播放 "},
		{en: "Like ", zh: "喜欢 "},
		{en: "Unlike ", zh: "取消喜欢 "},
		{en: "Dislike ", zh: "标记为不喜欢 "},
		{en: "Remove ", zh: "移除 "},
		{en: " from playback indicator", zh: "（来自播放指示器）"},
		{en: " is already ", zh: " 已被占用"},
		{en: " waiting to be sent.", zh: " 条等待发送。"},
		{en: " days", zh: " 天"},
		{en: " plays", zh: " 次播放"},
		{en: " play", zh: " 次播放"},
	}
	for i := range table {
		first, _ := utf8.DecodeRuneInString(table[i].en)
		table[i].anywhere = !unicode.IsLetter(first)
	}
	sort.SliceStable(table, func(i, j int) bool { return len(table[i].en) > len(table[j].en) })
	return table
}

// patterns are the messages that carry a number, which no table of whole
// strings can hold.
var patterns = []struct {
	re   *regexp.Regexp
	text string
}{
	{regexp.MustCompile(`^(\d+) plays are kept on this device, for the stats\. Meiro keeps the latest (\d+)\.$`), "$1 条播放记录保存在本设备，用于统计。Meiro 最多保留 $2 条。"},
	{regexp.MustCompile(`^Removed (\d+) tracks from the queue$`), "已从队列中移除 $1 首歌曲"},
	{regexp.MustCompile(`^Something went wrong while (.+)\. The details are in the log\. Copying them helps when you report the problem\.$`), "处理$1时出错。详情在日志里，复制它有助于你报告问题。"},
	{regexp.MustCompile(`^Songs you listen to in Meiro are counted here, for the last (\d+) days\.$`), "Meiro 中播放的歌曲会在这里统计，范围为最近 $1 天。"},
	{regexp.MustCompile(`^Songs and covers load through (.+)\.$`), "歌曲和封面通过 $1 加载。"},
}

// translate builds the Chinese for s out of the whole-string table, the
// number patterns, and the pieces of messages built at runtime.
func translate(s string) string {
	if translated, ok := words[s]; ok {
		return translated
	}
	for _, pattern := range patterns {
		if pattern.re.MatchString(s) {
			return pattern.re.ReplaceAllString(s, pattern.text)
		}
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		matched := false
		for _, p := range parts {
			if !p.anywhere && i > 0 {
				continue
			}
			if !strings.HasPrefix(s[i:], p.en) {
				continue
			}
			// A piece that ends in a word may not cut one in half, as " play"
			// would inside "playing", and one that starts with a word may not
			// begin inside another.
			if wholeWord(s, i, len(p.en)) {
				b.WriteString(p.zh)
				i += len(p.en)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+size])
		i += size
	}
	return b.String()
}

// wholeWord reports whether the match of n bytes at i in s starts and ends on a
// word boundary, considering only the sides the piece itself touches.
func wholeWord(s string, i, n int) bool {
	start := s[i : i+n]
	first, _ := utf8.DecodeRuneInString(start)
	last, _ := utf8.DecodeLastRuneInString(start)
	if i > 0 && isLetter(first) {
		before, _ := utf8.DecodeLastRuneInString(s[:i])
		if isLetter(before) {
			return false
		}
	}
	if rest := s[i+n:]; rest != "" && isLetter(last) {
		after, _ := utf8.DecodeRuneInString(rest)
		if isLetter(after) {
			return false
		}
	}
	return true
}

func isLetter(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
