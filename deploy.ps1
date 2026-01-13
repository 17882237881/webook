# deploy.ps1 - Windows PowerShell 部署脚本
# ============================================

Write-Host "========================================" -ForegroundColor Green
Write-Host "  Webook K8s 本地部署脚本" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# Step 1: 检查 Docker 是否运行
Write-Host "`n[Step 1/5] 检查 Docker 状态..." -ForegroundColor Yellow
docker info | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: Docker 未运行，请先启动 Docker Desktop" -ForegroundColor Red
    exit 1
}
Write-Host "Docker 运行正常 ✓" -ForegroundColor Green

# Step 2: 检查 kubectl 是否可用
Write-Host "`n[Step 2/5] 检查 kubectl 状态..." -ForegroundColor Yellow
kubectl cluster-info | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: kubectl 无法连接到集群，请确保 K8s 已启用" -ForegroundColor Red
    Write-Host "  - 如果使用 Docker Desktop: Settings -> Kubernetes -> Enable Kubernetes" -ForegroundColor Yellow
    Write-Host "  - 如果使用 Minikube: minikube start" -ForegroundColor Yellow
    exit 1
}
Write-Host "kubectl 连接正常 ✓" -ForegroundColor Green

# Step 3: 构建 Docker 镜像
Write-Host "`n[Step 3/5] 构建 Docker 镜像..." -ForegroundColor Yellow
docker build -t webook:latest .
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: Docker 镜像构建失败" -ForegroundColor Red
    exit 1
}
Write-Host "镜像构建成功 ✓" -ForegroundColor Green

# Step 4: 部署 K8s 资源
Write-Host "`n[Step 4/5] 部署 K8s 资源..." -ForegroundColor Yellow

Write-Host "  - 创建命名空间..." -ForegroundColor Cyan
kubectl apply -f k8s/namespace.yaml

Write-Host "  - 部署 ConfigMap 和 Secret..." -ForegroundColor Cyan
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

Write-Host "  - 部署 MySQL..." -ForegroundColor Cyan
kubectl apply -f k8s/mysql.yaml

Write-Host "  - 等待 MySQL 就绪 (最多 120 秒)..." -ForegroundColor Cyan
kubectl wait --for=condition=ready pod -l app=mysql -n webook --timeout=120s

Write-Host "  - 部署 Redis..." -ForegroundColor Cyan
kubectl apply -f k8s/redis.yaml

Write-Host "  - 等待 Redis 就绪..." -ForegroundColor Cyan
kubectl wait --for=condition=ready pod -l app=redis -n webook --timeout=60s

Write-Host "  - 部署 Webook 应用 (3 副本)..." -ForegroundColor Cyan
kubectl apply -f k8s/webook.yaml

# Step 5: 等待部署完成
Write-Host "`n[Step 5/5] 等待 Webook Pod 就绪..." -ForegroundColor Yellow
kubectl wait --for=condition=ready pod -l app=webook -n webook --timeout=120s

# 显示部署状态
Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  部署完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`n[Pod 状态]" -ForegroundColor Yellow
kubectl get pods -n webook

Write-Host "`n[Service 状态]" -ForegroundColor Yellow
kubectl get svc -n webook

Write-Host "`n[访问方式]" -ForegroundColor Yellow
Write-Host "  1. 运行端口转发:" -ForegroundColor Cyan
Write-Host "     kubectl port-forward svc/webook-service 8080:80 -n webook" -ForegroundColor White
Write-Host "  2. 然后访问: http://localhost:8080" -ForegroundColor White
