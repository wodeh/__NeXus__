package nexus.admission

import future.keywords.if
import future.keywords.in

# Deny pods without resource limits
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.resources.limits.memory
    msg := sprintf("Container %s must have memory limits set", [container.name])
}

violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.resources.limits.cpu
    msg := sprintf("Container %s must have CPU limits set", [container.name])
}

# Deny privileged containers
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    container.securityContext.privileged == true
    msg := sprintf("Container %s must not run as privileged", [container.name])
}

# Deny containers running as root
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    container.securityContext.runAsUser == 0
    msg := sprintf("Container %s must not run as root (UID 0)", [container.name])
}

# Require read-only root filesystem
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.securityContext.readOnlyRootFilesystem
    msg := sprintf("Container %s must have readOnlyRootFilesystem set to true", [container.name])
}

# Require non-root user
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.securityContext.runAsNonRoot
    msg := sprintf("Container %s must have runAsNonRoot set to true", [container.name])
}

# Deny host network access
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    input.review.object.spec.hostNetwork == true
    msg := "Pod must not use hostNetwork"
}

# Deny host PID namespace
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    input.review.object.spec.hostPID == true
    msg := "Pod must not use hostPID"
}

# Deny host IPC namespace
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    input.review.object.spec.hostIPC == true
    msg := "Pod must not use hostIPC"
}

# Require PodDisruptionBudget for production deployments
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    input.review.object.metadata.namespace == "production"
    not input.review.object.metadata.annotations["nexus-platform.io/pdb-required"]
    msg := "Production deployments must have a PodDisruptionBudget"
}

# Require network policies
violation[{"msg": msg}] {
    input.review.object.kind == "Namespace"
    not data.kubernetes.networkpolicies[input.review.object.metadata.name]
    msg := sprintf("Namespace %s must have a default deny NetworkPolicy", [input.review.object.metadata.name])
}

# Tenant isolation: ensure pods have tenant labels
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    not input.review.object.metadata.labels["nexus-platform.io/tenant-id"]
    msg := "Pod must have nexus-platform.io/tenant-id label for tenant isolation"
}

# Deny images from untrusted registries
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not startswith(container.image, "ghcr.io/nexus-platform/")
    not startswith(container.image, "registry.nexus-platform.internal/")
    not startswith(container.image, "public.ecr.aws/")
    msg := sprintf("Container %s uses image from untrusted registry: %s", [container.name, container.image])
}

# Require resource requests (not just limits)
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.resources.requests.memory
    msg := sprintf("Container %s must have memory requests set", [container.name])
}

violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.resources.requests.cpu
    msg := sprintf("Container %s must have CPU requests set", [container.name])
}

# GPU workloads must have tolerations
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    container.resources.limits["nvidia.com/gpu"]
    not pod_has_gpu_toleration(input.review.object)
    msg := sprintf("Container %s requesting GPU must have nvidia.com/gpu toleration", [container.name])
}

pod_has_gpu_toleration(pod) if {
    some toleration in pod.spec.tolerations
    toleration.key == "nvidia.com/gpu"
}

# Require seccomp profile
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    not input.review.object.metadata.annotations["seccomp.security.alpha.kubernetes.io/pod"]
    not input.review.object.spec.securityContext.seccompProfile
    msg := "Pod must have seccomp profile configured"
}

# Deny latest tag
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    endswith(container.image, ":latest")
    msg := sprintf("Container %s must not use 'latest' tag", [container.name])
}

# Require liveness and readiness probes
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.livenessProbe
    msg := sprintf("Container %s must have a livenessProbe", [container.name])
}

violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.readinessProbe
    msg := sprintf("Container %s must have a readinessProbe", [container.name])
}

# Require service account (not default)
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    input.review.object.spec.serviceAccountName == "default"
    msg := "Pod must not use the default service account"
}

# Deny capabilities beyond NET_BIND_SERVICE
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    some cap in container.securityContext.capabilities.add
    cap != "NET_BIND_SERVICE"
    msg := sprintf("Container %s has disallowed capability: %s", [container.name, cap])
}

# Require drop ALL capabilities
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not container.securityContext.capabilities.drop
    msg := sprintf("Container %s must drop ALL capabilities", [container.name])
}

violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    container := input.review.object.spec.containers[_]
    not "ALL" in container.securityContext.capabilities.drop
    msg := sprintf("Container %s must drop ALL capabilities", [container.name])
}

# Storage: require encryption at rest
violation[{"msg": msg}] {
    input.review.object.kind == "PersistentVolumeClaim"
    not input.review.object.spec.storageClassName
    msg := "PersistentVolumeClaim must specify a storage class"
}

# Ingress: require TLS
violation[{"msg": msg}] {
    input.review.object.kind == "Ingress"
    some rule in input.review.object.spec.rules
    not input.review.object.spec.tls
    msg := "Ingress must have TLS configured"
}

# Ingress: require specific annotations
violation[{"msg": msg}] {
    input.review.object.kind == "Ingress"
    not input.review.object.metadata.annotations["cert-manager.io/cluster-issuer"]
    msg := "Ingress must have cert-manager annotation for automatic TLS"
}

