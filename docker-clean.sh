#!/bin/bash
# Docker Cleanup Script
# Safely removes all stopped containers, unused images, networks, and volumes.

echo "🛑 Stopping all running containers..."
docker stop $(docker ps -q) 2>/dev/null

echo "🧹 Removing all containers..."
docker rm -f $(docker ps -aq) 2>/dev/null

echo "🧱 Removing all images..."
docker rmi -f $(docker images -q) 2>/dev/null

echo "📦 Removing all volumes..."
docker volume rm $(docker volume ls -q) 2>/dev/null

echo "🌐 Removing all networks (except defaults)..."
docker network rm $(docker network ls -q | grep -vE 'bridge|host|none') 2>/dev/null

echo "🔥 Running system prune to catch leftovers..."
docker system prune -a --volumes -f

echo "✅ Docker cleanup complete!"
docker system df