# Directives de Génération de Code pour le SDK Clever Cloud

Ce document décrit les règles et patterns utilisés pour générer le code du SDK à partir de la spécification OpenAPI de Clever Cloud.

## Architecture Générale

### Niveau 1: API Directe par x-service
- **Structure**: Un package par valeur de `x-service` dans `services/{x-service}/`
- **Fichiers**: Un fichier par `operationId` OpenAPI
- **Pattern de fonction**: `{OperationId}(ctx context.Context, client, tracer, ...params, ...queryOptions)`

### Niveau 2: Builder Pattern
- **Structure**: Interface fluide dans le package racine
- **Pattern**: `sdk.V4.{PathSegments}.{Operation}(ctx, ...options)`

## Règles de Génération

### 1. Organisation des Fichiers

```
services/
├── {x-service}/
│   ├── models.go         # Types spécifiques au service
│   ├── enums.go         # Constantes et enums
│   ├── options.go       # Options de query parameters
│   ├── {operationId}.go # Une fonction par opération
│   └── ...
models/
├── models.go            # Types communs
└── enums.go            # Constantes globales
```

### 2. Génération des Fonctions (Niveau 1)

**Template de fonction**:
```go
// {Description de l'opération depuis OpenAPI}
// Operation ID: {operationId}
// Path: {path}
// x-service: {x-service}
func {OperationId}(
    ctx context.Context,
    c client.Client,
    tracer trace.Tracer,
    {pathParams...},
    {bodyParam si POST/PUT/PATCH},
    opts ...{OperationId}Option,
) ({ReturnType}, error) {
    ctx, span := tracer.Start(ctx, "{operationId}")
    defer span.End()
    
    // Traitement des options
    options := &{OperationId}Options{}
    for _, opt := range opts {
        opt(options)
    }
    
    // Construction du path
    path := fmt.Sprintf("{path avec pathParams}")
    
    // Appel API
    response := client.{Method}[{ResponseType}](ctx, c, path, {body?})
    
    if response.HasError() {
        span.RecordError(response.Error())
        return {zeroValue}, response.Error()
    }
    
    result := response.Data()
    span.SetAttributes(
        // Attributs OpenTelemetry pertinents
    )
    
    return result, nil
}
```

### 3. Génération des Options (Query Parameters)

**Pour chaque opération avec query parameters**:
```go
type {OperationId}Options struct {
    {QueryParam} *{Type} `url:"{param_name},omitempty"`
    // ...
}

type {OperationId}Option func(*{OperationId}Options)

func With{QueryParam}({param} {Type}) {OperationId}Option {
    return func(o *{OperationId}Options) {
        o.{QueryParam} = &{param}
    }
}
```

### 4. Génération du Builder Pattern (Niveau 2)

**Analyse du path pour créer la hiérarchie**:
- `/v4/addon-providers/addon-pulsar/addons/{addonId}/persistent-topics`
- Devient: `V4.AddonProviders.AddonPulsar.Addons(addonId).PersistentTopics()`

**Rules pour les builders**:
1. Chaque segment statique devient une propriété
2. Chaque paramètre devient une méthode avec paramètre
3. Le dernier segment détermine les opérations disponibles

### 5. Génération des Types

**Models depuis les composants OpenAPI**:
```go
type {ComponentName} struct {
    {Field} {Type} `json:"{json_name}"`
    // Utiliser omitempty pour les champs optionnels
}
```

**Enums depuis les valeurs enum OpenAPI**:
```go
type {EnumName} {BaseType}

const (
    {EnumName}{Value1} {EnumName} = "{value1}"
    {EnumName}{Value2} {EnumName} = "{value2}"
    // ...
)

// {MethodName} returns the underlying {baseType} value
func (e {EnumName}) {MethodName}() {BaseType} {
    return {BaseType}(e)
}
```

