gopherjs-live
-------------
Automatic watching and recompiling for gopherjs

The `-w` option for gopherjs wasn't working for me.
LiveReload supports custom commands but would fail silently.

This tool:

- Watches for changes in your current directory (and any directories in it)
- Automatically recompiles when something changes
- Only watches go files
- Is safe for use with editors which use atomic saves
- Passes along any flags and arguments to the `gopherjs build` command
- Prints out any errors to the console and chimes to let you know there was one


### Installation

- Install go version >= 0. Latest version is recommended.
- Follow [these instructions](https://golang.org/doc/code.html) for setting up your go workspace
- Run `go get -u github.com/albrow/gopherjs-live`

### Usage

From a directory with some gopherjs-compatible go code in it, just run `gopherjs-live`.
Any arguments and flags will be passed through, so you can run something like this:
`gopherjs-live -m -o js/app.js`. See `gopherjs help build` for more information about the
supported flags.
