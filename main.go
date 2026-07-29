package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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

	var resolutions []string

	resolutionSelect := widget.NewSelect(
		resolutions,
		func(s string) {},
	)
	resolutionSelect.Hide()

	format := widget.NewSelect(
		[]string{"mp3", "mp4"},
		func(s string) {},
	)
	format.Hide()

	url.OnChanged = func(s string) {
		if strings.Contains(s, "youtube") || strings.Contains(s, "youtu.be") {
			format.Show()
		}
	}

	downloadStatus := widget.NewLabel("Status: None")

	downloadBtn := widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {
		downloadStatus.SetText("Status: Downloading...")
		if resolutionSelect.Selected != "" {
			go func() {
				if err := downloadVideoWithResolution(ytdlpPath.Text, outputPath.Text, format.Selected, ffmpegPath.Text, url.Text, resolutionSelect.Selected); err != nil {
					dialog.ShowError(err, w)
				} else {
					fyne.Do(func() {
						downloadStatus.SetText("Status: Download Completed")
					})
					dialog.ShowInformation("Completed", "File saved to "+outputPath.Text, w)
				}
			}()
			return
		}
		go func() {
			if err := downloadVideo(ytdlpPath.Text, outputPath.Text, format.Selected, ffmpegPath.Text, url.Text); err != nil {
				dialog.ShowError(err, w)
			} else {
				fyne.Do(func() {
					downloadStatus.SetText("Status: Download Completed")
				})
				dialog.ShowInformation("Completed", "File saved to "+outputPath.Text, w)
			}
		}()
	})
	downloadBtn.Disable()

	// only make the download button available when selected mp3 or selected mp4 and resolution
	format.OnChanged = func(s string) {
		if s == "mp3" {
			resolutionSelect.Hide()
			downloadBtn.Enable()
			return
		}
		downloadBtn.Disable()
		resolutionSelect.Show()

		resolutions = nil
		go func() {
			resolutions, err := getResolution(ytdlpPath.Text, url.Text)
			if err != nil {
				dialog.ShowError(err, w)
			}
			fyne.Do(func() {
				resolutionSelect.SetOptions(resolutions)
			})
		}()
	}

	resolutionSelect.OnChanged = func(s string) {
		downloadBtn.Enable()
	}

	downloadUI := container.NewVBox(outputPath, outputSelect, url, format, resolutionSelect, downloadBtn, downloadStatus)

	ytdlpPathSelect := widget.NewButtonWithIcon("Select yt-dlp executable", theme.FolderIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			defer reader.Close()
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func downloadVideoWithResolution(ytdlpPath string, outputPath string, format string, ffmpegPath string, url string, resolution string) error {
	cmd := exec.Command(ytdlpPath, "-P", outputPath, "-t", format, "-S", "res:"+strings.TrimSuffix(resolution, "p"), "--ffmpeg-location", ffmpegPath, "--extractor-args", "youtube:player_js_version=actual", "--no-check-certificate", url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getResolution(ytdlpPath string, url string) ([]string, error) {
	var resolution []string
	cmd := exec.Command(ytdlpPath, "-F", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return []string{}, nil
	}
	if strings.Contains(string(output), "1080p") {
		resolution = append(resolution, "1080p")
	}
	if strings.Contains(string(output), "720p") {
		resolution = append(resolution, "720p")
	}
	if strings.Contains(string(output), "480p") {
		resolution = append(resolution, "480p")
	}
	if strings.Contains(string(output), "360p") {
		resolution = append(resolution, "360p")
	}
	if strings.Contains(string(output), "144p") {
		resolution = append(resolution, "144p")
	}

	return resolution, nil
}
