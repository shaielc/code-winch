// Package memory provides deterministic, concurrency-safe implementations of
// application ports for use-case and adapter contract tests.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// FailurePlan injects errors in FIFO order per operation. A nil entry means
// that invocation succeeds, making intermittent failures deterministic.
type FailurePlan struct {
	mu     sync.Mutex
	queued map[string][]error
}

func (p *FailurePlan) Inject(operation string, failures ...error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.queued == nil {
		p.queued = make(map[string][]error)
	}
	p.queued[operation] = append(p.queued[operation], failures...)
}

func (p *FailurePlan) next(operation string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	values := p.queued[operation]
	if len(values) == 0 {
		return nil
	}
	p.queued[operation] = values[1:]
	return values[0]
}

type WorkspaceRepository struct {
	Failures FailurePlan
	mu       sync.RWMutex
	items    map[domain.WorkspaceID]application.Workspace
	puts     []application.Workspace
}

func (r *WorkspaceRepository) Put(_ context.Context, value application.Workspace) error {
	if err := r.Failures.next("put"); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.items == nil {
		r.items = make(map[domain.WorkspaceID]application.Workspace)
	}
	r.items[value.ID] = value
	r.puts = append(r.puts, value)
	return nil
}
func (r *WorkspaceRepository) Get(_ context.Context, id domain.WorkspaceID) (application.Workspace, error) {
	if err := r.Failures.next("get"); err != nil {
		return application.Workspace{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.items[id]
	if !ok {
		return application.Workspace{}, application.ErrNotFound
	}
	return value, nil
}
func (r *WorkspaceRepository) PutCalls() []application.Workspace {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]application.Workspace(nil), r.puts...)
}

type versionedRun struct {
	record  application.RunRecord
	version uint64
}
type RunSaveCall struct {
	Record          application.RunRecord
	ExpectedVersion uint64
}
type RunRepository struct {
	Failures FailurePlan
	mu       sync.RWMutex
	items    map[domain.RunID]versionedRun
	saves    []RunSaveCall
}

