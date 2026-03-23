package components

import (
	"testing"
	"time"
)

func TestHistoryStore_AddAndGet(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore()

	id := store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "nginx",
		Namespace: "default",
		Message:   "Scaled nginx from 1 to 3 replicas",
		Undoable:  true,
		UndoData:  UndoData{PreviousReplicas: 1},
	})

	if id != 0 {
		t.Errorf("expected first ID = 0, got %d", id)
	}

	if store.Len() != 1 {
		t.Errorf("expected Len() = 1, got %d", store.Len())
	}

	rec, ok := store.Get(id)
	if !ok {
		t.Fatal("expected Get to return true")
	}

	if rec.Resource != "nginx" {
		t.Errorf("expected Resource = nginx, got %s", rec.Resource)
	}

	if rec.UndoData.PreviousReplicas != 1 {
		t.Errorf(
			"expected PreviousReplicas = 1, got %d",
			rec.UndoData.PreviousReplicas,
		)
	}
}

func TestHistoryStore_RecordsNewestFirst(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "first",
		Namespace: "default",
		Timestamp: time.Now().Add(-time.Minute),
	})

	store.Add(OperationRecord{
		Type:      OpRestartDeployment,
		Resource:  "second",
		Namespace: "default",
	})

	records := store.Records()
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Newest first
	if records[0].Resource != "second" {
		t.Errorf("expected first record = second, got %s", records[0].Resource)
	}

	if records[1].Resource != "first" {
		t.Errorf("expected second record = first, got %s", records[1].Resource)
	}
}

func TestHistoryStore_MarkUndone(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore()

	id := store.Add(OperationRecord{
		Type:      OpSuspendCronJob,
		Resource:  "cleanup",
		Namespace: "default",
		Undoable:  true,
	})

	store.MarkUndone(id)

	rec, ok := store.Get(id)
	if !ok {
		t.Fatal("expected Get to return true")
	}

	if !rec.Undone {
		t.Error("expected record to be marked as undone")
	}
}

func TestHistoryStore_GetInvalidID(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore()

	_, ok := store.Get(-1)
	if ok {
		t.Error("expected Get(-1) to return false")
	}

	_, ok = store.Get(999)
	if ok {
		t.Error("expected Get(999) to return false")
	}
}

func TestHistoryStore_TimestampAutoSet(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore()

	before := time.Now()

	store.Add(OperationRecord{
		Type:      OpDeleteResource,
		Resource:  "pod/test",
		Namespace: "default",
	})

	after := time.Now()

	rec, _ := store.Get(0)
	if rec.Timestamp.Before(before) || rec.Timestamp.After(after) {
		t.Error("expected timestamp to be auto-set to current time")
	}
}

func TestOperationLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op    OperationType
		label string
	}{
		{OpScaleDeployment, "Scale Deployment"},
		{OpRestartDaemonSet, "Restart DaemonSet"},
		{OpDeleteResource, "Delete"},
		{OpSuspendCronJob, "Suspend CronJob"},
		{OpResumeCronJob, "Resume CronJob"},
		{OpEditResource, "Edit Resource"},
		{OpTriggerCronJob, "Trigger CronJob"},
	}

	for _, tt := range tests {
		if got := operationLabel(tt.op); got != tt.label {
			t.Errorf(
				"operationLabel(%d) = %q, want %q",
				tt.op, got, tt.label,
			)
		}
	}
}

func TestOperationLabel_Unknown(t *testing.T) {
	t.Parallel()

	label := operationLabel(OperationType(999))
	if label != "Unknown(999)" {
		t.Errorf("expected Unknown(999), got %q", label)
	}
}
