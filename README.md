# charts

On macOS, this requires some `gnu`-based tools:

1. [`gnu-tar`](https://formulae.brew.sh/formula/gnu-tar)
1. [`gnu-sed`](https://formulae.brew.sh/formula/gnu-sed)

```
export PATH="/opt/homebrew/opt/gnu-tar/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/gnu-sed/libexec/gnubin:$PATH"
```

On CI, this runs a scheduled job: `DIST=gh-pages go run mage.go import`.
