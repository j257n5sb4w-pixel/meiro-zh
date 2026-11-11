package main

// The proxy settings: where Meiro sends its traffic when YouTube is out of
// reach, and the card that sets it. Google's sign-in and music.youtube.com
// answer nothing on a network that cannot reach them, so the address has to
// work before anything else can.

import (
	"context"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/elianiva/meiro/internal/proxy"
	"github.com/elianiva/meiro/m3"
)

// applyProxy makes the saved proxy the one every client and child process
// uses. It runs at startup and whenever the user saves another address.
func (a *app) applyProxy() {
	if err := proxy.Set(strings.TrimSpace(a.settings.Proxy)); err != nil {
		// A file that names a proxy the app cannot read keeps the app going
		// without one, and the fields show why once they are opened.
		a.proxyError = err.Error()
		return
	}
	proxy.Install()
	a.proxyText = proxy.Address()
	a.proxyError = ""
}

// setProxy takes the address in the field, and keeps it only when it reads.
func (a *app) setProxy(value string) {
	if err := proxy.Set(strings.TrimSpace(value)); err != nil {
		a.proxyError = err.Error()
		return
	}
	proxy.Install()
	a.proxyError = ""
	a.settings.Proxy = proxy.Address()
	a.proxyText = proxy.Address()
	a.saveSettings()
	a.refresh()
}

// checkProxy asks the proxy for music.youtube.com off the main thread, so the
// user learns whether an address works before signing in through it.
func (a *app) checkProxy() {
	if proxy.Address() == "" {
		a.proxyError = "Set a proxy address first."
		return
	}
	startTask(a, &a.proxyCheck, 30*time.Second, func(context.Context) (string, error) {
		if err := proxy.Check(); err != nil {
			return "", err
		}
		return "The proxy reaches YouTube Music.", nil
	}, func(message string, err error) {
		if err != nil {
			a.proxyError = err.Error()
		} else {
			a.proxyError = ""
			a.notice = message
		}
		a.refresh()
	})
}

// proxyField is the proxy address and the buttons that save and test it. The
// sign-in dialog and the settings card draw the same one.
func (a *app) proxyField(c *ui.Context, key, label, body string) {
	sc := m3.Of(c).Scheme
	if body != "" {
		settingLabel(c, label, body)
	}
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		input := ui.TextInputBase(c.Key(key), &a.proxyText).
			Height(48).MinWidth(0).Grow(1).Padding(12).Radius(m3.Medium).Background(sc.SurfaceContainerHigh).
			TextColor(sc.OnSurface).FontSize(15).Label(label).Placeholder("http://127.0.0.1:7890")
		if input.Submitted() {
			a.setProxy(a.proxyText)
		}
		if m3.Button(c, m3.ButtonSpec{Label: "Save", Kind: m3.Tonal, Size: m3.Small40, Key: key + "-save"}).Clicked() {
			a.setProxy(a.proxyText)
		}
		if a.proxyCheck.busy() {
			m3.LoadingIndicator(c, 32, false)
		} else if m3.Button(c, m3.ButtonSpec{Label: "Test", Kind: m3.Outlined, Size: m3.Small40, Key: key + "-test"}).Clicked() {
			a.checkProxy()
		}
	})
	if a.proxyError != "" {
		m3.Text(c, m3.BodyMedium, a.proxyError).TextColor(sc.Error).MaxLines(3)
	} else if proxy.On() {
		m3.Text(c, m3.BodyMedium, "Songs and covers load through "+proxy.Address()+".").TextColor(sc.OnSurfaceVariant)
	}
}

// networkCard is where the proxy is set after sign-in: the same field, with
// what it covers.
func (a *app) networkCard(c *ui.Context) {
	a.settingsCard(c, "Network", m3.IconRefresh, func() {
		a.proxyField(c, "proxy-address", "Proxy",
			"Send everything Meiro loads through a proxy: YouTube Music, covers, lyrics, Last.fm, and the ffmpeg and yt-dlp it starts. Write it as a browser does, like http://127.0.0.1:7890 or socks5://127.0.0.1:1080. Leave it empty to go direct.")
	})
}

// proxyChip stands in the top bar while a proxy is set, so the app never
// quietly reaches the network through one without saying so.
func (a *app) proxyChip(c *ui.Context) {
	if !proxy.On() {
		return
	}
	sc := m3.Of(c).Scheme
	chip := ui.ButtonBase(c.Key("proxy-chip"))
	chip.Height(32).PaddingX(12).Radius(m3.Full).AlignItems(ui.Center).
		Label("Proxy: " + proxy.Address()).Tooltip("Proxy: " + proxy.Address()).
		Background(m3.StateFill(sc.SecondaryContainer, sc.OnSurface, chip.Hovered(), chip.Pressed(), chip.FocusVisible())).
		Cursor(ui.CursorPointer)
	chip.Children(func() {
		ui.Icon(c, m3.IconRefresh).FontSize(16).TextColor(sc.OnSecondaryContainer)
		m3.Text(c, m3.LabelSmall, "Proxy").TextColor(sc.OnSecondaryContainer)
	})
	if chip.Clicked() {
		if a.router.Location() == routeByID(routeSettings).path("") {
			// Already there: the page stays, so it takes the filter now.
			a.settingsSearch = settingsSearchState{query: "proxy", open: true}
		} else {
			// Going there closes the filter the page had, so the filter
			// waits for the visit to start.
			a.settingsSearchPending = "proxy"
			a.navigate(routeSettings)
		}
	}
}
