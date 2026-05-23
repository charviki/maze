package runtime

// buildSemaphore 限制同时执行的 docker build 并发数，防止镜像缓存批量失效时
// 引发重建风暴导致宿主机 CPU/IO 高负载。
// 缓存命中和容器启动不受此限制。
var buildSemaphore = make(chan struct{}, 2)
