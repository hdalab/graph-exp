# graph-exp — демо и CLI

Демо/CLI для модуля [`github.com/hdalab/ga`](https://github.com/hdalab/ga).
Репозиторий закреплён на [ga v0.1.3](https://github.com/hdalab/ga/releases/tag/v0.1.3).
Спецификации могут быть написаны в исходном формате Gexp или в JSON.

## Установка

Собрать и установить CLI можно с помощью:

```bash
go install ./cmd/spath@latest
```

Или локальной сборки:

```bash
go build -o spath ./cmd/spath
./spath --help
```

## Примеры использования

### matrix

Команда `matrix` выводит структурную матрицу графа.

```bash
go run ./cmd/spath matrix -in examples/x.json
```

Пример вывода:

```
    0  1  2  3  4  5
0 [ 0  a  b  0  0  0 ]
1 [ 0  0  c  d  0  0 ]
2 [ 0  0  0  e  f  i ]
3 [ 0  0  0  0  g  0 ]
4 [ 0  0  0  0  0  h ]
5 [ 0  0  0  0  0  0 ]
```

### mdnf

Команда `mdnf` формирует минимальную дизъюнктивную нормальную форму путей между `s` и `t`.

```bash
go run ./cmd/spath mdnf -in examples/x.json --stats
```

Пример выражения:

```
b*i+b*f*h+b*e*g*h+a*d*g*h+a*c*i+a*c*f*h+a*c*e*g*h
```

Пример вывода флагов `--stats`:

```
stats: file=examples/x.json n=6 m=9 s=0 t=5 paths=3 expanded=10 pruned=2 elapsed=1.2ms (0.4µs/path)
```

## Commands
```bash
go run ./cmd/spath matrix -in examples/x.gexp
go run ./cmd/spath matrix -in examples/x.json
go run ./cmd/spath mdnf   -in examples/x.gexp
go run ./cmd/spath mdnf   -in examples/x.json
```

## Metrics
The `mdnf` command can report search metrics. Add `--stats` to print a short summary
or `--stats-json stats.json` to write the full `Stats` structure to a file.

Available fields include:
- `NodesExpanded` — how many nodes were expanded during search
- `Pruned` — how many branches were pruned
- `NsPerPath` — average time per path

Example:
```bash
go run ./cmd/spath mdnf -in examples/x.json --stats --stats-json stats.json
```
Sample `--stats` output:
```
stats: file=examples/x.json n=6 m=9 s=0 t=5 paths=3 expanded=10 pruned=2 elapsed=1.2ms (0.4µs/path)
```
And `stats.json` will contain something like:
```json
{
  "NumPaths": 3,
  "NodesExpanded": 10,
  "Pruned": 2,
  "ElapsedNS": 1200000,
  "NsPerPath": 400
}
```
These metrics help interpret performance: `NodesExpanded` and `Pruned` reflect search effort, while `NsPerPath` shows average time spent per discovered path.
