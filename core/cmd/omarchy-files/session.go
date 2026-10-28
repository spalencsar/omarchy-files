package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/duxlabs-dev/dux-go/ipc/params"
	"github.com/duxlabs-dev/dux-go/log"
	"github.com/spalencsar/omarchy-files/core/internal/ipc"
	"github.com/spalencsar/omarchy-files/core/internal/sandbox"
)

type event struct {
	Topic string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type reply struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
	Code   string          `json:"code"`
}

type callError struct {
	Code    string
	Message string
}

func (e *callError) Error() string { return e.Message }

type session struct {
	conn    net.Conn
	events  <-chan event
	release func()

	mu      sync.Mutex
	nextID  int
	pending map[int]chan reply
	readErr error
}

func runningSocket() (string, bool) {
	if socket := os.Getenv("OMARCHY_FILES_SOCKET"); socket != "" {
		return socket, true
	}
	socket, err := ipc.FindRunningSocket()
	return socket, err == nil
}

func dialDaemon() (*session, error) {
	socket, ok := runningSocket()
	if !ok {
		return nil, errors.New("omarchy-files is not running")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(conn)
	if _, err := r.ReadBytes('\n'); err != nil {
		conn.Close()
		return nil, fmt.Errorf("read capabilities: %w", err)
	}
	return newSession(conn, r, func() {}), nil
}

func openSession(build func(bus *ipc.EventBus) (ipc.Deps, func())) (*session, error) {
	if s, err := dialDaemon(); err == nil {
		return s, nil
	}

	sandbox.ExportHostDirs(sandbox.Current())
	bus := ipc.NewEventBus()
	deps, closeDeps := build(bus)
	deps.Bus = bus

	client, server := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		serveLocal(ctx, server, deps)
	}()
	return newSession(client, bufio.NewReader(client), func() {
		cancel()
		server.Close()
		<-done
		closeDeps()
	}), nil
}

func serveLocal(ctx context.Context, conn net.Conn, deps ipc.Deps) {
	defer conn.Close()
	writer := ipc.NewConnWriter(conn)
	sub := deps.Bus.NewSubscriber(ctx, writer)
	defer sub.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var wg sync.WaitGroup
	defer wg.Wait()
	for scanner.Scan() {
		var req ipc.Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		if req.Method == "subscribe" {
			topics := params.StringSlice(req.Params, "topics")
			sub.Subscribe(topics...)
			writer.Respond(req.ID, map[string]any{"topics": sub.Topics()})
			if deps.Files != nil {
				deps.Files.Attach(topics)
			}
			if deps.Search != nil {
				deps.Search.Attach(topics)
			}
			continue
		}
		wg.Go(func() { ipc.Route(ctx, writer, req, deps) })
	}
	if err := scanner.Err(); err != nil {
		log.Debugf("local session: %v", err)
	}
}

func newSession(conn net.Conn, r *bufio.Reader, release func()) *session {
	in := make(chan event)
	s := &session{conn: conn, pending: map[int]chan reply{}, events: queue(in), release: release}
	go s.read(r, in)
	return s
}

func (s *session) Close() {
	s.conn.Close()
	s.release()
}

func (s *session) read(r *bufio.Reader, events chan<- event) {
	defer close(events)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			s.fail(err)
			return
		}
		var probe struct {
			Event *string `json:"event"`
		}
		if json.Unmarshal(line, &probe) == nil && probe.Event != nil {
			var ev event
			if json.Unmarshal(line, &ev) == nil {
				events <- ev
			}
			continue
		}
		var rep reply
		if err := json.Unmarshal(line, &rep); err != nil {
			continue
		}
		s.mu.Lock()
		ch := s.pending[rep.ID]
		delete(s.pending, rep.ID)
		s.mu.Unlock()
		if ch != nil {
			ch <- rep
		}
	}
}

func (s *session) fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readErr = err
	for id, ch := range s.pending {
		ch <- reply{ID: id, Error: "connection closed: " + err.Error()}
		delete(s.pending, id)
	}
}

// Unbounded, so a reply never waits behind events nobody has read yet.
func queue(in <-chan event) <-chan event {
	out := make(chan event)
	go func() {
		defer close(out)
		var buffered []event
		for in != nil || len(buffered) > 0 {
			var send chan<- event
			var next event
			if len(buffered) > 0 {
				send = out
				next = buffered[0]
			}
			select {
			case ev, ok := <-in:
				if !ok {
					in = nil
					continue
				}
				buffered = append(buffered, ev)
			case send <- next:
				buffered = buffered[1:]
			}
		}
	}()
	return out
}

func (s *session) call(method string, p map[string]any, out any) error {
	ch := make(chan reply, 1)
	s.mu.Lock()
	if s.readErr != nil {
		s.mu.Unlock()
		return s.readErr
	}
	s.nextID++
	id := s.nextID
	s.pending[id] = ch
	s.mu.Unlock()

	data, err := json.Marshal(ipc.Request{ID: id, Method: method, Params: p})
	if err != nil {
		return err
	}
	if _, err := s.conn.Write(append(data, '\n')); err != nil {
		return err
	}

	rep := <-ch
	if rep.Error != "" {
		return &callError{Code: rep.Code, Message: rep.Error}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(rep.Result, out)
}

func (s *session) subscribe(topics ...string) error {
	return s.call("subscribe", map[string]any{"topics": topics}, nil)
}

func withSession(build func(bus *ipc.EventBus) (ipc.Deps, func()), fn func(s *session) error) error {
	s, err := openSession(build)
	if err != nil {
		return err
	}
	defer s.Close()
	return fn(s)
}

func absolutePaths(args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		out = append(out, abs)
	}
	return out, nil
}
