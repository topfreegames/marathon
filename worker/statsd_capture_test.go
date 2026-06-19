/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 *
 * Test helper: captures the raw statsd datagrams a worker emits so tests can
 * assert the exact metric name, value, type and tags that reach the wire.
 */

package worker_test

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/worker"
)

// captureStatsd points the worker's statsd client at an ephemeral local UDP
// socket, runs fn, and returns every datagram line that was emitted. It mirrors
// production config by setting Namespace = "marathon." so assertions can match
// the fully-qualified metric name. The worker's original client is restored.
func captureStatsd(w *worker.Worker, fn func()) []string {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	defer pc.Close()

	client, err := statsd.New(pc.LocalAddr().String())
	Expect(err).NotTo(HaveOccurred())
	client.Namespace = "marathon."

	orig := w.Statsd
	w.Statsd = client
	defer func() { w.Statsd = orig }()

	var (
		mu    sync.Mutex
		lines []string
		done  = make(chan struct{})
		exit  = make(chan struct{})
	)
	go func() {
		defer close(exit)
		buf := make([]byte, 8192)
		for {
			select {
			case <-done:
				return
			default:
			}
			pc.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			n, _, rerr := pc.ReadFrom(buf)
			if n > 0 {
				mu.Lock()
				for _, l := range strings.Split(strings.TrimSpace(string(buf[:n])), "\n") {
					if l != "" {
						lines = append(lines, l)
					}
				}
				mu.Unlock()
			}
			if rerr != nil {
				continue // read deadline exceeded; loop and re-check done
			}
		}
	}()

	fn()
	// Incr sends synchronously over UDP (non-buffered client), but give the
	// reader a moment to drain the socket buffer before stopping it.
	time.Sleep(250 * time.Millisecond)
	close(done)
	<-exit

	mu.Lock()
	defer mu.Unlock()
	out := make([]string, len(lines))
	copy(out, lines)
	return out
}

// countMetric returns how many captured lines exactly equal want.
func countMetric(lines []string, want string) int {
	n := 0
	for _, l := range lines {
		if l == want {
			n++
		}
	}
	return n
}
