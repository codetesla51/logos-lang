<p align="center">
  <img src="doc/logo.png" width="140" alt="Logos">
</p>

<h1 align="center">Logos</h1>

<p align="center">
  A readable scripting language with C-like syntax, sane error handling, built-in concurrency, and build-to-binary support.
</p>

<p align="center">
  <a href="https://logos-lang.dev/docs">Documentation</a> &nbsp;|&nbsp;
  <a href="https://github.com/codetesla51/logos-lang">GitHub</a> &nbsp;|&nbsp;
  <code>lgs</code>
</p>

---

**Logos** (from Greek, meaning word, reason, or logic) is a scripting language designed to be read and understood, not deciphered.

It is what you reach for when bash becomes unreadable after line three. A CLI-first language built for the kind of work bash handles poorly: readable logic, proper error handling, and code you can come back to a week later and still understand.

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
- Built-in file I/O, HTTP, JSON, and shell execution with no imports needed
- Sane error handling via result tables, not exceptions
- Concurrency with `spawn`
- A module system and standard library written in Logos itself
- `lgs build` to compile your script to a standalone binary
- Embeddable in Go applications like Lua

## Install

```sh
curl -fsSL https://install.logos-lang.dev | sh
```

Works on Linux and macOS.

## Usage

```sh
lgs script.lgs        # run a file
lgs fmt script.lgs    # format a file
lgs build script.lgs  # compile to binary
lgs                   # start the REPL
```

## Documentation

Full documentation is available at [logos-lang.dev/docs](https://logos-lang.dev/docs).

| Page | Description |
|------|-------------|
| [Installation](https://logos-lang.dev/docs/install) | Install Logos and get started |
| [Syntax](https://logos-lang.dev/docs/syntax) | Language tour and syntax reference |
| [Standard Library](https://logos-lang.dev/docs/stdlib) | Built-in modules and functions |
| [Embedding](https://logos-lang.dev/docs/embedding) | Embed Logos in a Go application |
| [Building Binaries](https://logos-lang.dev/docs/building) | Compile scripts to standalone binaries |
| [Examples](https://logos-lang.dev/docs/examples) | Real scripts and example programs |

## License

MIT