**Note**: Chaque enum génère une méthode pour récupérer sa valeur native:
- `string` → méthode `String()` (implémente `fmt.Stringer`)
- `int` → méthode `Int()`
- `int32` → méthode `Int32()`
- `int64` → méthode `Int64()`
- `float32` → méthode `Float32()`
- `float64` → méthode `Float64()`
- `bool` → méthode `Bool()`

### 6. Gestion des x-service

**Mapping x-service → package**:
- `x-service: "service"` → `services/service/`
- `x-service: "product"` → `services/product/`
- Si pas de x-service → `services/default/`

### 7. Tracing OpenTelemetry

**Chaque fonction génère**:
- Span avec nom = operationId
- Attributs pour les paramètres importants
- Enregistrement des erreurs

### 8. Gestion des Erreurs Enum Struct

**Pour les enums complexes (struct avec enum)**:
```go
type {EnumName} struct {
    value string
}

func (e {EnumName}) String() string { return e.value }
func (e {EnumName}) MarshalJSON() ([]byte, error) { return json.Marshal(e.value) }
func (e *{EnumName}) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &e.value)
}

var (
    {EnumName}{Value1} = {EnumName}{"value1"}
    {EnumName}{Value2} = {EnumName}{"value2"}
)
```

## Process de Génération

1. **Parser la spec OpenAPI** avec `libopenapi`
2. **Extraire les opérations** groupées par x-service
3. **Générer les models** depuis les composants
4. **Créer les services** avec une fonction par operationId
5. **Construire les builders** depuis l'analyse des paths
6. **Générer les options** pour les query parameters

## Commandes de Génération

```bash
# Le document source est la spec OpenAPI unifiée d'AXO (dépôt `axo`, généré par
# `cargo run -p openapi-unified -- --root crates/api/api --offline`). Le Makefile la lit
# dans un checkout voisin (../../../rust/axo) ; OPENAPI=<fichier> pour en utiliser un autre.
cd gen
make build                                   # clean + models + services + builder
make build OPENAPI=/chemin/vers/unified-spec.json
make check                                   # les générateurs et le SDK compilent
make test-generators                         # tests de gen/internal (élagage, nommage)

# Mettre à jour les dépendances
cd .. && go mod tidy && go mod vendor
```

### Élagage du document (`gen/internal/spec`)

Le document unifié est fusionné à chaque régénération, jamais reconstruit : il peut porter
ce qu'un générateur ne sait pas traiter. `spec.Load` le normalise avant génération et
journalise chaque décision :

- les opérations dont `x-source` commence par `addon-api-` sont exclues : ce sont les
  routes que chaque addon provider sert sur son propre hôte (`/resources/{addonId}`,
  `/dashboard/…`), injoignables via api.clever-cloud.com et sans schéma ;
- une opération dont un `$ref` ne résout pas est abandonnée ;
- deux chemins qui ne diffèrent que par le nom d'un paramètre (`{addon_id}` / `{addonId}`)
  sont une seule route : on garde celle qui porte un `x-service`, sinon celle qui a des
  tags, sinon la première par ordre alphabétique ;
- un `operationId` déjà pris reçoit un suffixe dérivé du chemin (`…V2`), après application
  de `naming.OperationIDOverrides` ; le journal indique la clé à ajouter pour choisir ;
- une opération sans `x-service` en reçoit un depuis son premier tag, sinon depuis son
  chemin ;
- un paramètre de chemin nommé par le template mais non déclaré est déclaré (string).

### Nommage (`gen/internal/naming`)

Les noms de packages (`PackageName`, table `ServicePackageExceptions`), de fonctions
(`FunctionName`) et les exceptions par opération vivent dans un seul package utilisé par
les générateurs de services et de builder.

## Points d'Extension

- **Authentification**: Utiliser les options du client existant
- **Rate Limiting**: Intégrable dans le client wrapper
- **Retry Logic**: À implémenter dans les fonctions de base
- **Caching**: Possible via des decorators sur les fonctions