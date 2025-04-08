# Build without Wazero / WASM

!!! Danger "This is unsupported"
    We do not offer any kind of support for deployments of GoToSocial built with the `nowasm` tag described in this section. Such builds should be considered strictly experimental, and any issues that come when running them are none of our business! Please don't open issues on the repo looking for help debugging deployments of `nowasm` builds.

On [supported platforms](../../getting_started/releases.md#supported-platforms), GoToSocial uses the WebAssembly runtime [Wazero](https://wazero.io/) to sandbox `ffmpeg`, `ffprobe`, and `sqlite3` WebAssembly binaries, allowing these applications to be packaged and run inside the GoToSocial binary, without requiring admins to install + manage any external dependencies.

This has the advantage of making it easier for admins to maintain their GoToSocial instance, as their GtS binary is completely isolated from any changes to their system-installed `ffmpeg`, `ffprobe`, and `sqlite`. It's also a bit safer to run `ffmpeg` in this way, as GoToSocial wraps the `ffmpeg` binary in a very constrained file system that doesn't permit the `ffmpeg` binary to access any files other than the ones it's decoding + reencoding. In other words, GoToSocial on supported platforms offers most of the functionality of `ffmpeg` and so on, without some of the headaches.

However, not all platforms are capable of running Wazero in the much-faster "compiler" mode, and have to fall back to the very slow (and resource-heavy) "interpreter" mode. See [this table](https://github.com/tetratelabs/wazero?tab=readme-ov-file#conformance) from Wazero for conformance.

"Interpreter" mode runs so poorly for GoToSocial's use case that it's simply not feasible to run a GoToSocial instance in a stable manner on platforms that aren't 64-bit Linux or 64-bit FreeBSD, as all the memory and CPU get gobbled up by media processing.

However! To enable folks to run **experimental, unsupported deployments of GoToSocial**, we expose the `nowasm` build tag, which can be used to compile a build of GoToSocial that does not use Wazero or WASM at all.

A GoToSocial binary built with `nowasm` will use the [modernc version of SQLite](https://pkg.go.dev/modernc.org/sqlite) instead of the WASM one, and will use on-system `ffmpeg` and `ffprobe` binaries for media processing.

!!! tip
    To test if your system is compatible with the standard builds, you can use this command:
    `if grep -qE '^flags.* (sse4|LSE)' /proc/cpuinfo; then echo "Your system is supporting GTS!"; else echo "Your system is not supporting GTS, you'll have to use the 'nowasm' builds :("; fi`

To build GoToSocial with the `nowasm` tag, you can pass the tag into our convenience `build.sh` script like so:

```bash
GO_BUILDTAGS=nowasm ./scripts/build.sh
```

In order to run a version of GoToSocial built in this way, you must ensure that `ffmpeg` and `ffprobe` are installed on the host. This is usually as simple as running a command like `doas -u root pkg_add ffmpeg` (OpenBSD), or `sudo apt install ffmpeg` (Debian etc.). 

!!! Danger "No really though, it's unsupported"
    Again, if running builds of GoToSocial with `nowasm` works for your OS/Arch combination, that's great, but we do not support such builds and we won't be able to help debugging why something doesn't work.
