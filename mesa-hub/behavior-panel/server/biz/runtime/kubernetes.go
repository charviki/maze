package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/builder"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesRuntime 通过 Kubernetes API 实现的容器运行时
type KubernetesRuntime struct {
	kube      config.KubernetesConfig
	workspace config.WorkspaceConfig
	logger    logutil.Logger
	clientset kubernetes.Interface
}

// NewKubernetesRuntime 创建 KubernetesRuntime
// K8s 客户端在构造时初始化，失败时 clientset 为 nil，后续方法将返回错误
func NewKubernetesRuntime(kube config.KubernetesConfig, workspace config.WorkspaceConfig, logger logutil.Logger) *KubernetesRuntime {
	var clientset kubernetes.Interface

	var restConfig *rest.Config
	var err error
	if kube.Kubeconfig != "" {
		// 从 kubeconfig 文件加载
		restConfig, err = clientcmd.BuildConfigFromFlags("", kube.Kubeconfig)
	} else {
		// in-cluster 模式
		restConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		logger.Errorf("kubernetes client init failed: %v", err)
		return &KubernetesRuntime{kube: kube, workspace: workspace, logger: logger}
	}

	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("kubernetes clientset create failed: %v", err)
	}

	return &KubernetesRuntime{
		kube:      kube,
		workspace: workspace,
		logger:    logger,
		clientset: clientset,
	}
}

// checkClient 确认 K8s 客户端已初始化
func (k *KubernetesRuntime) checkClient() error {
	if k.clientset == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}
	return nil
}

// imageExistsLocally 检查指定镜像是否已存在于本地 docker 中
func (k *KubernetesRuntime) imageExistsLocally(imageName string) bool {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// checkDockerfileHash 从镜像 label 中读取 dockerfile-hash 与期望值比较
func (k *KubernetesRuntime) checkDockerfileHash(imageName, expectedHash string) bool {
	cmd := exec.Command("docker", "inspect", "--format",
		"{{index .Config.Labels \"maze.dockerfile-hash\"}}", imageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == expectedHash
}

// buildDockerImage 使用 docker build 从 Dockerfile 内容构建镜像
// 复用 Docker 模式相同的动态构建逻辑，K8s 模式下只是把 docker run 换成创建 Deployment
func (k *KubernetesRuntime) buildDockerImage(spec *protocol.HostDeploySpec, dockerfileContent string) (string, error) {
	imageName := fmt.Sprintf("maze-agent:%s", spec.Name)
	expectedHash := extractDockerfileHash(dockerfileContent)

	// 优先检查 Host 专属镜像是否已存在且 hash 匹配
	if k.imageExistsLocally(imageName) {
		if k.checkDockerfileHash(imageName, expectedHash) {
			k.logger.Infof("image %s already exists, skip build", imageName)
			return imageName, nil
		}
		// hash 不匹配，删除旧镜像触发重建
		k.logger.Infof("image %s hash mismatch, rebuilding", imageName)
		exec.Command("docker", "rmi", imageName).Run()
	}

	// 检查工具组合镜像是否已存在且 hash 匹配
	comboTag := builder.ToolsetImageTag(spec.Tools)
	if k.imageExistsLocally(comboTag) {
		if k.checkDockerfileHash(comboTag, expectedHash) {
			k.logger.Infof("combo image %s exists, tagging as %s", comboTag, imageName)
			cmd := exec.Command("docker", "tag", comboTag, imageName)
			if cmd.Run() == nil {
				return imageName, nil
			}
		}
		// hash 不匹配，删除旧缓存触发重建
		k.logger.Infof("combo image %s hash mismatch, rebuilding", comboTag)
		exec.Command("docker", "rmi", comboTag).Run()
	}

	// 获取构建槽位，防止重建风暴
	buildSemaphore <- struct{}{}
	defer func() { <-buildSemaphore }()

	// 创建临时构建上下文目录
	tmpDir, err := os.MkdirTemp("", "maze-agent-build-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入 Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return "", fmt.Errorf("write dockerfile: %w", err)
	}

	// 执行 docker build，启用 BuildKit 加速构建
	cmd := exec.Command("docker", "build", "-f", dockerfilePath, "-t", imageName, "--cache-from", imageName, tmpDir)
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker build failed: %s: %w", string(output), err)
	}

	k.logger.Infof("built image %s for host %s", imageName, spec.Name)

	// 构建完成后打上组合标签，供后续相同组合的 Host 复用
	tagCmd := exec.Command("docker", "tag", imageName, comboTag)
	tagCmd.Run()

	return imageName, nil
}

// removeDockerImage 清理动态构建的 Agent 镜像
func (k *KubernetesRuntime) removeDockerImage(name string) {
	imageName := fmt.Sprintf("maze-agent:%s", name)
	cmd := exec.Command("docker", "rmi", imageName, "-f")
	_ = cmd.Run()
}

// DeployHost 部署 Host 到 Kubernetes 集群：docker build → 创建 PVC → Deployment → Service
func (k *KubernetesRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	if err := k.checkClient(); err != nil {
		return nil, err
	}

	ns := k.kube.Namespace
	appName := fmt.Sprintf("maze-agent-%s", spec.Name)

	// 第一步：动态构建镜像（复用 Docker 模式的 Dockerfile 生成逻辑）
	var image string
	if dockerfileContent != "" {
		var err error
		image, err = k.buildDockerImage(spec, dockerfileContent)
		if err != nil {
			return nil, fmt.Errorf("build image: %w", err)
		}
	} else {
		// 无 Dockerfile 时使用 Agent 基础镜像（测试场景）
		image = "maze-agent:latest"
	}

	// 第二步：创建持久卷（PVC 模式时创建，hostPath 模式时跳过）
	if k.kube.VolumeType == "pvc" {
		if err := k.createPVC(ctx, ns, appName, spec.Name); err != nil {
			return nil, fmt.Errorf("create pvc: %w", err)
		}
	}

	// 第三步：创建 Deployment
	if err := k.createDeployment(ctx, ns, appName, spec, image); err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}

	// 第四步：创建 Service（集群内 DNS 可达）
	if err := k.createService(ctx, ns, appName, spec.Name); err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &protocol.CreateHostResponse{
		Name:        spec.Name,
		Tools:       spec.Tools,
		ImageTag:    image,
		ContainerID: appName,
		Status:      "running",
	}, nil
}

