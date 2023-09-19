# istio-registry-sync

![Version: 0.1.1](https://img.shields.io/badge/Version-0.1.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.3.0](https://img.shields.io/badge/AppVersion-v0.3.0-informational?style=flat-square)

A Helm chart for deploying Istio Registry Sync

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| cloudmap.region | string | `"us-east-1"` | AWS Region to use to connect to Cloud Map |
| consul.endpoint | string | `""` | Consul's endpoint to query service catalog |
| consul.namespace | string | `""` | Consul's namespace to search service catalog |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"docker.cloudsmith.io/tetrate/tis-containers/istio-registry-sync"` |  |
| image.tag | string | `"v0.4.0"` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| publishNamespace | string | `""` | Istio Registry Sync publishes ServiceEntry into this namespace or the namespace it is deployed into |
| resources | object | `{}` |  |
| resyncPeriod | int | `5` | Interval in seconds between syncing |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` | If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |

