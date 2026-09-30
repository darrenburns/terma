package terma

import (
	"errors"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

type imageSession struct {
	detector *imageDetector
	pointer  *pixelPointer
	kitty    *kittyImages
	sixel    *sixelImages
	hadSixel bool
	timer    *time.Timer
}

func newImageSession(t *uv.Terminal, p *pixelPointer) *imageSession {
	return &imageSession{detector: newImageDetector(t.ColorProfile()), pointer: p, kitty: newKittyImages(), sixel: newSixelImages()}
}
func (s *imageSession) wake() {
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(40*time.Millisecond, scheduleRender)
}
func (s *imageSession) close(t imageTerminal) {
	if s.timer != nil {
		s.timer.Stop()
	}
	_, _ = t.WriteString(s.kitty.cleanup())
	_ = t.Flush()
}
func (s *imageSession) reset(t imageTerminal) {
	_, _ = t.WriteString(s.kitty.cleanup())
	s.kitty = newKittyImages()
	s.hadSixel = false
}
func (s *imageSession) geometry(r *Renderer) {
	w, h, ok := s.pointer.imageCellSize()
	if !ok {
		return
	}
	if r.imageCellWidth != w || r.imageCellHeight != h {
		r.imageCellWidth, r.imageCellHeight = w, h
		if r.imageUsed {
			r.fullRenderRequired = true
		}
	}
}
func (s *imageSession) handle(ev uv.Event, r *Renderer) {
	s.detector.handle(ev)
	if e, ok := ev.(uv.KittyGraphicsEvent); ok {
		s.kitty.handle(e)
	}
	s.geometry(r)
}

// The pinned uv.Terminal leaves hardware scrolling disabled; the output-level
// scrolling regression protects that assumption when updating the dependency.
//
// present prepares outside synchronization, then brackets cleanup, Display's
// internal flush, Sixel and cursor restoration in one synchronized update.
func (s *imageSession) present(t imageTerminal, r *Renderer, debug func(), debugRows int) (err error) {
	if r.images == nil {
		debug()
		return t.Display()
	}
	s.pointer.imageGeometry = true
	now := time.Now()
	if query := s.detector.query(now); query != "" {
		if _, err = t.WriteString(query); err != nil {
			return err
		}
		if err = t.Flush(); err != nil {
			return err
		}
	}
	_, _, valid := s.pointer.imageCellSize()
	protocol := s.detector.protocol(now, valid)
	cw, ch := r.imageCellSize()
	// Recreate the presentation each frame, including after terminal Erase.
	copyImageCells(t, r.images, r.width, r.height)
	if protocol == "kitty" {
		s.kitty.paint(t, r.images, cw, ch)
	}
	debug()
	var graphics string
	if protocol == "sixel" {
		graphics = s.sixel.output(r.images, t, cw, ch, debugRows)
	}
	if protocol == "kitty" {
		batch := s.kitty.batch(now)
		if batch != "" {
			if _, err = t.WriteString(batch); err != nil {
				return err
			}
			if err = t.Flush(); err != nil {
				return err
			}
		}
	}
	if !s.detector.done || s.kitty.pending() {
		s.wake()
	}
	synchronized := s.detector.syncOutput
	if synchronized {
		// Install cleanup before BSU: even a partially failed write gets ESU.
		defer func() {
			_, endErr := t.WriteString(ansi.ResetModeSynchronizedOutput)
			err = errors.Join(err, endErr, t.Flush())
		}()
		if _, err = t.WriteString(ansi.SetModeSynchronizedOutput); err != nil {
			return err
		}
	}
	if s.hadSixel {
		// Sixel erasure is terminal dependent. A full screen erase followed by
		// complete cell presentation is conservative and leaves no removed residue.
		t.Erase()
		copyImageCells(t, r.images, r.width, r.height)
		if protocol == "kitty" {
			s.kitty.paint(t, r.images, cw, ch)
		}
		debug()
	}
	if err = t.Display(); err != nil {
		return err
	}
	if graphics != "" {
		// A failed write can still leave a partial image on the terminal.
		// Remember it before writing so the next frame erases any residue.
		s.hadSixel = true
		defer func() {
			_, restoreErr := t.WriteString("\x1b8")
			err = errors.Join(err, restoreErr)
			if !synchronized {
				err = errors.Join(err, t.Flush())
			}
		}()
		if _, err = t.WriteString("\x1b7" + graphics); err != nil {
			return err
		}
	}
	s.hadSixel = graphics != ""
	s.kitty.finish()
	return nil
}
