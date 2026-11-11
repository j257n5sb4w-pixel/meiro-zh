package lang

import "testing"

func TestTranslateWhole(t *testing.T) {
	Set(Chinese)
	cases := map[string]string{
		"Settings":        "设置",
		"Sign in":         "登录",
		"Nothing playing": "未在播放",
	}
	for in, want := range cases {
		if got := T(in); got != want {
			t.Errorf("T(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTranslateBuilt(t *testing.T) {
	Set(Chinese)
	cases := map[string]string{
		"Added “Song” to your library":    "已添加“Song”保存到媒体库",
		"Could not like “Song”: HTTP 403": "无法喜欢“Song”： HTTP 403",
		"Removed 3 tracks from the queue": "已从队列中移除 3 首歌曲",
		"Playing Song Name":               "正在播放 Song Name",
		"3 plays are kept on this device, for the stats. Meiro keeps the latest 500.": "3 条播放记录保存在本设备，用于统计。Meiro 最多保留 500 条。",
	}
	for in, want := range cases {
		if got := T(in); got != want {
			t.Errorf("T(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnglishPassesThrough(t *testing.T) {
	Set(English)
	if got := T("Settings"); got != "Settings" {
		t.Errorf("T(Settings) = %q in English", got)
	}
	if got := T("Added “x”"); got != "Added “x”" {
		t.Errorf("T changed a built message in English: %q", got)
	}
}

func TestUnknownStaysEnglish(t *testing.T) {
	Set(Chinese)
	const name = "A song title that is not interface text"
	if got := T(name); got != name {
		t.Errorf("T(%q) = %q", name, got)
	}
}