// createPVC 创建持久卷声明，已存在时跳过
func (k *KubernetesRuntime) createPVC(ctx context.Context, ns, appName, hostName string) error {
	pvcName := appName + "-data"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: ns,
			Labels: map[string]string{
				"app":             "maze-agent",
				"maze-agent-name": hostName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(k.kube.PVCSize),
				},
			},
		},
	}
	// 配置 StorageClass：空字符串表示使用集群默认
	if k.kube.PVCStorageClass != "" {
		pvc.Spec.StorageClassName = &k.kube.PVCStorageClass
	}

	_, err := k.clientset.CoreV1().PersistentVolumeClaims(ns).Create(ctx, pvc, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		k.logger.Infof("pvc %s already exists, skip", pvcName)
		return nil
	}
	return err
}

// createDeployment 创建 Deployment 工作负载
func (k *KubernetesRuntime) createDeployment(ctx context.Context, ns, appName string, spec *protocol.HostDeploySpec, image string) error {
	managerAddr := k.kube.ManagerAddr
	if managerAddr == "" {
		managerAddr = "agent-manager:8080"
	}
	// 确保 ManagerAddr 带有 http:// 前缀
	if !strings.HasPrefix(managerAddr, "http") {
		managerAddr = "http://" + managerAddr
	}

	externalAddr := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", appName, ns)

	// 构建容器环境变量
	envs := []corev1.EnvVar{
		{Name: "AGENT_NAME", Value: spec.Name},
		{Name: "AGENT_EXTERNAL_ADDR", Value: externalAddr},
		{Name: "AGENT_ADVERTISED_ADDR", Value: externalAddr},
		{Name: "AGENT_GRPC_ADDR", Value: fmt.Sprintf("%s.%s.svc.cluster.local:9090", appName, ns)},
		{Name: "AGENT_CONTROLLER_ADDR", Value: managerAddr},
		// Agent 自身 API 鉴权使用全局 auth token
		{Name: "AGENT_SERVER_AUTH_TOKEN", Value: spec.ServerAuthToken},
		// Host 注册/心跳使用 Host 专属令牌
		{Name: "AGENT_CONTROLLER_AUTH_TOKEN", Value: spec.AuthToken},
	}

	container := corev1.Container{
		Name:            "agent",
		Image:           image,
		ImagePullPolicy: corev1.PullPolicy(k.kube.ImagePullPolicy),
		Ports: []corev1.ContainerPort{
			{ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
			{ContainerPort: 9090, Protocol: corev1.ProtocolTCP},
		},
		Env: envs,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "agent-data",
				MountPath: "/home/agent",
			},
		},
	}

	// 将 protocol.ResourceLimits 映射为 K8s ResourceList
	if spec.Resources.CPULimit != "" || spec.Resources.MemoryLimit != "" {
		limits := make(corev1.ResourceList)
		if spec.Resources.CPULimit != "" {
			// Docker 格式 "0.5" / "1" / "2" 需要加上单位后缀才能被 K8s 解析
			cpuVal := spec.Resources.CPULimit
			limits[corev1.ResourceCPU] = resource.MustParse(cpuVal)
		}
		if spec.Resources.MemoryLimit != "" {
			memVal := spec.Resources.MemoryLimit
			// Docker 风格 "512m" / "1g" → K8s 兼容 "512Mi" / "1Gi"
			memVal = strings.ReplaceAll(memVal, "m", "Mi")
			memVal = strings.ReplaceAll(memVal, "g", "Gi")
			limits[corev1.ResourceMemory] = resource.MustParse(memVal)
		}
		container.Resources.Limits = limits
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns,
			Labels: map[string]string{
				"app":             "maze-agent",
				"maze-agent-name": spec.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":             "maze-agent",
					"maze-agent-name": spec.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":             "maze-agent",
						"maze-agent-name": spec.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes: []corev1.Volume{
						k.buildAgentVolume(appName, spec.Name),
					},
				},
			},
		},
	}

	// 可选：ServiceAccount
	if k.kube.ServiceAccount != "" {
		deployment.Spec.Template.Spec.ServiceAccountName = k.kube.ServiceAccount
	}

	// 可选：ImagePullSecrets
	if k.kube.ImagePullSecret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: k.kube.ImagePullSecret},
		}
	}

	_, err := k.clientset.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		k.logger.Infof("deployment %s already exists, skip", appName)
		return nil
	}
	return err
}

