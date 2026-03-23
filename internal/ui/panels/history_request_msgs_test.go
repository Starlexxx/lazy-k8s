package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

func TestDeploymentsPanel_RKeyEmitsRestartRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("r key should return a command")
	}

	result := cmd()

	restartMsg, ok := result.(RestartDeploymentRequestMsg)
	if !ok {
		t.Fatalf("expected RestartDeploymentRequestMsg, got %T", result)
	}

	if restartMsg.DeploymentName != "test-deploy" {
		t.Errorf(
			"DeploymentName = %q, want %q",
			restartMsg.DeploymentName, "test-deploy",
		)
	}

	if restartMsg.Namespace != "default" {
		t.Errorf(
			"Namespace = %q, want %q",
			restartMsg.Namespace, "default",
		)
	}
}

func TestDeploymentsPanel_RKeyEmptyList(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{}
	panel.filtered = panel.deployments
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("r key with no deployments should return nil cmd")
	}
}

func TestStatefulSetsPanel_RKeyEmitsRestartRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewStatefulSetsPanel(client, styles)

	panel.statefulsets = []appsv1.StatefulSet{testStatefulSet()}
	panel.filtered = panel.statefulsets
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("r key should return a command")
	}

	result := cmd()

	restartMsg, ok := result.(RestartStatefulSetRequestMsg)
	if !ok {
		t.Fatalf("expected RestartStatefulSetRequestMsg, got %T", result)
	}

	if restartMsg.StatefulSetName != "test-sts" {
		t.Errorf(
			"StatefulSetName = %q, want %q",
			restartMsg.StatefulSetName, "test-sts",
		)
	}
}

func TestStatefulSetsPanel_RKeyEmptyList(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewStatefulSetsPanel(client, styles)

	panel.statefulsets = []appsv1.StatefulSet{}
	panel.filtered = panel.statefulsets
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("r key with no statefulsets should return nil cmd")
	}
}

func TestDaemonSetsPanel_RKeyEmitsRestartRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDaemonSetsPanel(client, styles)

	panel.daemonsets = []appsv1.DaemonSet{testDaemonSet()}
	panel.filtered = panel.daemonsets
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("r key should return a command")
	}

	result := cmd()

	restartMsg, ok := result.(RestartDaemonSetRequestMsg)
	if !ok {
		t.Fatalf("expected RestartDaemonSetRequestMsg, got %T", result)
	}

	if restartMsg.DaemonSetName != "test-ds" {
		t.Errorf(
			"DaemonSetName = %q, want %q",
			restartMsg.DaemonSetName, "test-ds",
		)
	}
}

func TestDaemonSetsPanel_RKeyEmptyList(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDaemonSetsPanel(client, styles)

	panel.daemonsets = []appsv1.DaemonSet{}
	panel.filtered = panel.daemonsets
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("r key with no daemonsets should return nil cmd")
	}
}

func TestCronJobsPanel_SKeyEmitsToggleSuspendRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{testCronJob()}
	panel.filtered = panel.cronjobs
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("S key should return a command")
	}

	result := cmd()

	suspendMsg, ok := result.(ToggleSuspendCronJobRequestMsg)
	if !ok {
		t.Fatalf(
			"expected ToggleSuspendCronJobRequestMsg, got %T",
			result,
		)
	}

	if suspendMsg.CronJobName != "test-cronjob" {
		t.Errorf(
			"CronJobName = %q, want %q",
			suspendMsg.CronJobName, "test-cronjob",
		)
	}

	if suspendMsg.CurrentSuspend {
		t.Error("expected CurrentSuspend = false for active cronjob")
	}
}

func TestCronJobsPanel_SKeyEmptyList(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{}
	panel.filtered = panel.cronjobs
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("S key with no cronjobs should return nil cmd")
	}
}

func TestCronJobsPanel_TKeyEmitsTriggerRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{testCronJob()}
	panel.filtered = panel.cronjobs
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("t key should return a command")
	}

	result := cmd()

	triggerMsg, ok := result.(TriggerCronJobRequestMsg)
	if !ok {
		t.Fatalf(
			"expected TriggerCronJobRequestMsg, got %T",
			result,
		)
	}

	if triggerMsg.CronJobName != "test-cronjob" {
		t.Errorf(
			"CronJobName = %q, want %q",
			triggerMsg.CronJobName, "test-cronjob",
		)
	}
}

func TestCronJobsPanel_TKeyEmptyList(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{}
	panel.filtered = panel.cronjobs
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("t key with no cronjobs should return nil cmd")
	}
}

func TestCronJobsPanel_SKeySuspendedCronJob(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	cj := testCronJob()
	suspended := true
	cj.Spec.Suspend = &suspended

	panel.cronjobs = []batchv1.CronJob{cj}
	panel.filtered = panel.cronjobs
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("S key should return a command")
	}

	result := cmd()

	suspendMsg, ok := result.(ToggleSuspendCronJobRequestMsg)
	if !ok {
		t.Fatalf(
			"expected ToggleSuspendCronJobRequestMsg, got %T",
			result,
		)
	}

	if !suspendMsg.CurrentSuspend {
		t.Error("expected CurrentSuspend = true for suspended cronjob")
	}
}
