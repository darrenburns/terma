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
	// The sixel images on screen, as last drawn. When sixelUnknown is set,
	// what's on screen isn't known (a write failed part-way), so the next
	// frame clears it.
	sixelShown   map[sixelDraw]bool
	sixelUnknown bool
	timer        *time.Timer
}

func newImageSession(t *uv.Terminal, p *pixelPointer) *imageSession {
	s := &imageSession{detector: newImageDetector(t.ColorProfile()), pointer: p, kitty: newKittyImages(), sixel: newSixelImages()}
	// Encode and quantize off the event loop.
	s.kitty.worker.async = true
	s.sixel.worker.async = true
	return s
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

// reset deletes every image from the terminal, for when it has been handed
// to another program that may have changed or deleted them.
func (s *imageSession) reset(t imageTerminal) {
	_, _ = t.WriteString(s.kitty.reset())
	s.erased()
}

// erased records that the screen has been cleared, taking any sixel images
// with it. Kitty images outlive it: their placements go with the cells that
// show them, which are drawn again.
func (s *imageSession) erased() {
	s.sixelShown = nil
	s.sixelUnknown = false
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

// sixelChanges compares the sixel images a frame needs with those on screen.
// Sixel pixels stay until something clears them, and how writing cells over
// them clears them depends on the terminal, so an image gone or changed means
// clearing the screen and drawing everything again (erase). Otherwise only
// the images not yet on screen are drawn (fresh), and those already there are
// left alone.
func (s *imageSession) sixelChanges(draws []sixelDraw) (erase bool, fresh []sixelDraw) {
	needed := make(map[sixelDraw]bool, len(draws))
	for _, d := range draws {
		needed[d] = true
	}
	erase = s.sixelUnknown
	for d := range s.sixelShown {
		if !needed[d] {
			erase = true
			break
		}
	}
	if erase {
		return true, draws
	}
	for _, d := range draws {
		if !s.sixelShown[d] {
			fresh = append(fresh, d)
		}
	}
	return false, fresh
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
	var draws []sixelDraw
	if protocol == "sixel" {
		draws = s.sixel.draws(r.images, t, cw, ch, debugRows)
	}
	erase, fresh := s.sixelChanges(draws)
	graphics, drawn := s.sixel.output(fresh)
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
	if erase {
		// Sixel erasure is terminal dependent. A full screen erase followed by
		// complete cell presentation is conservative and leaves no removed residue.
		t.Erase()
		copyImageCells(t, r.images, r.width, r.height)
		if protocol == "kitty" {
			s.kitty.paint(t, r.images, cw, ch)
		}
		debug()
		s.erased()
	}
	if err = t.Display(); err != nil {
		return err
	}
	if graphics != "" {
		// A failed write can still leave a partial image on the terminal.
		// Mark the screen unknown before writing so the next frame clears it.
		s.sixelUnknown = true
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
		s.sixelUnknown = false
		if s.sixelShown == nil {
			s.sixelShown = make(map[sixelDraw]bool)
		}
		for _, d := range drawn {
			s.sixelShown[d] = true
		}
	}
	s.kitty.finish()
	return nil
}
