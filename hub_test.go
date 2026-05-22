package main

import (
	"testing"
	"time"
)

func TestSubscribeAddsChannel(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.mu.Lock()
	_, ok := h.subs[ch]
	h.mu.Unlock()
	if !ok {
		t.Fatal("Subscribe did not register channel in subs map")
	}
}

func TestUnsubscribeRemovesChannel(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.Unsubscribe(ch)
	h.mu.Lock()
	_, ok := h.subs[ch]
	h.mu.Unlock()
	if ok {
		t.Fatal("Unsubscribe did not remove channel from subs map")
	}
}

func TestBroadcastDelivers(t *testing.T) {
	h := NewHub()
	ch := make(chan []byte, 1)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	h.broadcast([]byte("hello"))

	select {
	case msg := <-ch:
		if string(msg) != "hello" {
			t.Errorf("got %q, want %q", string(msg), "hello")
		}
	default:
		t.Fatal("no message received in subscriber channel")
	}
}

func TestBroadcastDropsSlowSubscriber(t *testing.T) {
	h := NewHub()
	// unbuffered channel with no reader simulates a full/slow subscriber
	ch := make(chan []byte)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	done := make(chan struct{})
	go func() {
		h.broadcast([]byte("drop me"))
		close(done)
	}()

	select {
	case <-done:
		// broadcast returned without blocking — correct
	case <-time.After(time.Second):
		t.Fatal("broadcast blocked on slow subscriber")
	}
}
