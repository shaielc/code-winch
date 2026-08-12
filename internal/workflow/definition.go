// Package workflow defines and validates provider-neutral workflow graphs.
package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	SchemaVersion      = 1
	MaxDefinitionBytes = 1 << 20
	MaxSteps           = 1000
	MaxConcurrency     = 100
	MaxRetryAttempts   = 100
)

var stableID = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{0,127}$`)

type ValueType string

const (
	TypeString   ValueType = "string"
	TypeNumber   ValueType = "number"
	TypeBoolean  ValueType = "boolean"
	TypeObject   ValueType = "object"
	TypeArray    ValueType = "array"
	TypeRun      ValueType = "run"
	TypeEvent    ValueType = "event"
	TypeApproval ValueType = "approval"
	TypeArtifact ValueType = "artifact"
)

type Definition struct {
	SchemaVersion int                  `json:"schemaVersion"`
	DefinitionID  string               `json:"definitionId"`
	Version       string               `json:"version"`
	Inputs        map[string]ValueType `json:"inputs,omitempty"`
	Steps         []Step               `json:"steps"`
}

// InstancePins records the immutable definition and execution profiles selected
// when an instance is created. Profile values are identifiers, never secrets.
type InstancePins struct {
	DefinitionID      string `json:"definitionId"`
	DefinitionVersion string `json:"definitionVersion"`
	HarnessProfile    string `json:"harnessProfile"`
	SandboxProfile    string `json:"sandboxProfile"`
}

type Step struct {
	ID             string             `json:"id"`
	Type           string             `json:"type"`
	DependsOn      []string           `json:"dependsOn,omitempty"`
	Inputs         map[string]Binding `json:"inputs,omitempty"`
	Timeout        string             `json:"timeout,omitempty"`
	Retry          *RetryPolicy       `json:"retry,omitempty"`
	CompensateWith string             `json:"compensateWith,omitempty"`
	MaxConcurrency int                `json:"maxConcurrency,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts int    `json:"maxAttempts"`
	Backoff     string `json:"backoff,omitempty"`
}

// Binding is either a JSON literal with an explicit type or a typed reference.
// Exactly one of Literal and Ref must be populated.
type Binding struct {
	Type    ValueType       `json:"type"`
	Literal json.RawMessage `json:"literal,omitempty"`
	Ref     *Reference      `json:"ref,omitempty"`
}

