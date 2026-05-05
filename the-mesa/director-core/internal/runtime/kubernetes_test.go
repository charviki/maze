package runtime

import (
	"context"
	"os"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewKubernetesRuntime_InitError 验证非 K8s 环境下构造函数返回错误而非静默创建 nil-clientset 实例
func TestNewKubernetesRuntime_InitError(t *testing.T) {
	kube := config.KubernetesConfig{
		Namespace:   "maze",
		DirectorCoreAddr: "director-core.maze.svc.cluster.local:8080",
	}
	_, err := NewKubernetesRuntime(kube, config.WorkspaceConfig{}, logutil.NewNop())
	if err == nil {
		t.Fatal("期望非 K8s 环境下返回错误，实际返回 nil")
	}
}

// newFakeClientRuntime 使用 fake K8s client 创建 KubernetesRuntime
func newFakeClientRuntime() *KubernetesRuntime {
	kube := config.KubernetesConfig{
		Namespace:       "maze",
		DirectorCoreAddr: "director-core.maze.svc.cluster.local:8080",
		AgentImageTag:   "latest",
		ImagePullPolicy: "IfNotPresent",
		PVCSize:         "10Gi",
		VolumeType:      "pvc",
	}
	return &KubernetesRuntime{
		kube:      kube,
		workspace: config.WorkspaceConfig{MountDir: "/data"},
		logger:    logutil.NewNop(),
		clientset: fake.NewClientset(),
	}
}

// --- fake client 正常路径 ---

func TestKubernetesRuntime_DeployHost(t *testing.T) {
	rt := newFakeClientRuntime()
	ctx := context.Background()

	spec := &protocol.HostDeploySpec{
		Name:  "my-agent",
		Tools: []string{"claude", "go"},
	}

	resp, err := rt.DeployHost(ctx, spec, "")
	if err != nil {
		t.Fatalf("DeployHost 失败: %v", err)
	}

	if resp.Name != "my-agent" {
		t.Errorf("Name 不匹配: got %q, want %q", resp.Name, "my-agent")
	}
	if resp.Status != "running" {
		t.Errorf("Status 不匹配: got %q, want %q", resp.Status, "running")
	}

	// 无 Dockerfile 时使用默认镜像名 maze-agent:latest
	wantImage := "maze-agent:latest"
	if resp.ImageTag != wantImage {
		t.Errorf("ImageTag 不匹配: got %q, want %q", resp.ImageTag, wantImage)
	}

	// 验证 PVC 已创建
	appName := "maze-agent-my-agent"
	pvc, err := rt.clientset.CoreV1().PersistentVolumeClaims("maze").Get(ctx, appName+"-data", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("查询 PVC 失败: %v", err)
	}
	if pvc.Name != appName+"-data" {
		t.Errorf("PVC 名称不匹配: got %q, want %q", pvc.Name, appName+"-data")
	}

	// 验证 Deployment 已创建
	deploy, err := rt.clientset.AppsV1().Deployments("maze").Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("查询 Deployment 失败: %v", err)
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		t.Fatal("Deployment 缺少容器")
	}
	if deploy.Spec.Template.Spec.Containers[0].Image != wantImage {
		t.Errorf("容器镜像不匹配: got %q, want %q", deploy.Spec.Template.Spec.Containers[0].Image, wantImage)
	}

	// 验证 Service 已创建
	svc, err := rt.clientset.CoreV1().Services("maze").Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("查询 Service 失败: %v", err)
	}
	if len(svc.Spec.Ports) == 0 || svc.Spec.Ports[0].Port != 8080 {
		t.Errorf("Service 端口不正确: %v", svc.Spec.Ports)
	}
}