func cloneRun(value application.RunRecord) application.RunRecord {
	value.Attempts = append([]domain.Attempt(nil), value.Attempts...)
	return value
}
func (r *RunRepository) Save(_ context.Context, value application.RunRecord, expected uint64) (uint64, error) {
	if err := r.Failures.next("save"); err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saves = append(r.saves, RunSaveCall{cloneRun(value), expected})
	if r.items == nil {
		r.items = make(map[domain.RunID]versionedRun)
	}
	current, exists := r.items[value.ID]
	if (!exists && expected != 0) || (exists && current.version != expected) {
		return 0, application.ErrConflict
	}
	next := expected + 1
	r.items[value.ID] = versionedRun{cloneRun(value), next}
	return next, nil
}
func (r *RunRepository) Get(_ context.Context, id domain.RunID) (application.RunRecord, uint64, error) {
	if err := r.Failures.next("get"); err != nil {
		return application.RunRecord{}, 0, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.items[id]
	if !ok {
		return application.RunRecord{}, 0, application.ErrNotFound
	}
	return cloneRun(value.record), value.version, nil
}
func (r *RunRepository) SaveCalls() []RunSaveCall {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := append([]RunSaveCall(nil), r.saves...)
	for i := range out {
		out[i].Record = cloneRun(out[i].Record)
	}
	return out
}

type AppendCall struct {
	RunID            domain.RunID
	ExpectedSequence uint64
	Events           []application.UnsequencedEvent
}
type EventStore struct {
	Failures FailurePlan
	mu       sync.RWMutex
	events   map[domain.RunID][]protocol.Event
	appends  []AppendCall
}

func cloneBytes(v []byte) []byte { return append([]byte(nil), v...) }
func cloneUnsequenced(values []application.UnsequencedEvent) []application.UnsequencedEvent {
	out := append([]application.UnsequencedEvent(nil), values...)
	for i := range out {
		out[i].Payload = cloneBytes(out[i].Payload)
		out[i].Extensions = cloneExtensions(out[i].Extensions)
	}
	return out
}
func cloneExtensions(values map[string][]byte) map[string][]byte {
	if values == nil {
		return nil
	}
	out := make(map[string][]byte, len(values))
	for k, v := range values {
		out[k] = cloneBytes(v)
	}
	return out
}
func cloneEvents(values []protocol.Event) []protocol.Event {
	out := append([]protocol.Event(nil), values...)
	for i := range out {
		out[i].Payload = cloneBytes(out[i].Payload)
		out[i].Extensions = cloneRaw(out[i].Extensions)
	}
	return out
}
func cloneRaw(values map[string]json.RawMessage) map[string]json.RawMessage {
	if values == nil {
		return nil
	}
	out := make(map[string]json.RawMessage, len(values))
	for k, v := range values {
		out[k] = cloneBytes(v)
	}
	return out
}

func (s *EventStore) Append(_ context.Context, runID domain.RunID, expected uint64, values []application.UnsequencedEvent) ([]protocol.Event, error) {
	if err := s.Failures.next("append"); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appends = append(s.appends, AppendCall{runID, expected, cloneUnsequenced(values)})
	if s.events == nil {
		s.events = make(map[domain.RunID][]protocol.Event)
	}
	current := s.events[runID]
	if uint64(len(current)) != expected {
		return nil, fmt.Errorf("%w: run=%s expected_sequence=%d actual_sequence=%d", application.ErrConflict, runID, expected, len(current))
	}
	added := make([]protocol.Event, len(values))
	seen := make(map[string]struct{}, len(current)+len(values))
	for _, event := range current {
		seen[event.EventID] = struct{}{}
	}
	for i, value := range values {
		id := value.EventID.String()
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("%w: run=%s event_id=%s", application.ErrConflict, runID, id)
		}
		seen[id] = struct{}{}
		ext := make(map[string]json.RawMessage, len(value.Extensions))
		for k, v := range value.Extensions {
			ext[k] = cloneBytes(v)
		}
		added[i] = protocol.Event{EventID: id, RunID: runID.String(), Sequence: expected + uint64(i) + 1, OccurredAt: value.OccurredAt.Time(), Kind: value.Kind, SchemaVersion: value.SchemaVersion, Source: value.Source, Sensitivity: value.Sensitivity, Payload: cloneBytes(value.Payload), Extensions: ext}
	}
	s.events[runID] = append(current, cloneEvents(added)...)
	return cloneEvents(added), nil
}
func (s *EventStore) Read(_ context.Context, runID domain.RunID, after uint64, limit int) ([]protocol.Event, error) {
	if err := s.Failures.next("read"); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := s.events[runID]
	if after >= uint64(len(values)) || limit <= 0 {
		return []protocol.Event{}, nil
	}
	end := len(values)
	if end > int(after)+limit {
		end = int(after) + limit
	}
	return cloneEvents(values[int(after):end]), nil
}
func (s *EventStore) AppendCalls() []AppendCall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]AppendCall(nil), s.appends...)
	for i := range out {
		out[i].Events = cloneUnsequenced(out[i].Events)
	}
	return out
}

type OutboxPublisher struct {
	Failures FailurePlan
	mu       sync.RWMutex
	messages map[domain.CommandID]application.OutboxMessage
	calls    []application.OutboxMessage
}

func cloneMessage(v application.OutboxMessage) application.OutboxMessage {
	v.Payload = cloneBytes(v.Payload)
	return v
}
func (p *OutboxPublisher) Publish(_ context.Context, value application.OutboxMessage) error {
	if err := p.Failures.next("publish"); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, cloneMessage(value))
	if p.messages == nil {
		p.messages = make(map[domain.CommandID]application.OutboxMessage)
	}
	if _, ok := p.messages[value.ID]; !ok {
		p.messages[value.ID] = cloneMessage(value)
	}
	return nil
}
func (p *OutboxPublisher) Calls() []application.OutboxMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := append([]application.OutboxMessage(nil), p.calls...)
	for i := range out {
		out[i] = cloneMessage(out[i])
	}
	return out
}

