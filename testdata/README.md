# testdata

Tests are organized like this:

## `file/`

```
test-name/
    gold.json    # expected output
    flags        # -var flags to pass to temple
    src.tpl.json # template to render
```

## `tree/`

```
test-name/
    flags  # -var flags to pass to temple
    src/   # template files to render
    dst/   # files that already exist at destination
    gold/  # expected output tree
```
