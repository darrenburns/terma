package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	t "github.com/darrenburns/terma"
)

// FileInfo holds metadata about a file or directory.
type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

// FileSearchDemo demonstrates using CommandPalette for file/directory searching.
type FileSearchDemo struct {
	palette      *t.CommandPaletteState
	selectedFile t.Signal[string]
	rootPath     string
	files        []FileInfo
}

func NewFileSearchDemo() *FileSearchDemo {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	app := &FileSearchDemo{
		selectedFile: t.NewSignal("No file selected"),
		rootPath:     cwd,
	}

	// Scan the directory for files
	app.files = app.scanDirectory(cwd, 3) // max depth 3
	items := app.filesToItems(app.files)
	app.palette = t.NewCommandPaletteState("Search Files", items)

	return app
}

// scanDirectory recursively scans a directory up to maxDepth levels.
func (a *FileSearchDemo) scanDirectory(root string, maxDepth int) []FileInfo {
	var files []FileInfo

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip hidden files/directories (starting with .)
		name := d.Name()
		if strings.HasPrefix(name, ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate depth
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil // Skip root itself
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth >= maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, FileInfo{
			Name:  name,
			Path:  relPath,
			IsDir: d.IsDir(),
			Size:  info.Size(),
		})

		return nil
	})

	// Sort: directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Path) < strings.ToLower(files[j].Path)
	})

	return files
}

// filesToItems converts FileInfo slice to CommandPaletteItem slice.
func (a *FileSearchDemo) filesToItems(files []FileInfo) []t.CommandPaletteItem {
	items := make([]t.CommandPaletteItem, 0, len(files))

	for _, f := range files {
		file := f // Capture for closure
		item := t.CommandPaletteItem{
			Label:      file.Name,
			FilterText: file.Path, // Allow searching by full path
			Data:       file,
		}

		if file.IsDir {
			item.Hint = "dir"
			// Directories have children (lazy-loaded)
			item.ChildrenTitle = file.Path
			item.Children = func() []t.CommandPaletteItem {
				subPath := filepath.Join(a.rootPath, file.Path)
				subFiles := a.scanDirectory(subPath, 2)
				return a.filesToItems(subFiles)
			}
		} else {
			item.Hint = formatSize(file.Size)
			item.Action = func() {
				a.selectedFile.Set(file.Path)
				a.palette.Close(false)
			}
		}

		items = append(items, item)
	}

	return items
}

// formatSize formats a file size in human-readable form.
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fGB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1fMB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1fKB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func (a *FileSearchDemo) togglePalette() {
	if a.palette.Visible.Peek() {
		a.palette.Close(false)
		return
	}
	a.palette.Open()
}

func (a *FileSearchDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+p", Name: "Search files", Action: a.togglePalette},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *FileSearchDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Stack{
		Children: []t.Widget{
			t.Dock{
				Top: []t.Widget{
					t.Column{
						Style: t.Style{
							Padding: t.EdgeInsetsAll(2),
						},
						Spacing: 1,
						Children: []t.Widget{
							t.Text{
								Content: "File Search Demo",
								Style: t.Style{
									ForegroundColor: theme.TextOnPrimary,
									BackgroundColor: theme.Primary,
									Padding:         t.EdgeInsetsXY(2, 0),
									Bold:            true,
								},
							},
							t.Text{
								Content: "Press Ctrl+P to search files",
								Style: t.Style{
									ForegroundColor: theme.TextMuted,
								},
							},
							t.Text{
								Content: "Root: " + a.rootPath,
								Style: t.Style{
									ForegroundColor: theme.TextMuted,
								},
							},
							t.Spacer{Height: t.Cells(1)},
							t.Row{
								Spacing: 1,
								Children: []t.Widget{
									t.Text{
										Content: "Selected:",
										Style: t.Style{
											ForegroundColor: theme.Text,
											Bold:            true,
										},
									},
									t.Text{
										Content: a.selectedFile.Get(),
										Style: t.Style{
											ForegroundColor: theme.Accent,
										},
									},
								},
							},
						},
					},
				},
				Bottom: []t.Widget{
					t.KeybindBar{
						Style: t.Style{
							BackgroundColor: theme.Surface,
							Padding:         t.EdgeInsetsXY(1, 0),
						},
					},
				},
				Body: t.Spacer{},
			},
			t.CommandPalette{
				ID:          "file-search",
				State:       a.palette,
				Placeholder: "Search files by name or path...",
				Position:    t.FloatPositionTopCenter,
				Offset:      t.Offset{Y: 2},
				RenderItem:  a.renderFileItem(ctx),
			},
		},
	}
}

// renderFileItem provides a custom renderer for file items.
func (a *FileSearchDemo) renderFileItem(ctx t.BuildContext) func(item t.CommandPaletteItem, active bool, match t.MatchResult) t.Widget {
	theme := ctx.Theme()

	return func(item t.CommandPaletteItem, active bool, match t.MatchResult) t.Widget {
		fileInfo, ok := item.Data.(FileInfo)
		if !ok {
			// Fallback for items without FileInfo data
			return t.Text{Content: item.Label}
		}

		// Choose icon based on file type
		icon := "  " // File icon
		if fileInfo.IsDir {
			icon = "  " // Folder icon
		}

		// Style based on active state
		itemStyle := t.Style{
			Padding: t.EdgeInsetsXY(1, 0),
			Width:   t.Flex(1),
		}
		labelStyle := t.Style{ForegroundColor: theme.Text}
		hintStyle := t.Style{ForegroundColor: theme.TextMuted}
		iconStyle := t.Style{ForegroundColor: theme.Accent}

		if active {
			itemStyle.BackgroundColor = theme.ActiveCursor
			labelStyle.ForegroundColor = theme.SelectionText
			hintStyle.ForegroundColor = theme.SelectionText
			iconStyle.ForegroundColor = theme.SelectionText
		}

		// Build label with match highlighting
		var labelWidget t.Widget
		if match.Matched && len(match.Ranges) > 0 {
			labelWidget = t.Text{
				Spans: t.HighlightSpans(item.Label, match.Ranges, t.MatchHighlightStyle(theme)),
				Style: labelStyle,
			}
		} else {
			labelWidget = t.Text{
				Content: item.Label,
				Style:   labelStyle,
			}
		}

		return t.Row{
			Style:      itemStyle,
			CrossAlign: t.CrossAxisCenter,
			Children: []t.Widget{
				t.Text{Content: icon, Style: iconStyle},
				labelWidget,
				t.Spacer{Width: t.Flex(1)},
				t.Text{Content: item.Hint, Style: hintStyle},
			},
		}
	}
}

func main() {
	app := NewFileSearchDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
