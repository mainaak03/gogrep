### gogrep
---
`gogrep` is a lightweight command-line tool written in Go for fast regex-based (and literal) pattern matching in files. Currently it only supports matching in a single file.

This project is currently being extended to use the high **(er)** -performance [Rust `regex` crate](https://docs.rs/regex/) via `cgo` for faster pattern matching.

#### Usage

```
gogrep -p "<pattern>" -f <filename> [additional-flags]
```

#### Example

##### Windows 
```
./gogrep.exe -p "Othello" -f testfiles/shakespeare.txt -n
```
```
(truncated)
[87892]:     Man but a rush against Othello's breast,
[87893]:     And he retires. Where should Othello go?
[87910]:   OTHELLO. That's he that was Othello. Here I am.
[87918]:   LODOVICO. O thou Othello, that wert once so good,
```

#### Flags

| Flag            | Short | Description                                      | Required |
|-----------------|-------|--------------------------------------------------|----------|
| `--pattern`     | `-p`  | Regex pattern or plain string to search for      | Yes      |
| `--filename`    | `-f`  | File to search in                                | Yes      |
| `--line-number` | `-n`  | Prefix matching lines with line numbers          | No       |
| `--ignore-case` | `-i`  | *(currently not implemented)* Case-insensitive matching   | No       |


#### Build
```
go build -o gogrep
```

#### Benchmark

```
hyperfine  \
'grep Othello testfiles/shakespeare.txt > NUL' \
'gogrep.exe -p Othello -f testfiles/shakespeare.txt > NUL' \
--runs 200 \
--warmup 10
```

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `grep Othello testfiles/shakespeare.txt > NUL` | 32.5 ± 3.0 | 28.3 | 43.3 | 1.00 |
| `gogrep.exe -p Othello -f testfiles/shakespeare.txt > NUL` | 46.4 ± 3.6 | 40.3 | 63.8 | 1.43 ± 0.17 |

