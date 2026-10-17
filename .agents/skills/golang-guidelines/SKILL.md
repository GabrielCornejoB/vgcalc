---
name: golang-guidelines
description: Guidelines for writing golang code for this repo
---

## Golang guidelines

### Slice index vs len(slice)-1 for lookup maps

When building a slice plus a `map[id]index` in the same loop:

- 1:1 transform (every source row appends exactly once, no `continue`/filter): use the range index `i`.
  ```go
  for i, row := range rows {
      abilities = append(abilities, Ability{PokeapiID: row.ID, Slug: row.Identifier})
      abilitiesMap[row.ID] = i
  }
  ```
- Filter/skip loop (any `continue`, dedupe, conditional append, e.g. `seenSlugs` in `loadItems`): input index diverges from output index, so use `len(slice)-1` after `append`.
  ```go
  for _, row := range rows {
      if _, seen := seenSlugs[row.Identifier]; seen {
          continue
      }
      items = append(items, Item{PokeapiID: row.ID, Slug: row.Identifier})
      itemsMap[row.ID] = len(items) - 1
  }
  ```
