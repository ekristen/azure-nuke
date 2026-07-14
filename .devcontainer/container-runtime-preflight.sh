#!/usr/bin/env bash
set -euo pipefail

SOCKET_PATH="/var/run/docker.sock"
RUNTIME="unknown"

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

is_socket_ready() {
  if [[ ! -S "${SOCKET_PATH}" ]]; then
    return 1
  fi

  if have_cmd docker; then
    docker version >/dev/null 2>&1
    return $?
  fi

  if have_cmd podman; then
    podman version >/dev/null 2>&1
    return $?
  fi

  return 1
}

echo ""
echo "Container runtime preflight"
echo "--------------------------"

if have_cmd docker; then
  RUNTIME="docker"
elif have_cmd podman; then
  RUNTIME="podman"
fi

echo "Detected CLI: ${RUNTIME}"
echo "Expected socket: ${SOCKET_PATH}"

if is_socket_ready; then
  echo "Runtime socket check: OK"
  exit 0
fi

echo "Runtime socket check: FAILED"
echo ""
echo "No Docker-compatible API socket is currently available at ${SOCKET_PATH}."
echo "This devcontainer is configured to use Docker-in-Docker by default."
echo "If Docker remains unavailable after startup, use one of the host runtime options below."
echo ""
echo "If you are using Docker Desktop:"
echo "  - Start Docker Desktop and reopen the devcontainer"
echo ""
echo "If you are using Colima:"
echo "  - Start Colima: colima start"
echo "  - Ensure your Docker context points at Colima: docker context use colima"
echo ""
echo "If you are using Podman on macOS:"
echo "  - Start Podman machine: podman machine start"
echo "  - Enable/locate API socket: podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'"
echo "  - Provide a Docker-compatible socket at /var/run/docker.sock for VS Code"
echo "    (for example via a host symlink or Podman Docker API service)"
echo ""
echo "If you are using Windows + Docker Desktop:"
echo "  - Ensure Docker Desktop is running with WSL integration enabled"
echo "  - In Docker Desktop: Settings > Resources > WSL Integration"
echo "  - Reopen the devcontainer from VS Code after Docker is healthy"
echo ""
echo "If you are using WSL + Podman:"
echo "  - Start Podman system service in your WSL distro"
echo "  - Export DOCKER_HOST to the Podman API socket path"
echo "  - If you choose host-socket mounting, ensure /var/run/docker.sock maps"
echo "    to a Docker-compatible API socket"
echo ""
echo "See .devcontainer/README.md for detailed setup examples."
exit 0
