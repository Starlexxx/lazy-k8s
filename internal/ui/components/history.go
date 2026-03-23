package components

import (
	"fmt"
	"sync"
	"time"
)

// OperationType categorizes each recorded operation.
type OperationType int

const (
	OpScaleDeployment OperationType = iota
	OpScaleStatefulSet
	OpRestartDeployment
	OpRestartStatefulSet
	OpRestartDaemonSet
	OpRollbackDeployment
	OpDeleteResource
	OpPortForward
	OpExec
	OpSuspendCronJob
	OpResumeCronJob
	OpEditHPAMin
	OpEditHPAMax
	OpEditResource
	OpTriggerCronJob
)

// UndoData captures previous state needed to reverse an operation.
// Uses a flat struct to avoid type assertions (forcetypeassert linter).
type UndoData struct {
	PreviousReplicas    int32
	PreviousSuspend     bool
	PreviousMinReplicas int32
	PreviousMaxReplicas int32
}

// OperationRecord represents a single logged operation.
type OperationRecord struct {
	ID        int
	Type      OperationType
	Timestamp time.Time
	Resource  string // e.g. "deployment/nginx"
	Namespace string
	Message   string // human-readable summary
	Undoable  bool
	UndoData  UndoData
	Undone    bool
}

// HistoryStore is an append-only in-memory log of operations.
// Safe for concurrent use from tea.Cmd goroutines.
type HistoryStore struct {
	mu      sync.Mutex
	records []OperationRecord
	nextID  int
}

func NewHistoryStore() *HistoryStore {
	return &HistoryStore{
		records: make([]OperationRecord, 0),
	}
}

// Add appends a record and returns its assigned ID.
func (h *HistoryStore) Add(rec OperationRecord) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	rec.ID = h.nextID
	h.nextID++

	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now()
	}

	h.records = append(h.records, rec)

	return rec.ID
}

// Records returns all entries newest-first.
func (h *HistoryStore) Records() []OperationRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]OperationRecord, len(h.records))

	for i, rec := range h.records {
		out[len(h.records)-1-i] = rec
	}

	return out
}

// Get returns the record with the given ID, if it exists.
func (h *HistoryStore) Get(id int) (OperationRecord, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if id < 0 || id >= len(h.records) {
		return OperationRecord{}, false
	}

	return h.records[id], true
}

// MarkUndone flags a record as having been undone.
func (h *HistoryStore) MarkUndone(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if id >= 0 && id < len(h.records) {
		h.records[id].Undone = true
	}
}

// Len returns the total number of recorded operations.
func (h *HistoryStore) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.records)
}

// operationLabel returns a human-readable label for the operation type.
func operationLabel(op OperationType) string {
	labels := map[OperationType]string{
		OpScaleDeployment:    "Scale Deployment",
		OpScaleStatefulSet:   "Scale StatefulSet",
		OpRestartDeployment:  "Restart Deployment",
		OpRestartStatefulSet: "Restart StatefulSet",
		OpRestartDaemonSet:   "Restart DaemonSet",
		OpRollbackDeployment: "Rollback Deployment",
		OpDeleteResource:     "Delete",
		OpPortForward:        "Port Forward",
		OpExec:               "Exec",
		OpSuspendCronJob:     "Suspend CronJob",
		OpResumeCronJob:      "Resume CronJob",
		OpEditHPAMin:         "Edit HPA Min",
		OpEditHPAMax:         "Edit HPA Max",
		OpEditResource:       "Edit Resource",
		OpTriggerCronJob:     "Trigger CronJob",
	}

	if label, ok := labels[op]; ok {
		return label
	}

	return fmt.Sprintf("Unknown(%d)", op)
}
