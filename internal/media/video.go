/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/abema/go-mp4"
	"github.com/disintegration/imaging"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func decodeVideo(r io.Reader, contentType string) (*mediaMeta, error) {
	if config.GetMediaVideoFFMPEGEnabled() {
		return decodeVideoFFMPEG(r, contentType)
	}
	return decodeVideoNative(r, contentType)
}

func decodeVideoFFMPEG(r io.Reader, contentType string) (*mediaMeta, error) {
	prog := "ffprobe"
	args := []string{
		"pipe:",
		"-select_streams", "v",
		"-show_entries", "stream=width,height",
		"-of", "json",
	}
	cmd := exec.Command(prog, args...)
	cmd.Stdin = r
	out := bytes.NewBuffer(make([]byte, 0, 512))
	cmd.Stdout = out
	cmdErrc := make(chan error, 1)
	cmdErrOut, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	go func() {
		out, err := ioutil.ReadAll(cmdErrOut)
		if err != nil {
			cmdErrc <- err
			return
		}
		cmd.Wait()
		if cmd.ProcessState.Success() {
			cmdErrc <- nil
			return
		}
		cmdErrc <- fmt.Errorf("metadata probe subprocess failed:\n%s", out)
	}()
	select {
	case err := <-cmdErrc:
		if err != nil {
			return nil, err
		}
	case <-time.After(time.Second):
		return nil, fmt.Errorf("timeout during metadata probe process")
	}
	streamInfo := &struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}{}
	if err := json.Unmarshal(out.Bytes(), &streamInfo); err != nil {
		return nil, fmt.Errorf("failed parsing metadata: %w", err)
	}
	if len(streamInfo.Streams) == 0 {
		return nil, fmt.Errorf("media container did not contain any video streams")
	}
	s := streamInfo.Streams[0]
	return &mediaMeta{
		width:  s.Width,
		height: s.Height,
		size:   s.Height * s.Width,
		aspect: float64(s.Width) / float64(s.Height),
	}, nil
}
func decodeVideoNative(r io.Reader, contentType string) (*mediaMeta, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading video data: %w", err)
	}
	buff := bytes.NewReader(data)
	var (
		height int
		width  int
	)
	_, err = mp4.ReadBoxStructure(buff, func(h *mp4.ReadHandle) (interface{}, error) {
		switch h.BoxInfo.Type {
		case mp4.BoxTypeTkhd():
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			// update MessageData
			tkhd := box.(*mp4.Tkhd)
			h := int(tkhd.GetHeight())
			if h > height {
				height = h
			}
			w := int(tkhd.GetWidth())
			if w > width {
				width = w
			}
		}
		if h.BoxInfo.IsSupportedType() {
			return h.Expand()
		}
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing video data: %w", err)
	}

	log.Debugf("detected media size: %dx%d", width, height)

	return &mediaMeta{
		width:  width,
		height: height,
		size:   height * width,
		aspect: float64(width) / float64(height),
	}, nil
}

func extractFromVideo(r io.Reader) (image.Image, error) {
	if config.GetMediaVideoFFMPEGEnabled() {
		return extractFromVideoFFMPEG(r)
	}
	return extractFromVideoNative(r)
}
func extractFromVideoFFMPEG(r io.Reader) (image.Image, error) {
	prog := "ffmpeg"
	args := []string{
		"-i", "pipe:",
		"-vf", "thumbnail=n=10",
		"-frames:v", "1",
		"-f", "image2pipe",
		"-c:v", "png",
		"pipe:1",
	}
	cmd := exec.Command(prog, args...)
	cmd.Stdin = r
	out := bytes.NewBuffer(make([]byte, 0, 1024))
	cmd.Stdout = out
	cmdErrc := make(chan error, 1)
	cmdErrOut, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	go func() {
		out, err := ioutil.ReadAll(cmdErrOut)
		if err != nil {
			cmdErrc <- err
			return
		}
		cmd.Wait()
		if cmd.ProcessState.Success() {
			cmdErrc <- nil
			return
		}
		cmdErrc <- fmt.Errorf("thumbnail subprocess failed:\n%s", out)
	}()
	select {
	case err := <-cmdErrc:
		if err != nil {
			return nil, err
		}
	case <-time.After(time.Second * 5):
		return nil, fmt.Errorf("timeout during thumbnailing process")
	}
	return imaging.Decode(out, imaging.AutoOrientation(true))
}
func extractFromVideoNative(r io.Reader) (image.Image, error) {
	// As of writing this, no go native way exists to extract a thumbnail from a
	// video so we fall back by just returning a static file
	file, err := os.Open(path.Join(config.GetWebAssetBaseDir(), "logo.png"))
	if err != nil {
		return nil, err
	}
	img, err := imaging.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, err
}
