ROOT_MK_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
PROJECT_ROOT := $(abspath $(ROOT_MK_DIR)/..)

PLATFORM ?= kubernetes
ENV ?= dev
TEST_NAME ?=

MODULES := \
	fabrication/cradle \
	the-mesa/director-core \
	sweetwater/black-ridge/server

COVERAGE_MODULES := \
	fabrication/cradle \
	fabrication/tests/integration \
	the-mesa/director-core \
	sweetwater/black-ridge/server

FRONTEND_MODULES := \
	fabrication/skin \
	the-mesa/arrival-gate \
	the-mesa/director-console \
	sweetwater/black-ridge/web

CRADLE_API_DIR := $(PROJECT_ROOT)/fabrication/cradle/api
SKIN_DIR := $(PROJECT_ROOT)/fabrication/skin
TEST_DIR := $(PROJECT_ROOT)/fabrication/tests/integration
MOLDS_DIR := $(PROJECT_ROOT)/fabrication/molds

DIRECTOR_CORE_DOCKERFILE := $(PROJECT_ROOT)/the-mesa/director-core/Dockerfile
WEB_DOCKERFILE := $(PROJECT_ROOT)/the-mesa/Dockerfile.web
AGENT_DOCKERFILE := $(PROJECT_ROOT)/sweetwater/black-ridge/Dockerfile

DIRECTOR_CORE_IMAGE := maze-director-core:latest
WEB_IMAGE := maze-web:latest
AGENT_IMAGE := maze-agent:latest

COMPOSE_FILE := $(PROJECT_ROOT)/the-mesa/docker-compose.yml
COMPOSE_TEST_FILE := $(PROJECT_ROOT)/fabrication/tests/integration/docker-compose.test.yml

ifeq ($(ENV),dev)
  K8S_NAMESPACE := maze-dev
  K8S_OVERLAY := overlays/dev
  COMPOSE_PROJECT := maze-dev
  HOST_DATA_DIR := $(HOME)/.maze-dev
  PORT_DIRECTOR_CORE := 7090
  PORT_WEB := 7080
  PORT_POSTGRES := 5432
  POSTGRES_HOSTPATH := /tmp/maze-dev/postgresql/data
  AGENT_HOSTPATH_BASE := /tmp/maze-dev/kubernetes/agents
else ifeq ($(ENV),test)
  K8S_NAMESPACE := maze-test
  K8S_OVERLAY := overlays/test
  COMPOSE_PROJECT := maze-test
  HOST_DATA_DIR := $(HOME)/.maze-test
  PORT_DIRECTOR_CORE := 9090
  PORT_WEB := 9080
  PORT_POSTGRES := 5433
  POSTGRES_HOSTPATH := /tmp/maze-test/postgresql/data
  AGENT_HOSTPATH_BASE := /tmp/maze-test/kubernetes/agents
else ifeq ($(ENV),prod)
  K8S_NAMESPACE := maze-prod
  K8S_OVERLAY := overlays/production
  COMPOSE_PROJECT := maze-prod
  HOST_DATA_DIR := $(HOME)/.maze-prod
  PORT_DIRECTOR_CORE := 8090
  PORT_WEB := 10800
  PORT_POSTGRES := 5434
  POSTGRES_HOSTPATH :=
  AGENT_HOSTPATH_BASE :=
else
  $(error Unsupported ENV '$(ENV)'; expected dev, test, or prod)
endif

K8S_OVERLAY_DIR := $(PROJECT_ROOT)/fabrication/kubernetes/$(K8S_OVERLAY)

# 这里必须延迟展开，才能让 test-integration 这类 target-specific override
# 正确改写 COMPOSE_PROJECT，而不是意外落回默认的 dev 项目名。
DOCKER_COMPOSE = docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT)
DOCKER_TEST_COMPOSE = docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT)

# Docker Compose 通过环境变量插值解析端口和数据目录；统一在这里导出，避免根和子目录重复维护。
export PORT_WEB PORT_DIRECTOR_CORE PORT_POSTGRES HOST_DATA_DIR COMPOSE_PROJECT