// createService 创建 ClusterIP Service 暴露 Agent 端口
func (k *KubernetesRuntime) createService(ctx context.Context, ns, appName, hostName string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns,
			Labels: map[string]string{
				"app":             "maze-agent",
				"maze-agent-name": hostName,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app":             "maze-agent",
				"maze-agent-name": hostName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "grpc",
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := k.clientset.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		k.logger.Infof("service %s already exists, skip", appName)
		return nil
	}
	return err
}

// RemoveHost 按顺序删除 Service → Deployment → PVC，忽略 "not found"
// StopHost 停止运行时资源（Deployment + Service），保留持久化数据
func (k *KubernetesRuntime) StopHost(ctx context.Context, name string) error {
	if err := k.checkClient(); err != nil {
		return err
	}

	ns := k.kube.Namespace
	appName := fmt.Sprintf("maze-agent-%s", name)
	deletePolicy := metav1.DeletePropagationBackground
	deleteOpts := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}

	if err := k.clientset.CoreV1().Services(ns).Delete(ctx, appName, deleteOpts); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete service: %w", err)
		}
		k.logger.Infof("service %s not found, skip", appName)
	}

	if err := k.clientset.AppsV1().Deployments(ns).Delete(ctx, appName, deleteOpts); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete deployment: %w", err)
		}
		k.logger.Infof("deployment %s not found, skip", appName)
	}

	return nil
}