// Published returns the deduplicated external effects, keyed by message ID.
func (p *OutboxPublisher) Published() []application.OutboxMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]application.OutboxMessage, 0, len(p.messages))
	for _, value := range p.messages {
		out = append(out, cloneMessage(value))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

type SecretReferenceStore struct {
	Failures FailurePlan
	mu       sync.RWMutex
	values   map[application.SecretReference][]byte
	calls    []application.SecretReference
}

func (s *SecretReferenceStore) Put(_ context.Context, ref application.SecretReference, value application.ResolvedSecret) error {
	if err := s.Failures.next("put"); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.values == nil {
		s.values = make(map[application.SecretReference][]byte)
	}
	s.values[ref] = cloneBytes(value.Bytes)
	return nil
}
func (s *SecretReferenceStore) Resolve(_ context.Context, ref application.SecretReference) (application.ResolvedSecret, error) {
	if err := s.Failures.next("resolve"); err != nil {
		return application.ResolvedSecret{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, ref)
	value, ok := s.values[ref]
	if !ok {
		return application.ResolvedSecret{}, application.ErrNotFound
	}
	return application.ResolvedSecret{Bytes: cloneBytes(value)}, nil
}
func (s *SecretReferenceStore) ResolveCalls() []application.SecretReference {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]application.SecretReference(nil), s.calls...)
}

type RunnerGateway struct {
	Failures FailurePlan
	mu       sync.RWMutex
	messages map[string]protocol.RunnerMessage
	calls    []protocol.RunnerMessage
}

func cloneRunner(v protocol.RunnerMessage) protocol.RunnerMessage {
	v.Payload = cloneBytes(v.Payload)
	return v
}
func (g *RunnerGateway) Send(_ context.Context, value protocol.RunnerMessage) error {
	if err := g.Failures.next("send"); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls = append(g.calls, cloneRunner(value))
	if g.messages == nil {
		g.messages = make(map[string]protocol.RunnerMessage)
	}
	if _, ok := g.messages[value.CommandID]; !ok {
		g.messages[value.CommandID] = cloneRunner(value)
	}
	return nil
}
func (g *RunnerGateway) Calls() []protocol.RunnerMessage {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := append([]protocol.RunnerMessage(nil), g.calls...)
	for i := range out {
		out[i] = cloneRunner(out[i])
	}
	return out
}

// Delivered returns the deduplicated external effects, keyed by command ID.
func (g *RunnerGateway) Delivered() []protocol.RunnerMessage {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]protocol.RunnerMessage, 0, len(g.messages))
	for _, value := range g.messages {
		out = append(out, cloneRunner(value))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CommandID < out[j].CommandID })
	return out
}

// Clock advances only when instructed, so tests never need sleeps.
type Clock struct {
	mu  sync.RWMutex
	now domain.Timestamp
}

func NewClock(now domain.Timestamp) *Clock { return &Clock{now: now} }
func (c *Clock) Now() domain.Timestamp     { c.mu.RLock(); defer c.mu.RUnlock(); return c.now }
func (c *Clock) Set(now domain.Timestamp)  { c.mu.Lock(); defer c.mu.Unlock(); c.now = now }

// IDSource returns preloaded IDs and panics with a content-free diagnostic when
// a test consumes more IDs than it arranged.
type IDSource struct {
	mu            sync.Mutex
	WorkspaceIDs  []domain.WorkspaceID
	RunIDs        []domain.RunID
	AttemptIDs    []domain.AttemptID
	EventIDs      []domain.EventID
	CommandIDs    []domain.CommandID
	ArtifactIDs   []domain.ArtifactID
	CredentialIDs []domain.CredentialID
	WorkflowIDs   []domain.WorkflowID
}

func pop[T any](values *[]T, name string) T {
	if len(*values) == 0 {
		panic("memory ID source exhausted: " + name)
	}
	v := (*values)[0]
	*values = (*values)[1:]
	return v
}
func (s *IDSource) NewWorkspaceID() domain.WorkspaceID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.WorkspaceIDs, "workspace")
}
func (s *IDSource) NewRunID() domain.RunID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.RunIDs, "run")
}
func (s *IDSource) NewAttemptID() domain.AttemptID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.AttemptIDs, "attempt")
}
func (s *IDSource) NewEventID() domain.EventID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.EventIDs, "event")
}
func (s *IDSource) NewCommandID() domain.CommandID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.CommandIDs, "command")
}
func (s *IDSource) NewArtifactID() domain.ArtifactID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.ArtifactIDs, "artifact")
}
func (s *IDSource) NewCredentialID() domain.CredentialID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.CredentialIDs, "credential")
}
func (s *IDSource) NewWorkflowID() domain.WorkflowID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pop(&s.WorkflowIDs, "workflow")
}
