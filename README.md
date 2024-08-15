<!-- deno-fmt-ignore-start -->

> [!WARNING]
> NOT PRODUCTION READY. I **DO NOT CARE** ABOUT AND **CANNOT BE HELD RESPONSIBLE**
> FOR ANY DAMAGE TO YOUR SYSTEM OR TO YOUR DATA CAUSED BY ATTEMPTING TO RUN THIS
> PROGRAM! YOU HAVE BEEN WARNED AND _STRONGLY_ DISCOURAGED.

<!-- deno-fmt-ignore-end -->

# bgenix

A fast, maintainable and structured implementation of Agenix in Go. Aims to
replace Agenix with minimal effort and minimal compromise. Heavily WIP and in a
Proof of Concept state.

_the B also stands for better name pending_

## Motivation

I was bored.

_Also I find Bash to be an insanely unwieldly choice for everyday tools. A 210~
LoC script should not be handling the entirety of my secrets management... No
shade, just personal preference. Go on another hand is a tool I'm familiar with
and is somewhat on the eyes. Rust could've been a better choice, but I am not at
all in the mood to wrestle the borrow checker for something this simple._

## Goals & Non-goals

In the order that they are given.

### Goals

- Works on _my_ machine
- Not bash
- Relatively fast
- Structured and readable code
- Somewhat stable
- Does not blow up your system

### Non-goals

- Works on _your_ machine
- Rage support
  - In my experience Rage is not at all backwards compatible with Age. As such,
    I do not care to support it.
- Anything I don't care about
- Mainstream usage

## TODO

- Better configuration management
- Properly implement `-r`/`--rekey`
- Flake-parts support (?)
  - See agenix-rekey

## Contributing

Feel free. This is _probably_ not going to replace Agenix on your system but it
will on mine and _might_ on others' systems. The goal is to have a working
Agenix replica in Go.

There is no stability promise, but it's ultimately a goal.

## License

bgenix is licensed under [GPL3](LICENSE).

### Attributions

[@ryantm]: https://github.com/ryantm
[agenix]: https://github.com/ryantm/agenix

Credits for design and implementation ideas go to [@ryantm] for the creation of
[agenix].
