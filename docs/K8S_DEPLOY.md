# Webook K8s 部署指南

本文档介绍如何使用 Kubernetes 部署 Webook 应用（3 副本）。

## 环境要求

- Docker Desktop（已启用 Kubernetes）
- Go 1.25+
- kubectl

## 快速开始

### 1. 构建镜像

```powershell
# 编译 Linux 二进制文件
$env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
go build -o webook-linux .

# 构建 Docker 镜像
docker build -t webook:latest -f deploy/docker/Dockerfile .
```

### 2. 部署到 K8s

```powershell
# 按顺序执行
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/mysql.yaml
kubectl apply -f deploy/k8s/redis.yaml
kubectl apply -f deploy/k8s/webook.yaml
```

### 3. 验证部署

```powershell
# 查看 Pod 状态（应该有 5 个 Pod：1 MySQL + 1 Redis + 3 Webook）
kubectl get pods -n webook

# 查看 Service
kubectl get svc -n webook
```

### 4. 访问服务

```powershell
# Docker Desktop 自动映射 LoadBalancer 到 localhost
curl http://localhost/users/login
```

## 常用命令

| 操作 | 命令 |
|------|------|
| 查看 Pod | `kubectl get pods -n webook` |
| 查看日志 | `kubectl logs -f <pod-name> -n webook` |
| 进入容器 | `kubectl exec -it <pod-name> -n webook -- sh` |
| 扩容 | `kubectl scale deployment webook --replicas=5 -n webook` |
| 缩容 | `kubectl scale deployment webook --replicas=1 -n webook` |
| 重启 | `kubectl rollout restart deployment webook -n webook` |

## 停止服务

```powershell
# 方式1：删除整个命名空间（推荐）
kubectl delete namespace webook

# 方式2：仅停止 Pod
kubectl scale deployment webook mysql redis --replicas=0 -n webook
```

## 目录结构

```
deploy/k8s/
├── namespace.yaml   # 命名空间
├── configmap.yaml   # 环境变量配置
├── secret.yaml      # 敏感信息（JWT密钥等）
├── mysql.yaml       # MySQL 部署 + 持久化存储
├── redis.yaml       # Redis 部署
└── webook.yaml      # Webook 应用（3 副本）
```

## 配置说明

### ConfigMap（非敏感配置）
- `DB_DSN`: MySQL 连接地址
- `REDIS_ADDR`: Redis 连接地址
- `CORS_ORIGIN`: CORS 允许的域名

### Secret（敏感配置）
- `JWT_SECRET`: JWT 签名密钥
- `SESSION_SECRET`: Session 密钥
- `MYSQL_ROOT_PASSWORD`: MySQL 密码

> ⚠️ 生产环境请更换 `secret.yaml` 中的默认密钥！