type Reference struct {
	StepID string `json:"stepId,omitempty"`
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

type Diagnostic struct{ Code, Path string }

func (d Diagnostic) Error() string {
	return fmt.Sprintf("workflow definition invalid: code=%s path=%s", d.Code, d.Path)
}

func diag(code, path string) error { return Diagnostic{Code: code, Path: path} }

// ParseDefinition performs bounded, strict JSON decoding followed by graph and
// type validation. Diagnostics contain paths and stable codes but no input data.
func ParseDefinition(data []byte) (*Definition, error) {
	if len(data) > MaxDefinitionBytes {
		return nil, diag("WF_LIMIT_BYTES", "$")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d Definition
	if err := dec.Decode(&d); err != nil {
		return nil, diag("WF_INVALID_JSON", jsonErrorPath(err))
	}
	if err := ensureEOF(dec); err != nil {
		return nil, diag("WF_INVALID_JSON", "$")
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

func ensureEOF(dec *json.Decoder) error {
	var v any
	err := dec.Decode(&v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
func jsonErrorPath(err error) string {
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) && ute.Field != "" {
		return "$." + ute.Field
	}
	return "$"
}

var stepOutputs = map[string]map[string]ValueType{
	"run.start": {"run": TypeRun}, "run.send": {}, "run.stop": {},
	"event.wait": {"event": TypeEvent}, "approval.wait": {"approval": TypeApproval},
	"condition": {"result": TypeBoolean}, "parallel": {}, "foreach": {},
	"artifact.publish": {"artifact": TypeArtifact},
}

var requiredInputs = map[string]map[string]ValueType{
	"run.start": {"prompt": TypeString}, "run.send": {"run": TypeRun, "message": TypeString},
	"run.stop": {"run": TypeRun}, "event.wait": {"run": TypeRun, "eventKind": TypeString},
	"approval.wait": {"approval": TypeApproval}, "condition": {"value": TypeBoolean},
	"foreach": {"items": TypeArray}, "artifact.publish": {"artifact": TypeArtifact},
}

func (d Definition) Validate() error {
	if d.SchemaVersion != SchemaVersion {
		return diag("WF_UNSUPPORTED_VERSION", "$.schemaVersion")
	}
	if !stableID.MatchString(d.DefinitionID) {
		return diag("WF_INVALID_ID", "$.definitionId")
	}
	if strings.TrimSpace(d.Version) == "" {
		return diag("WF_REQUIRED", "$.version")
	}
	if len(d.Steps) == 0 {
		return diag("WF_REQUIRED", "$.steps")
	}
	if len(d.Steps) > MaxSteps {
		return diag("WF_LIMIT_STEPS", "$.steps")
	}
	byID := make(map[string]int, len(d.Steps))
	for i, s := range d.Steps {
		p := fmt.Sprintf("$.steps[%d]", i)
		if !stableID.MatchString(s.ID) {
			return diag("WF_INVALID_ID", p+".id")
		}
		if _, ok := byID[s.ID]; ok {
			return diag("WF_DUPLICATE_STEP", p+".id")
		}
		byID[s.ID] = i
		if _, ok := stepOutputs[s.Type]; !ok {
			return diag("WF_UNKNOWN_STEP_TYPE", p+".type")
		}
		if (s.Type == "parallel" || s.Type == "foreach") && (s.MaxConcurrency < 1 || s.MaxConcurrency > MaxConcurrency) {
			return diag("WF_CONCURRENCY_BOUND_REQUIRED", p+".maxConcurrency")
		}
		if s.Type != "parallel" && s.Type != "foreach" && s.MaxConcurrency != 0 {
			return diag("WF_FIELD_NOT_ALLOWED", p+".maxConcurrency")
		}
		if s.Timeout != "" {
			v, err := time.ParseDuration(s.Timeout)
			if err != nil || v <= 0 {
				return diag("WF_INVALID_TIMEOUT", p+".timeout")
			}
		}
		if s.Retry != nil {
			if s.Retry.MaxAttempts < 1 || s.Retry.MaxAttempts > MaxRetryAttempts {
				return diag("WF_INVALID_RETRY", p+".retry.maxAttempts")
			}
			if s.Retry.Backoff != "" {
				v, err := time.ParseDuration(s.Retry.Backoff)
				if err != nil || v < 0 {
					return diag("WF_INVALID_RETRY", p+".retry.backoff")
				}
			}
		}
		for name, typ := range requiredInputs[s.Type] {
			b, ok := s.Inputs[name]
			if !ok {
				return diag("WF_REQUIRED_INPUT", p+".inputs."+name)
			}
			if b.Type != typ {
				return diag("WF_TYPE_MISMATCH", p+".inputs."+name+".type")
			}
		}
		for name, b := range s.Inputs {
			if err := validateBinding(b, p+".inputs."+name); err != nil {
				return err
			}
		}
	}
	graph := make(map[string][]string, len(d.Steps))
	for i, s := range d.Steps {
		p := fmt.Sprintf("$.steps[%d]", i)
		for j, dep := range s.DependsOn {
			if _, ok := byID[dep]; !ok {
				return diag("WF_UNKNOWN_STEP", fmt.Sprintf("%s.dependsOn[%d]", p, j))
			}
			graph[s.ID] = append(graph[s.ID], dep)
		}
		if s.CompensateWith != "" {
			if s.CompensateWith == s.ID {
				return diag("WF_INVALID_COMPENSATION", p+".compensateWith")
			}
			if _, ok := byID[s.CompensateWith]; !ok {
				return diag("WF_UNKNOWN_STEP", p+".compensateWith")
			}
		}
		for name, b := range s.Inputs {
			if b.Ref == nil {
				continue
			}
			if b.Ref.Input != "" {
				typ, ok := d.Inputs[b.Ref.Input]
				if !ok {
					return diag("WF_UNKNOWN_INPUT", p+".inputs."+name+".ref.input")
				}
				if typ != b.Type {
					return diag("WF_TYPE_MISMATCH", p+".inputs."+name+".type")
				}
				continue
			}
			source, ok := byID[b.Ref.StepID]
			if !ok {
				return diag("WF_UNKNOWN_STEP", p+".inputs."+name+".ref.stepId")
			}
			out, ok := stepOutputs[d.Steps[source].Type][b.Ref.Output]
			if !ok {
				return diag("WF_UNKNOWN_OUTPUT", p+".inputs."+name+".ref.output")
			}
			if out != b.Type {
				return diag("WF_TYPE_MISMATCH", p+".inputs."+name+".type")
			}
			graph[s.ID] = append(graph[s.ID], b.Ref.StepID)
		}
	}
	if node := cycleNode(graph); node != "" {
		return diag("WF_CYCLE", fmt.Sprintf("$.steps[%d]", byID[node]))
	}
	return nil
}

func validateBinding(b Binding, path string) error {
	valid := map[ValueType]bool{TypeString: true, TypeNumber: true, TypeBoolean: true, TypeObject: true, TypeArray: true, TypeRun: true, TypeEvent: true, TypeApproval: true, TypeArtifact: true}
	if !valid[b.Type] {
		return diag("WF_UNKNOWN_VALUE_TYPE", path+".type")
	}
	if (len(b.Literal) == 0) == (b.Ref == nil) {
		return diag("WF_BINDING_ONE_OF", path)
	}
	if b.Ref != nil {
		if (b.Ref.StepID == "") == (b.Ref.Input == "") {
			return diag("WF_REFERENCE_ONE_OF", path+".ref")
		}
		if b.Ref.StepID != "" && b.Ref.Output == "" {
			return diag("WF_REQUIRED", path+".ref.output")
		}
		return nil
	}
	var v any
	if json.Unmarshal(b.Literal, &v) != nil || !literalMatches(v, b.Type) {
		return diag("WF_TYPE_MISMATCH", path+".literal")
	}
	return nil
}

func literalMatches(v any, typ ValueType) bool {
	switch typ {
	case TypeString:
		_, ok := v.(string)
		return ok
	case TypeNumber:
		_, ok := v.(float64)
		return ok
	case TypeBoolean:
		_, ok := v.(bool)
		return ok
	case TypeObject:
		_, ok := v.(map[string]any)
		return ok
	case TypeArray:
		_, ok := v.([]any)
		return ok
	default:
		return false
	}
}
func cycleNode(g map[string][]string) string {
	state := map[string]uint8{}
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var visit func(string) string
	visit = func(n string) string {
		if state[n] == 1 {
			return n
		}
		if state[n] == 2 {
			return ""
		}
		state[n] = 1
		for _, d := range g[n] {
			if x := visit(d); x != "" {
				return x
			}
		}
		state[n] = 2
		return ""
	}
	for _, k := range keys {
		if x := visit(k); x != "" {
			return x
		}
	}
	return ""
}

func (p InstancePins) Validate() error {
	if !stableID.MatchString(p.DefinitionID) {
		return diag("WF_INVALID_ID", "$.definitionId")
	}
	if p.DefinitionVersion == "" {
		return diag("WF_REQUIRED", "$.definitionVersion")
	}
	if !stableID.MatchString(p.HarnessProfile) {
		return diag("WF_INVALID_PROFILE", "$.harnessProfile")
	}
	if !stableID.MatchString(p.SandboxProfile) {
		return diag("WF_INVALID_PROFILE", "$.sandboxProfile")
	}
	return nil
}
