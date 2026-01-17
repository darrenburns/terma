package main

import t "terma"

type App struct {
	activeTab t.Signal[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Height: t.Flex(1),
		Children: []t.Widget{
			// Tab bar
			t.Row{
				Spacing: 1,
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsHV(1, 0),
				},
				Children: []t.Widget{
					a.tabButton("home", "Home", ctx),
					a.tabButton("settings", "Settings", ctx),
					a.tabButton("profile", "Profile", ctx),
				},
			},
			// Content
			t.Switcher{
				Active: a.activeTab.Get(),
				Height: t.Flex(1),
				Children: map[string]t.Widget{
					"home":     HomeView{},
					"settings": SettingsView{},
					"profile":  ProfileView{},
				},
			},
		},
	}
}

func (a *App) tabButton(key, label string, ctx t.BuildContext) t.Button {
	theme := ctx.Theme()
	isActive := a.activeTab.Get() == key

	style := t.Style{Padding: t.EdgeInsetsXY(2, 0)}
	if isActive {
		style.BackgroundColor = theme.Primary
		style.ForegroundColor = theme.Primary.AutoText()
	}

	return t.Button{
		ID:      key,
		Label:   label,
		Style:   style,
		OnPress: func() { a.activeTab.Set(key) },
	}
}

func main() {
	t.Run(&App{activeTab: t.NewSignal("home")})
}