# ConfigMap: deny sensitive data in plain text
violation[{"msg": msg}] {
    input.review.object.kind == "ConfigMap"
    some key, value in input.review.object.data
    contains(lower(key), "password")
    msg := sprintf("ConfigMap contains potential password in key: %s", [key])
}

violation[{"msg": msg}] {
    input.review.object.kind == "ConfigMap"
    some key, value in input.review.object.data
    contains(lower(key), "secret")
    msg := sprintf("ConfigMap contains potential secret in key: %s", [key])
}

violation[{"msg": msg}] {
    input.review.object.kind == "ConfigMap"
    some key, value in input.review.object.data
    contains(lower(key), "token")
    msg := sprintf("ConfigMap contains potential token in key: %s", [key])
}

# Job/CronJob: require TTL
violation[{"msg": msg}] {
    input.review.object.kind == "Job"
    not input.review.object.spec.ttlSecondsAfterFinished
    msg := "Job must have ttlSecondsAfterFinished set"
}

violation[{"msg": msg}] {
    input.review.object.kind == "CronJob"
    not input.review.object.spec.jobTemplate.spec.ttlSecondsAfterFinished
    msg := "CronJob must have ttlSecondsAfterFinished set"
}

# HorizontalPodAutoscaler: require min replicas >= 2 for production
violation[{"msg": msg}] {
    input.review.object.kind == "HorizontalPodAutoscaler"
    input.review.object.metadata.namespace == "production"
    input.review.object.spec.minReplicas < 2
    msg := "Production HPA must have minReplicas >= 2"
}

# Namespace: require resource quotas
violation[{"msg": msg}] {
    input.review.object.kind == "Namespace"
    not data.kubernetes.resourcequotas[input.review.object.metadata.name]
    msg := sprintf("Namespace %s must have a ResourceQuota", [input.review.object.metadata.name])
}

# Namespace: require limit ranges
violation[{"msg": msg}] {
    input.review.object.kind == "Namespace"
    not data.kubernetes.limitranges[input.review.object.metadata.name]
    msg := sprintf("Namespace %s must have a LimitRange", [input.review.object.metadata.name])
}

# Service: require specific type restrictions
violation[{"msg": msg}] {
    input.review.object.kind == "Service"
    input.review.object.spec.type == "LoadBalancer"
    not input.review.object.metadata.annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]
    msg := "Internal LoadBalancer must be annotated as internal"
}

# Pod: require topology spread constraints for multi-AZ
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    count(input.review.object.spec.template.spec.topologySpreadConstraints) == 0
    msg := "Deployment must have topologySpreadConstraints for zone distribution"
}

# Deny hostPath volumes
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    some volume in input.review.object.spec.volumes
    volume.hostPath
    msg := sprintf("Pod must not use hostPath volumes: %s", [volume.name])
}

# Require PodSecurityContext
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    not input.review.object.spec.securityContext
    msg := "Pod must have a securityContext defined"
}

# Ensure fsGroup is set for volumes
violation[{"msg": msg}] {
    input.review.object.kind == "Pod"
    count(input.review.object.spec.volumes) > 0
    not input.review.object.spec.securityContext.fsGroup
    msg := "Pod with volumes must have securityContext.fsGroup set"
}

# IPTV-specific: require dedicated node affinity
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    input.review.object.metadata.labels["app.kubernetes.io/component"] == "iptv-transcode"
    not input.review.object.spec.template.spec.affinity.nodeAffinity
    msg := "IPTV transcoding must have nodeAffinity for GPU nodes"
}

# AI-specific: require GPU node selector
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    input.review.object.metadata.labels["app.kubernetes.io/component"] == "ai-inference"
    not input.review.object.spec.template.spec.nodeSelector["nvidia.com/gpu.present"]
    msg := "AI inference must have nvidia.com/gpu.present node selector"
}

# PMS-specific: require encrypted connections
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    input.review.object.metadata.labels["app.kubernetes.io/component"] == "pms-core"
    container := input.review.object.spec.template.spec.containers[_]
    env := container.env[_]
    env.name == "DATABASE_SSL_MODE"
    env.value != "require"
    msg := "PMS core must use SSL for database connections"
}

# IoT-specific: require MQTT TLS
violation[{"msg": msg}] {
    input.review.object.kind == "Deployment"
    input.review.object.metadata.labels["app.kubernetes.io/component"] == "iot-gateway"
    container := input.review.object.spec.template.spec.containers[_]
    env := container.env[_]
    env.name == "MQTT_TLS_ENABLED"
    env.value != "true"
    msg := "IoT gateway must have MQTT TLS enabled"
}

# Allow list for allowed container registries
allowed_registries := [
    "ghcr.io/nexus-platform",
    "registry.nexus-platform.internal",
    "public.ecr.aws",
    "gcr.io",
    "registry.k8s.io"
]

# Helper functions
startswith(s, prefix) if {
    count(prefix) <= count(s)
    prefix == substring(s, 0, count(prefix))
}
