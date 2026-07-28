package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("YTDLP-GUI")
	w := a.NewWindow("YTDLP GUI")
	w.Resize(fyne.NewSize(640, 480))
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	ytdlpPath := widget.NewEntry()
	ytdlpPath.SetText(filepath.Join(exeDir, "yt-dlp.exe"))
	ffmpegPath := widget.NewEntry()
	ffmpegPath.SetText("C:\\ffmpeg\\bin")

	outputPath := widget.NewEntry()
	outputPath.SetText(exeDir)

	outputSelect := widget.NewButtonWithIcon("Select output folder", theme.FolderIcon(), func() {
		fd := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil || lu == nil {
				return
			}
			outputPath.SetText(lu.Path())
		}, w)
		fd.Show()
	})

	url := widget.NewEntry()
	url.SetPlaceHolder("Enter YouTube URL here")

	format := widget.NewSelectEntry(
		[]string{"mp3", "mp4"},
	)
	format.SetPlaceHolder("Select format")

	downloadStatus := widget.NewLabel("Status: None")

	downloadBtn := widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {
		downloadStatus.SetText("Status: Downloading...")
		go func() {
			if err := downloadVideo(ytdlpPath.Text, outputPath.Text, format.Text, ffmpegPath.Text, url.Text); err != nil {
				dialog.ShowError(err, w)
			} else {
				downloadStatus.SetText("Status: Download Completed")
				dialog.ShowInformation("Completed", "File saved to "+outputPath.Text, w)
			}
		}()
	})

	downloadUI := container.NewVBox(outputPath, outputSelect, url, format, downloadBtn, downloadStatus)

	ytdlpPathSelect := widget.NewButtonWithIcon("Select yt-dlp executable", theme.FolderIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			ytdlpPath.SetText(reader.URI().Path())
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})
	ffmpegPathSelect := widget.NewButtonWithIcon("Select ffmpeg folder", theme.FolderIcon(), func() {
		fd := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil || lu == nil {
				return
			}
			ffmpegPath.SetText(lu.Path())
		}, w)
		fd.Show()
	})
	configUI := container.NewVBox(ytdlpPath, ytdlpPathSelect, ffmpegPath, ffmpegPathSelect)

	tabs := container.NewAppTabs(
		container.NewTabItem("Download", downloadUI),
		container.NewTabItem("Config", configUI),
	)
	w.SetContent(tabs)
	w.ShowAndRun()
}

func downloadVideo(ytdlpPath string, outputPath string, format string, ffmpegPath string, url string) error {
	cmd := exec.Command(ytdlpPath, "-P", outputPath, "-t", format, "--ffmpeg-location", ffmpegPath, "--extractor-args", "youtube:player_js_version=actual", "--no-check-certificate", url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