func TestKubernetesRuntime_RemoveHost(t *testing.T) {
	rt := newFakeClientRuntime()
	ctx := context.Background()
	ns := "maze"
	appName := "maze-agent-rm-test"

	// 先手动创建 PVC、Deployment、Service，再测试 RemoveHost 删除
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: appName + "-data", Namespace: ns},
	}
	if _, err := rt.clientset.CoreV1().PersistentVolumeClaims(ns).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
		t.Fatalf("预创建 PVC 失败: %v", err)
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: ns},
	}
	if _, err := rt.clientset.AppsV1().Deployments(ns).Create(ctx, deploy, metav1.CreateOptions{}); err != nil {
		t.Fatalf("预创建 Deployment 失败: %v", err)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: ns},
	}
	if _, err := rt.clientset.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		t.Fatalf("预创建 Service 失败: %v", err)
	}

	// 执行删除
	err := rt.RemoveHost(ctx, "rm-test")
	if err != nil {
		t.Fatalf("RemoveHost 失败: %v", err)
	}

	// 验证三资源均被删除
	if _, err := rt.clientset.CoreV1().Services(ns).Get(ctx, appName, metav1.GetOptions{}); err == nil {
		t.Error("Service 未被删除")
	}
	if _, err := rt.clientset.AppsV1().Deployments(ns).Get(ctx, appName, metav1.GetOptions{}); err == nil {
		t.Error("Deployment 未被删除")
	}
	if _, err := rt.clientset.CoreV1().PersistentVolumeClaims(ns).Get(ctx, appName+"-data", metav1.GetOptions{}); err == nil {
		t.Error("PVC 未被删除")
	}
}

func TestKubernetesRuntime_RemoveHost_Idempotent(t *testing.T) {
	rt := newFakeClientRuntime()

	err := rt.RemoveHost(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("删除不存在的 Host 应无错误，实际: %v", err)
	}
}

func TestKubernetesRuntime_RemoveHost_HostPathCleansDir(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := tmpDir + "/agents/test-host"
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	kube := config.KubernetesConfig{
		Namespace:   "maze",
		DirectorCoreAddr: "director-core.maze.svc.cluster.local:8080",
		VolumeType:  "hostpath",
	}
	rt := &KubernetesRuntime{
		kube:      kube,
		workspace: config.WorkspaceConfig{MountDir: tmpDir},
		logger:    logutil.NewNop(),
		clientset: fake.NewClientset(),
	}

	err := rt.RemoveHost(context.Background(), "test-host")
	if err != nil {
		t.Fatalf("RemoveHost 失败: %v", err)
	}

	if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
		t.Errorf("hostPath 目录应被清理，但仍存在: %s", agentDir)
	}
}

func TestKubernetesRuntime_InspectHost(t *testing.T) {
	rt := newFakeClientRuntime()
	ctx := context.Background()
	ns := "maze"

	// 创建一个带 label 的 Pod 模拟 Agent 运行时
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "maze-agent-inspect-test-abc123",
			Namespace: ns,
			Labels: map[string]string{
				"app":             "maze-agent",
				"maze-agent-name": "inspect-test",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "agent", Image: "maze-agent:latest"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	if _, err := rt.clientset.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("预创建 Pod 失败: %v", err)
	}

	info, err := rt.InspectHost(ctx, "inspect-test")
	if err != nil {
		t.Fatalf("InspectHost 失败: %v", err)
	}
	if info.Status != "running" {
		t.Errorf("Status 不匹配: got %q, want %q", info.Status, "running")
	}
	if info.Image != "maze-agent:latest" {
		t.Errorf("Image 不匹配: got %q, want %q", info.Image, "maze-agent:latest")
	}
}

func TestKubernetesRuntime_InspectHost_NotFound(t *testing.T) {
	rt := newFakeClientRuntime()

	_, err := rt.InspectHost(context.Background(), "no-such-agent")
	if err == nil {
		t.Fatal("期望查询不存在的 Host 时返回错误，实际返回 nil")
	}
}

func TestMapPodPhase(t *testing.T) {
	tests := []struct {
		phase corev1.PodPhase
		want  string
	}{
		{corev1.PodRunning, "running"},
		{corev1.PodPending, "created"},
		{corev1.PodSucceeded, "exited"},
		{corev1.PodFailed, "exited"},
		{corev1.PodPhase("UnknownPhase"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			got := mapPodPhase(tt.phase)
			if got != tt.want {
				t.Errorf("mapPodPhase(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}
