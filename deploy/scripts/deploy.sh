#!/bin/bash
# ============================================
# Webook K8s 部署脚本
# ============================================

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Webook K8s 部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 配置变量（请根据实际情况修改）
REGISTRY="your-registry"  # 镜像仓库地址
IMAGE_NAME="webook"
IMAGE_TAG="latest"

# Step 1: 构建 Docker 镜像
echo -e "\n${YELLOW}[Step 1/4] 构建 Docker 镜像...${NC}"
docker build -t ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} -f deploy/docker/Dockerfile .

# Step 2: 推送镜像到仓库
echo -e "\n${YELLOW}[Step 2/4] 推送镜像到仓库...${NC}"
docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

# Step 3: 创建命名空间
echo -e "\n${YELLOW}[Step 3/4] 部署 K8s 资源...${NC}"
kubectl apply -f deploy/k8s/namespace.yaml

# Step 4: 按顺序部署资源
echo -e "  - 部署 ConfigMap 和 Secret..."
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml

echo -e "  - 部署 MySQL..."
kubectl apply -f deploy/k8s/mysql.yaml

echo -e "  - 等待 MySQL 就绪..."
kubectl wait --for=condition=ready pod -l app=mysql -n webook --timeout=120s

echo -e "  - 部署 Redis..."
kubectl apply -f deploy/k8s/redis.yaml

echo -e "  - 等待 Redis 就绪..."
kubectl wait --for=condition=ready pod -l app=redis -n webook --timeout=60s

echo -e "  - 部署 Webook 应用..."
kubectl apply -f deploy/k8s/webook.yaml

# Step 5: 等待部署完成
echo -e "\n${YELLOW}[Step 4/4] 等待所有 Pod 就绪...${NC}"
kubectl wait --for=condition=ready pod -l app=webook -n webook --timeout=120s

# 显示部署状态
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${YELLOW}Pod 状态:${NC}"
kubectl get pods -n webook

echo -e "\n${YELLOW}Service 状态:${NC}"
kubectl get svc -n webook

echo -e "\n${YELLOW}访问方式:${NC}"
echo -e "kubectl port-forward svc/webook-service 8080:80 -n webook"
echo -e "然后访问: http://localhost:8080"
