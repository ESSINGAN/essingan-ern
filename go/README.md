# essingan-ern, la version Go

La bibliothèque Go des **ERN** (ESSINGAN Resource Name), les noms uniques des ressources d'Essingan, sur le modèle des ARN d'AWS, et des **ERI**, leurs identifiants courts.

- **Module :** `github.com/ESSINGAN/essingan-ern/go`, paquet `ern`.
- **Règles :** ADR-0051 (D1 à D12) et Doc 73-01 (repo `sre-devops-assistant`), contrat de l'API du Hub.
- **Pour qui :** management-api, essingan-iam, cno et les autres services en Go. La version TypeScript (Eric) vit dans `ts/`.
- **Ce qu'elle n'est pas :** ni un service, ni une base de données. Elle est intégrée au code de chaque programme et ne fait aucun appel réseau.

```go
import ern "github.com/ESSINGAN/essingan-ern/go"
```

## Ce qu'elle fait

| Fonction | Exemple | Rôle |
|---|---|---|
| Créer un identifiant | `ern.NewERI("p")` | Crée l'identifiant d'une nouvelle ressource, ex. `p-01m6r2gsg13nqk8n` |
| Nommer un objet qui existe déjà | `ern.DeriveERI("u", créé_le, id_stable)` | Donne toujours le même ERI au même objet, sans rien stocker |
| Fabriquer | `ern.NewProject(compte, org, projet)` | Crée le nom d'une ressource, ou renvoie une erreur |
| Lire | `ern.Parse(texte)` | Découpe un nom en morceaux |
| Valider | `e.Validate()`, `ern.IsERI(id)`, `ern.IsERIOf("p", id)` | Refuse un nom ou un identifiant mal formé |
| Vérifier le fournisseur | `e.CheckProvider(registre)` | Refuse un fournisseur cloud absent du registre |
| Comparer | `ern.Match(règle, nom)` | Dit si un nom correspond à une règle de droits |
| Vérifier une règle | `ern.ValidatePattern(règle)` | Refuse une règle de droits que `Match` ne saurait pas lire, avant de l'enregistrer |
| Lister les services | `ern.Services()`, `ern.IsService("iam")` | Les espaces de noms (`iam`, `account`, `scu`…), aussi ceux des actions `iam:CreateUser` |
| Afficher un compte | `ern.DisplayAccountID("482019375521")` | Écrit `4820-1937-5521` pour les personnes |

```go
projet, err := ern.NewProject("482019375521", "o-01jd2v8k4m7q9x3a", "p-01jf3w9n5r8t2y6b")
// ern:essingan:account::482019375521:o-01jd2v8k4m7q9x3a/project/p-01jf3w9n5r8t2y6b

ok, err := ern.Match("ern:essingan:account::482019375521:o-01jd2v8k4m7q9x3a/project/*", projet.String())
// ok == true
```

Si `Match` renvoie une erreur, la règle ou le nom est invalide : l'appelant doit refuser l'accès.

## Les identifiants (ERI)

Un ERI, c'est un préfixe de 1 à 4 lettres, un tiret, puis les 16 premiers caractères d'un ULID (ADR-0051 D10) :

```
p-01m6r2gsg0 13nqk8n
  └─ 10 ──┘ └─ 6 ─┘
   la date   tirés au hasard
   (à la ms)
```

- **Les 10 premiers :** l'heure de création, à la milliseconde, codée en base 32.
- **Les 6 suivants :** tirés au hasard.
- **Alphabet :** base32 de Crockford en minuscules, sans `i`, `l`, `o` ni `u`.
- **Triés par date :** trier les ERI comme du texte les range par date de création. La base peut donc lister et paginer dans l'ordre sans colonne de plus. Deux ERI créés dans la même milliseconde restent dans un ordre au hasard.
- **Index unique :** un doublon est très peu probable, mais pas impossible. La table garde donc un index unique sur l'ERI, et le service recommence en cas de doublon.
- `ern.ERITime(id)` relit la date de création d'un ERI.
- `ern.DeriveERI` remplace le hasard par une empreinte (SHA-256) d'un identifiant stable. L'IAM Essingan s'en sert pour les objets qui existent déjà dans son moteur. La formule est celle d'essingan-iam #129, pour que l'IAM Essingan passe à cette bibliothèque sans renommer quoi que ce soit.

## Les partitions

`essingan` (Hub de prod), `essingan-staging` (Hub bac à sable), `essingan-dev` (dev-001), et `essingan-<id>` pour une copie du plan de gestion chez un client (ADR-0051 D11).

## Les fournisseurs cloud

La liste n'est pas dans le code : elle vit dans le registre du catalogue (ADR-0051 D7). Ajouter un fournisseur ne demande donc aucune nouvelle version de la bibliothèque.

- `Parse` et `Validate` vérifient seulement la forme : `<fournisseur>-<région>`.
- `CheckProvider` vérifie que le fournisseur est bien enregistré, à partir de la liste que l'appelant lui donne.

## Les états de vie

`ern.IsState` et les constantes `ern.Active`, `ern.Deleted`, etc. portent les 7 états communs à tous les services : `PENDING`, `CREATING`, `ACTIVE`, `UPDATING`, `DELETING`, `DELETED`, `FAILED` (Doc 73-12).

## Les jokers

Ils sont permis dans les règles, jamais dans un nom stocké (ADR-0051 D10).

| Joker | Couvre | Exemple |
|---|---|---|
| `*` seul | Tout | `*` |
| `*` | N'importe quels caractères, dans un seul morceau : il ne franchit jamais `:` ni `/` | `…/instance/*` : les instances, pas ce qu'elles contiennent |
| `?` | Exactement un caractère | `…/db-01jg4x8n5r8t2y6?` |
| `**` | Un ou plusieurs morceaux entiers, seulement en dernier | `…/instance/db-…/**` : tout ce qui vit dans l'instance, pas l'instance elle-même |

`**` suit la règle d'un `/**` final dans `.gitignore` : il couvre ce qu'il y a dessous. Pour l'instance et son contenu, il faut deux règles : `…/instance/db-…` et `…/instance/db-…/**`.

## Les cas de test communs

`../testdata/vectors.json`, à la racine du dépôt, est le seul fichier de cas de test. La version Go et la version TypeScript le passent toutes les deux, et doivent donner les mêmes réponses. Un nouveau cas s'ajoute d'abord dans ce fichier.

Il est rangé par thèmes :

| Thème | Contenu |
|---|---|
| `valid`, `parts` | Des noms corrects, leur forme (E1 à E10) et leurs morceaux |
| `invalid` | Des noms à refuser, avec la raison (`why`) |
| `match`, `bad_patterns` | Des règles de droits, ce qu'elles couvrent, et des règles à refuser |
| `registry` | Des fournisseurs enregistrés ou non |
| `eri`, `derived_eri` | Une date et un nombre (ou un identifiant stable), et l'ERI attendu |
| `display_account_ids` | Un numéro de compte et son affichage |
| `lifecycle_states`, `bad_lifecycle_states` | Les 7 états de vie, et des états à refuser |

## Tester

```sh
cd go && go test ./...
```

## Choix à faire confirmer

Les sous-ressources de niveau compte, comme `user/u-…/mfa-device/mfa-…`, sont acceptées et classées E8. Le contrat les utilise, alors que le Doc 73-01 décrit E8 seulement sur E6 et E7. essingan-iam #129 fait le même choix.
