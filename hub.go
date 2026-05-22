package main

import (
	"bytes"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type Hub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
	once sync.Once
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan []byte]struct{})}
}

func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, 30) // ~2s buffer at 15fps
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	h.once.Do(func() { go h.loop() })
	return ch
}

func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *Hub) broadcast(p []byte) {
	chunk := make([]byte, len(p))
	copy(chunk, p)
	h.mu.Lock()
	for ch := range h.subs {
		select {
		case ch <- chunk:
		default: // slow subscriber: drop chunk rather than stall others
		}
	}
	h.mu.Unlock()
}

func (h *Hub) ServeStream(w http.ResponseWriter, r *http.Request) {
	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case chunk := <-ch:
			if _, err := w.Write(chunk); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func (h *Hub) loop() {
	for {
		h.runGStreamer()
		time.Sleep(time.Second)
	}
}

func (h *Hub) runGStreamer() {
	cmd := exec.Command("gst-launch-1.0",
		"-q",
		"libcamerasrc",
		"!", "video/x-raw,format=NV12,width=1280,height=720,framerate=15/1",
		"!", "videoconvert",
		"!", "jpegenc", "quality=75",
		"!", "multipartmux", "boundary=frame",
		"!", "fdsink", "fd=1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("hub: stdout pipe: %v", err)
		return
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		log.Printf("hub: gstreamer start: %v", err)
		return
	}
	defer func() {
		cmd.Wait()
		if out := stderrBuf.String(); out != "" {
			log.Printf("hub: gstreamer stderr:\n%s", out)
		}
	}()

	buf := make([]byte, 65536)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			h.broadcast(buf[:n])
		}
		if err != nil {
			return
		}
	}
}
