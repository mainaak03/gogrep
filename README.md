### gogrep
---
`gogrep` is a lightweight command-line tool written in Go and C for fast regex-based (and literal) pattern matching in files.

Under the hood, it uses the high **(er)** -performance Rust [regex](https://docs.rs/regex/) crate via `cgo` for pattern matching.

#### Usage
```
gogrep "<pattern>" <file> [additional-flags]
```

#### Example

##### Linux
```
./gogrep "Othello" testfiles/shakespeare.txt
```
```
(truncated)
testfiles/shakespeare.txt:     And he retires. Where should Othello go?
testfiles/shakespeare.txt:   OTHELLO. That's he that was Othello. Here I am.
testfiles/shakespeare.txt:   LODOVICO. O thou Othello, that wert once so good,
```

#### Flags

| Flag            | Short | Description                                      | Required |
|-----------------|-------|--------------------------------------------------|----------|
| `--ignore-case` | `-i`  | Enable case insensitive matching   | No       |
| `--recursive` | `-r`  |  Search recursivley inside dirs   | No       |
| `--line-number` | `-n`  | *(currently not implemented)* Prefix matching lines with line numbers          | No       |


#### How to build and run locally
1. Clone [rust regex engine](https://github.com/rust-lang/regex) repo and this repo in the same directory

2. Build the regex engine with
```
cargo build --release --manifest-path ./regex/Cargo.toml
```
This will generate the necesary `librure.so` file that we will need later while linking.

3. Set the following environment variables to let the linker know where our .so file lives
```
export LD_LIBRARY_PATH="$(pwd)/regex/target/release"
export CGO_LDFLAGS="-L$(pwd)/regex/target/release"
```

4. Inside the gogrep repo, build it with
```
go build -o gogrep
```

#### Benchmark

Some simple benchmarks are provided, these are not subjective and do NOT mean this tool is in any way better or comparable to the big boys (grep, ripgrep).

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `./gogrep "\w+\s+Othello" testfiles/shakespeare.txt` | 9.3 ± 2.4 | 6.5 | 18.9 | 1.38 ± 0.54 |
| `rg "\w+\s+Othello" testfiles/shakespeare.txt` | 6.7 ± 2.0 | 4.5 | 27.7 | 1.00 |
| `grep -P "\w+\s+Othello" testfiles/shakespeare.txt` | 17.8 ± 2.8 | 13.9 | 31.0 | 2.65 ± 0.88 |

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `./gogrep "Othello" testfiles/shakespeare.txt` | 7.1 ± 2.0 | 4.9 | 18.6 | 2.02 ± 0.86 |
| `rg "Othello" testfiles/shakespeare.txt` | 3.5 ± 1.1 | 2.0 | 11.3 | 1.00 |
| `grep -P "Othello" testfiles/shakespeare.txt` | 4.8 ± 1.6 | 3.3 | 16.7 | 1.38 ± 0.63 |

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `./gogrep "\b\w{5,}\b" testfiles/shakespeare.txt` | 40.4 ± 2.8 | 33.6 | 54.5 | 50.29 ± 21.47 |
| `rg "\b\w{5,}\b" testfiles/shakespeare.txt` | 24.5 ± 4.2 | 19.0 | 43.3 | 30.56 ± 13.91 |
| `grep -P "\b\w{5,}\b" testfiles/shakespeare.txt` | 0.8 ± 0.3 | 0.4 | 3.9 | 1.00 |

