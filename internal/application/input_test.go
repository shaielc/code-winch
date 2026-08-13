package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/shaielc/code-winch/internal/domain"
)

type inputStoreFake struct {
	mu      sync.Mutex
	byKey   map[string]InputResult
	accepts int
}

func (s *inputStoreFake) AcceptInput(_ context.Context, a InputAcceptance) (InputResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if got, ok := s.byKey[a.Request.IdempotencyKey]; ok {
		return got, nil
	}
	r := InputResult{CommandID: a.Request.CommandID, RunID: a.Request.RunID, Kind: a.Request.Kind, Accepted: true}
	s.byKey[a.Request.IdempotencyKey] = r
	s.accepts++
	return r, nil
}

type capabilitiesFake struct {
	value InputCapabilities
	err   error
}

func (f capabilitiesFake) InputCapabilities(context.Context, domain.RunID) (InputCapabilities, error) {
	return f.value, f.err
}

type authorizerFake struct {
	allowed bool
	calls   int
}

func (a *authorizerFake) AuthorizeRawTerminalInput(context.Context, string, domain.RunID) error {
	a.calls++
	if !a.allowed {
		return errors.New("denied")
	}
	return nil
}

func inputID[T any](t *testing.T, parse func(string) (T, error), suffix string) T {
	t.Helper()
	v, e := parse("00000000-0000-0000-0000-" + suffix)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func request(t *testing.T, kind InputKind) InputRequest {
	p := InputPayload{}
	switch kind {
	case InputText:
		p.Text = "content"
	case InputTerminalBytes:
		p.Bytes = []byte("content")
	case InputResize:
		p.Rows = 24
		p.Columns = 80
	}
	return InputRequest{CommandID: inputID(t, domain.ParseCommandID, "000000000001"), RunID: inputID(t, domain.ParseRunID, "000000000002"), IdempotencyKey: "key", ActorID: "actor", Kind: kind, Payload: p, ExpectedState: domain.RunStateRunning}
}

func TestInputDuplicateConcurrentReturnsOriginalResult(t *testing.T) {
	store := &inputStoreFake{byKey: map[string]InputResult{}}
	svc := NewInputService(store, capabilitiesFake{value: InputCapabilities{State: domain.RunStateRunning, Modes: map[InputKind]bool{InputText: true}}}, nil)
	const callers = 32
	req := request(t, InputText)
	results := make(chan InputResult, callers)
	errs := make(chan error, callers)
	for range callers {
		go func() { got, err := svc.Accept(context.Background(), req); results <- got; errs <- err }()
	}
	var first domain.CommandID
	for range callers {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
		got := <-results
		if first.IsZero() {
			first = got.CommandID
		}
		if got.CommandID != first {
			t.Fatalf("different replay result: %s != %s", got.CommandID, first)
		}
	}
	if store.accepts != 1 {
		t.Fatalf("durable accepts=%d, want 1", store.accepts)
	}
}

func TestInputCapabilityAndAuthorizationMatrix(t *testing.T) {
	tests := []struct {
		name      string
		kind      InputKind
		modes     map[InputKind]bool
		allowed   bool
		code      string
		authCalls int
	}{
		{"text supported", InputText, map[InputKind]bool{InputText: true}, false, "", 0},
		{"interrupt unsupported", InputInterrupt, map[InputKind]bool{}, true, InputErrorUnsupported, 0},
		{"resize supported", InputResize, map[InputKind]bool{InputResize: true}, true, "", 0},
		{"raw denied separately", InputTerminalBytes, map[InputKind]bool{InputTerminalBytes: true}, false, InputErrorUnauthorized, 1},
		{"raw authorized", InputTerminalBytes, map[InputKind]bool{InputTerminalBytes: true}, true, "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &inputStoreFake{byKey: map[string]InputResult{}}
			auth := &authorizerFake{allowed: tt.allowed}
			svc := NewInputService(store, capabilitiesFake{value: InputCapabilities{State: domain.RunStateRunning, Modes: tt.modes}}, auth)
			_, err := svc.Accept(context.Background(), request(t, tt.kind))
			var inputErr *InputError
			if tt.code == "" && err != nil {
				t.Fatal(err)
			}
			if tt.code != "" && (!errors.As(err, &inputErr) || inputErr.Code != tt.code) {
				t.Fatalf("error=%v want code=%s", err, tt.code)
			}
			if auth.calls != tt.authCalls {
				t.Fatalf("authorization calls=%d", auth.calls)
			}
		})
	}
}
