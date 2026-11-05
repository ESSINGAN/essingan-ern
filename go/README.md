# essingan-ern

La bibliothèque Go des **ERN** (ESSINGAN Resource Name), les noms uniques des ressources d'Essingan, sur le modèle des ARN d'AWS.

- **Règles :** ADR-0051 et Doc 73-01 (repo `sre-devops-assistant`), contrat `hub-api.openapi.yaml`.
- **Pour qui :** management-api, essingan-iam, cno et les autres services en Go. La version TypeScript vit dans `essingan-client-core` (Eric).
- **Ce qu'elle n'est pas :** ni un service, ni une base de données. Elle est intégrée au code de chaque programme et ne fait aucun appel réseau.

## Ce qu'elle fait

| Fonction | Exemple | Rôle |
|---|---|---|
| Fabriquer | `ern.NewProject(compte, org, projet)` | Crée le nom d'une ressource, ou renvoie une erreur |
| Lire | `ern.Parse(texte)` | Découpe un nom en morceaux |
| Valider | `e.Validate()`, `ern.IsERI(id)` | Refuse un nom mal formé |
| Vérifier le fournisseur | `e.CheckProvider(registre)` | Refuse un fournisseur cloud absent du registre |
| Comparer | `ern.Match(règle, nom)` | Dit si un nom correspond à une règle de droits |

```go
projet, err := ern.NewProject("482019375521", "o-01jd2v8k4m7q9x3a", "p-01jf3w9n5r8t2y6b")
// ern:essingan:account::482019375521:o-01jd2v8k4m7q9x3a/project/p-01jf3w9n5r8t2y6b

ok, err := ern.Match("ern:essingan:account::482019375521:o-01jd2v8k4m7q9x3a/project/*", projet.String())
// ok == true
```

Si `Match` renvoie une erreur, la règle ou le nom est invalide : l'appelant doit refuser l'accès.

## Les fournisseurs cloud

La liste n'est pas dans le code : elle vit dans le registre du catalogue (ADR-0051 D7). Ajouter un fournisseur ne demande donc aucune nouvelle version de la bibliothèque.

- `Parse` et `Validate` vérifient seulement la forme : `<fournisseur>-<région>`.
- `CheckProvider` vérifie que le fournisseur est bien enregistré, à partir de la liste que l'appelant lui donne.

## Les états de vie

`ern.IsState` et les constantes `ern.Active`, `ern.Deleted`, etc. portent les 7 états communs à tous les services : `PENDING`, `CREATING`, `ACTIVE`, `UPDATING`, `DELETING`, `DELETED`, `FAILED` (Doc 73-12).

## Les jokers

- `*` couvre n'importe quels caractères, et `?` exactement un.
- Ils restent dans un seul morceau : ils ne franchissent jamais `:` ni `/`. Ainsi, `.../instance/*` ne couvre pas `.../instance/db-…/database/ldb-…`.
- `*` seul couvre tout.
- Ils sont permis dans les règles, jamais dans un nom stocké.

## Les cas de test communs

`testdata/vectors.json` contient les cas partagés avec la version TypeScript : noms valides, noms invalides, comparaisons, règles invalides, registre des fournisseurs et états de vie. Les deux versions doivent donner les mêmes résultats.

## Tester

```sh
go test ./...
```

## État

Version locale, pas encore publiée.

**Questions posées au CTO le 5 novembre :**
- **Le repo définitif :** le chemin du module `github.com/ESSINGAN/essingan-ern` est provisoire.
- **La création des ERI :** le plan dit « 16 caractères dérivés d'un ULID », sans dire comment. La bibliothèque vérifie donc les ERI, mais n'en fabrique pas.
- **L'emplacement des cas de test communs.**

**Choix à faire confirmer :** les sous-ressources de niveau compte, comme `user/u-…/mfa-device/mfa-…`, sont acceptées et classées E8. Le contrat les utilise, alors que le Doc 73-01 décrit E8 seulement sur E6 et E7.
