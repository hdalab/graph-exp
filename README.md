# graph-exp — demo & CLI

Demo/CLI for the `github.com/hdalab/ga` module.
Specs may be written either in the original Gexp format or in JSON.

## Commands
```bash
go run ./cmd/spath matrix -in examples/x.gexp
go run ./cmd/spath matrix -in examples/x.json
go run ./cmd/spath mdnf   -in examples/x.gexp
go run ./cmd/spath mdnf   -in examples/x.json
```

# Метрики
Добавьте `--stats` к команде `mdnf`, чтобы увидеть краткие метрики в stderr, или `--stats-json file.json` — чтобы сохранить полный объект Stats в JSON.
