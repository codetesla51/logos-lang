<p align="center">
  <img src="doc/logo.png" width="80" alt="Logos">
</p>

<h1 align="center">Logos</h1>

<p align="center">
  A readable scripting language with C-like syntax, sane error handling, built-in concurrency, and build-to-binary support.
</p>

---

Logos is what you reach for when bash becomes unreadable after line three.

It is a CLI-first scripting language built for the kind of work bash handles poorly — readable logic, proper error handling, and code you can come back to a week later and still understand.

It is not built for speed. It is built for scripting, and it is good at that.

```logos
let res = fileRead("config.txt")
if !res.ok {
    print(res.error)
    exit(1)
}
print(res.value)
```

## What it has

- C-like syntax that reads like prose
- Built-in file I/O, HTTP, JSON, and shell execution — no imports needed
- Sane error handling via result tables, not exceptions
- Concurrency with `spawn`
- A module system and standard library written in Logos itself
- `lgs build` — compile your script to a standalone binary
- Embeddable in Go applications

## Install

```sh
curl -fsSL https://install.logos-lang.dev | sh
```

## Usage

```sh
lgs script.lgs        # run a file
lgs fmt script.lgs    # format a file
lgs build script.lgs  # compile to binary
lgs                   # start the REPL
```

## License

MIT
