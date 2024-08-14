<!-- deno-fmt-ignore-start -->

> [!WARNING]
> NOT PRODUCTION READY. I **DO NOT CARE** ABOUT AND **CANNOT BE HELD RESPONSIBLE**
> FOR ANY DAMAGE TO YOUR SYSTEM OR DATA BY ATTEMPTING TO RUN THIS PROGRAM!

<!-- deno-fmt-ignore-end -->

# bgenix

A Go implementation of Agenix.

## Motivation

I was bored.

_Also I find Bash to be an insanely unwieldly choice for everyday tools. A 210~
LoC script should not be handling the entirety of my secrets management... No
shade, just personal preference._

## Goals & Non-goals

In the order that they are given.

### Goals

- Works on _my_ machine
- Not bash
- Relatively fast
- Structured and readable code

### Non-goals

- Works on _your_ machine
- Rage support
- Anything I don't care about

## TODO

- Better configuration management
- Properly implement `-r`/`--rekey`
- Try using Age as a Go library instead of the Age binary
- Flake-parts support (?)
  - See agenix-rekey

## License

bgenix is licensed under [GPL3](LICENSE).

Credits for design and implementation ideas go to
[@ryantm](https://github.com/ryantm) for the creation of
[agenix](https://github.com/ryantm/agenix).