// RemoveHost 销毁 Host：停止运行时 + 清理持久化数据 + 清理镜像
func (k *KubernetesRuntime) RemoveHost(ctx context.Context, name string) error {
	if err := k.StopHost(ctx, name); err != nil {
		return err
	}

	if k.kube.VolumeType == "pvc" {
		pvcName := fmt.Sprintf("maze-agent-%s-data", name)
		deletePolicy := metav1.DeletePropagationBackground
		deleteOpts := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}
		if err := k.clientset.CoreV1().PersistentVolumeClaims(k.kube.Namespace).Delete(ctx, pvcName, deleteOpts); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("delete pvc: %w", err)
			}
			k.logger.Infof("pvc %s not found, skip", pvcName)
		}
	}

	if k.kube.VolumeType == "hostpath" && k.workspace.MountDir != "" {
		// hostPath 模式下 Manager 容器能直接看到宿主机目录；只有显式删除这里的目录，测试和下轮部署才不会吃到旧数据。
		agentDir := filepath.Join(k.workspace.MountDir, "agents", name)
		if err := os.RemoveAll(agentDir); err != nil {
			return fmt.Errorf("remove agent dir %s: %w", agentDir, err)
		}
		k.logger.Infof("removed agent dir %s", agentDir)
	}

	k.removeDockerImage(name)

	return nil
}

// InspectHost 通过 label selector 查询 Pod 状态
func (k *KubernetesRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	if err := k.checkClient(); err != nil {
		return nil, err
	}

	ns := k.kube.Namespace
	pods, err := k.clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("maze-agent-name=%s", name),
	})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pod found for host %s", name)
	}

	pod := pods.Items[0]
	status := mapPodPhase(pod.Status.Phase)

	return &protocol.ContainerInfo{
		ID:        string(pod.UID),
		Name:      pod.Name,
		Status:    status,
		Image:     pod.Spec.Containers[0].Image,
		CreatedAt: pod.CreationTimestamp.Time,
	}, nil
}

// buildAgentVolume 根据 VolumeType 构建 Agent 数据卷
// pvc 模式：引用 PVC；hostPath 模式：直接挂载宿主机目录
func (k *KubernetesRuntime) buildAgentVolume(appName, hostName string) corev1.Volume {
	if k.kube.VolumeType == "hostpath" {
		hostPath := filepath.Join(k.kube.HostPathBase, hostName)
		return corev1.Volume{
			Name: "agent-data",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
					Type: func() *corev1.HostPathType {
						t := corev1.HostPathDirectoryOrCreate
						return &t
					}(),
				},
			},
		}
	}
	return corev1.Volume{
		Name: "agent-data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: appName + "-data",
			},
		},
	}
}

// mapPodPhase 将 K8s Pod Phase 映射为统一的容器状态字符串
func mapPodPhase(phase corev1.PodPhase) string {
	switch phase {
	case corev1.PodRunning:
		return "running"
	case corev1.PodPending:
		return "created"
	case corev1.PodSucceeded, corev1.PodFailed:
		return "exited"
	default:
		return "unknown"
	}
}

// GetRuntimeLogs 通过 K8s Pod logs 获取运行日志
func (k *KubernetesRuntime) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	if err := k.checkClient(); err != nil {
		return "", err
	}

	ns := k.kube.Namespace
	pods, err := k.clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("maze-agent-name=%s", name),
	})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pod found for host %s", name)
	}

	tail := int64(tailLines)
	req := k.clientset.CoreV1().Pods(ns).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{
		TailLines: &tail,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("stream pod logs: %w", err)
	}
	defer stream.Close()

	data := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, readErr := stream.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}
	return string(data), nil
}

// IsHealthy 检查 K8s Deployment 是否存在且 ReadyReplicas > 0。
// Deployment 存在时（即使 ReadyReplicas == 0）也返回 true，由 K8s 自身处理 Pod 调度。
func (k *KubernetesRuntime) IsHealthy(ctx context.Context, name string) (bool, error) {
	if err := k.checkClient(); err != nil {
		return false, err
	}

	ns := k.kube.Namespace
	appName := fmt.Sprintf("maze-agent-%s", name)
	_, err := k.clientset.AppsV1().Deployments(ns).Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get deployment: %w", err)
	}
	return true, nil
}
