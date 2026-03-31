# Spec, status et commentaires Kubebuilder

Ce document résume le **rôle** de `spec` et `status` sur une ressource Kubernetes personnalisée (CRD), et les **commentaires (markers)** que Kubebuilder/controller-gen lit pour générer le schéma OpenAPI v3 et le sous-ressource status. Pour un cas concret (conditions + champs de status optionnels), voir aussi [`kafka-kafkatopic-status-conditions.md`](kafka-kafkatopic-status-conditions.md).

---

## Spec vs status : convention Kubernetes

| Partie | Rôle | Qui l’écrit typiquement |
|--------|------|-------------------------|
| **`metadata`** | Identité, labels, annotations, namespace, etc. | Utilisateur / contrôleurs (système) |
| **`spec`** | **État désiré** : ce que l’utilisateur (ou un contrôleur parent) demande | Principalement le client qui crée ou met à jour la ressource |
| **`status`** | **État observé** : ce que l’opérateur a constaté ou réussi à faire | Le **contrôleur** du Kind, via la sous-API `.../status` |

Règle d’architecture : seuls les composants qui “possèdent” l’observation (souvent ton contrôleur) doivent mettre à jour `status`. Les utilisateurs ne devraient pas dépendre d’écrire eux-mêmes des champs sensibles dans `status` pour piloter le comportement (ça reste du signal “observé”, pas de la configuration).

Référence : [API Conventions — Typical Status Properties](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).

---

## Structure typique d’un `Kind` en Go

Un type racine ressemble à ceci (ordre et champs habituels) :

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type MaRessource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`
	Spec   MaRessourceSpec   `json:"spec"`
	Status MaRessourceStatus `json:"status,omitzero"`
}
```

- **`// +kubebuilder:object:root=true`** — Ce type est une ressource racine (génération du CRD, enregistrement du schéma).
- **`// +kubebuilder:subresource:status`** — Active la sous-ressource **status** : mises à jour via `kubectl patch ... --subresource=status` et dans controller-runtime avec `r.Status().Update(...)` / `r.Status().Patch(...)`.

Les structures `...Spec` et `...Status` sont des types distincts : ça sépare clairement désir vs observé dans le code et dans la CRD.

---

## Tags JSON obligatoires

Sur **chaque** champ sérialisé, Kubebuilder s’attend à un **`json:"..."`** cohérent avec le nom du champ dans le manifest YAML (souvent `camelCase`).

Sans tag JSON correct, la génération ou la sérialisation peuvent être incorrectes.

---

## Marker `// +optional`

Indique un champ **non requis** dans le schéma OpenAPI produit pour la CRD.

- Sur le **spec** : champs que l’utilisateur peut omettre.
- Sur le **status** : presque tous les champs devraient être optionnels, sinon un simple `Status().Update` qui ne remplit qu’une partie du status (ex. seulement `conditions`) peut être **rejeté** par le serveur API (erreur de validation “Required value”). En pratique on combine souvent :
  - `// +optional`
  - et des tags **`json:"...,omitempty"`** pour les types où la valeur zéro ne doit pas apparaître ou où tu veux éviter de sérialiser des listes / chaînes vides inutilement.

Exemple (pattern recommandé pour des champs de status “libres”) :

```go
// +optional
Ready bool `json:"ready,omitempty"`
```

---

## Validation sur les champs (`common marker`)

Ces commentaires alimentent le schéma CRD (validation côté API server). Exemples fréquents :

| Marker | Effet |
|--------|--------|
| `// +kubebuilder:validation:Required` | Champ requis (souvent sur un champ **spec**). |
| `// +kubebuilder:validation:MinLength=1` | Longueur minimale pour une chaîne. |
| `// +kubebuilder:validation:MaxLength=100` | Longueur max. |
| `// +kubebuilder:validation:Minimum=1` | Borne inférieure (nombre). |
| `// +kubebuilder:validation:Pattern="^..."`  | Expression régulière. |
| `// +kubebuilder:default="valeur"` | Valeur par défaut dans le schéma (comportement selon version des outils ; à tester). |

Liste complète et détails : [Kubebuilder — Markers / CRD validation](https://book.kubebuilder.io/reference/markers/crd-validation.html).

---

## Marker sur le champ `Spec` au niveau racine

Sur le type racine, on peut forcer explicitement que `spec` est obligatoire dans le manifest :

```go
// +kubebuilder:validation:Required
Spec MaRessourceSpec `json:"spec"`
```

Le bloc **`status`** est en général marqué **`// +optional`** sur le struct racine, car une ressource nouvellement créée peut ne pas encore avoir de status rempli.

---

## Listes avec sémantique Kubernetes : `listType`, `listMapKey`

Pour les **conditions** (`[]metav1.Condition`) ou toute liste où chaque élément a une clé unique (souvent `type`), utilise :

```go
// +listType=map
// +listMapKey=type
// +optional
Conditions []metav1.Condition `json:"conditions,omitempty"`
```

Cela génère dans la CRD les extensions `x-kubernetes-list-type` et `x-kubernetes-list-map-keys`, ce qui permet des **strategic merge** / mises à jour cohérentes côté API.

Référence : [Markers — CRD processing](https://book.kubebuilder.io/reference/markers/crd-processing.html).

---

## Conditions (`metav1.Condition`)

La convention actuelle est d’exposer l’état synthétique via **`status.conditions`** (types `True` / `False` / `Unknown`, `reason`, `message`, `lastTransitionTime`). Les helpers `k8s.io/apimachinery/pkg/api/meta` (`SetStatusCondition`, `FindStatusCondition`, etc.) travaillent sur ce slice.

Le **spec** ne contient en principe **pas** les conditions : elles décrivent l’**observation** du contrôleur.

---

## Ressource : scope, noms, colonnes d’impression

Non spécifiques à spec/status, mais souvent sur le même fichier :

```go
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
```

---

## Après modification des types ou des markers

Depuis la racine du dépôt :

```bash
make generate  # DeepCopy, etc.
make manifests # CRD + RBAC à partir des markers
```

Ne pas éditer à la main `zz_generated.*.go` ni `config/crd/bases/*.yaml` sauf exception documentée.

---

## Pièges fréquents (résumé)

1. **Contrôleur et type désynchronisés** — Ex. utiliser `status.conditions` en Go alors que le struct `Status` n’a pas de champ `Conditions` → erreur de compilation.
2. **Trop de champs “required” dans le status** — Oublier `+optional` / `omitempty` sur le status alors que le reconcile ne met à jour qu’un sous-ensemble → **422** sur `Status().Update`.
3. **Spec et status inversés** — Mettre des informations purement observées (dernière erreur, “ready”) dans le spec : à éviter ; garder le spec pour l’intention.
4. **Oublier `+kubebuilder:subresource:status`** — Sans lui, pas de sous-ressource status dédiée ; les contrôleurs ne peuvent pas suivre le modèle habituel `r.Status().Update`.

---

## Liens utiles

- [Kubebuilder book — Generating CRDs](https://book.kubebuilder.io/reference/generating-crd.html)
- [Markers reference](https://book.kubebuilder.io/reference/markers.html)
- [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
