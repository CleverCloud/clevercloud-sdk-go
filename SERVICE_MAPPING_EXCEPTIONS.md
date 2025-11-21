# Service Mapping Exceptions

Ce document décrit les exceptions au mapping `x-service` pour gérer les incohérences de classification dans la spécification OpenAPI.

## Problème Identifié

La spécification OpenAPI a des incohérences dans la classification `x-service` des opérations Pulsar :
- Certaines opérations Pulsar ont `x-service: product`
- D'autres opérations Pulsar ont `x-service: service`
- Toutes devraient logiquement être dans `x-service: pulsar`

## Mapping d'Exception

### Operations Pulsar mal classées (x-service: product → pulsar)

| OperationId | Path | Current x-service | Target Service | Reason |
|-------------|------|-------------------|----------------|--------|
| `getPulsar` | `/v4/addon-providers/addon-pulsar/addons/{addonId}` | product | pulsar | Spécifique à Pulsar addon |
| `deletePulsar` | `/v4/addon-providers/addon-pulsar/addons/{addonId}` | product | pulsar | Spécifique à Pulsar addon |
| `getPulsarPolicies` | `/v4/addon-providers/addon-pulsar/addons/{addonId}/storage-policies` | product | pulsar | Gestion des politiques Pulsar |
| `setStoragePolicies` | `/v4/addon-providers/addon-pulsar/addons/{addonId}/storage-policies` | product | pulsar | Gestion des politiques Pulsar |
| `renewPulsarToken` | `/v4/addon-providers/addon-pulsar/addons/{addonId}/renew-biscuit` | product | pulsar | Token Pulsar spécifique |
| `createPulsarTenantAndNamespace` | `/v4/addon-providers/addon-pulsar/addons/{addonId}/create-tenant-and-namespace` | product | pulsar | Gestion namespace Pulsar |
| `deletePulsarTenantAndNamespace` | `/v4/addon-providers/addon-pulsar/addons/{addonId}/delete-tenant-and-namespace` | product | pulsar | Gestion namespace Pulsar |
| `createPulsar` | `/v4/addon-providers/addon-pulsar/addons` | product | pulsar | Création addon Pulsar |
| `getPulsarProviderInfo` | `/v4/addon-providers/addon-pulsar` | product | pulsar | Info provider Pulsar |
| `getPulsarCluster` | `/v4/addon-providers/addon-pulsar/clusters/{clusterId}` | product | pulsar | Cluster Pulsar |
| `listPulsarConsumptions` | `/v4/pulsar/organisations/{ownerId}/pulsar/consumptions` | product | pulsar | Consommation Pulsar |

### Operations Functions/Pulsar (x-service: product → pulsar)

| OperationId | Path | Current x-service | Target Service | Reason |
|-------------|------|-------------------|----------------|--------|
| `getTriggerPulsar` | `/v4/functions/organisations/{ownerId}/functions/{functionId}/trigger-pulsar` | product | pulsar | Trigger Pulsar |
| `createTriggerPulsar` | `/v4/functions/organisations/{ownerId}/functions/{functionId}/trigger-pulsar` | product | pulsar | Trigger Pulsar |
| `deleteTriggerPulsar` | `/v4/functions/organisations/{ownerId}/functions/{functionId}/trigger-pulsar` | product | pulsar | Trigger Pulsar |

### Operations V2 Pulsar (x-service: product → pulsar)

| OperationId | Path | Current x-service | Target Service | Reason |
|-------------|------|-------------------|----------------|--------|
| `provisionPulsar` (POST) | `/v2/providers/addon-pulsar/resources` | product | pulsar | Provision Pulsar V2 |
| `getPulsarV2` | `/v2/providers/addon-pulsar/resources/{addonId}` | product | pulsar | Get Pulsar V2 |
| `provisionPulsar` (DELETE) | `/v2/providers/addon-pulsar/resources/{addonId}` | product | pulsar | Delete Pulsar V2 |

## Règles de Mapping d'Exception

1. **Toute opération avec "pulsar" dans le path** → service `pulsar`
2. **Toute opération avec "Pulsar" dans l'operationId** → service `pulsar`
3. **Operations avec tag "pulsar"** → service `pulsar`

## Implémentation

Les opérations listées ci-dessus doivent être générées dans `services/pulsar/` au lieu de `services/product/` pour maintenir la cohérence logique du SDK.