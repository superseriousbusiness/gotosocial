# 无 Wazero / WASM 版构建

!!! danger "不受支持"
    我们不提供对使用本节描述的 `nowasm` 标签构建的 GoToSocial 部署的任何支持。这样的构建在任何情况下都应被视为实验性构建，任何运行时出现的问题与我们无关！请不要在存储库中提交寻求 `nowasm` 构建的调试帮助的相关问题。

在[支持的平台](../../getting_started/releases.md#支持的平台)上，GoToSocial 使用 WebAssembly 运行时 [Wazero](https://wazero.io/) 对 `ffmpeg`、`ffprobe` 和 `sqlite3` WebAssembly 二进制文件进行沙盒化，使这些应用程序可以被打包并在 GoToSocial 二进制文件中运行，无需管理员安装和管理任何外部依赖。

这使得管理员更容易维护他们的 GoToSocial 实例，因为他们的 GtS 二进制文件完全与系统安装的 `ffmpeg`、`ffprobe` 和 `sqlite` 的更改隔离开来。以这种方式运行 `ffmpeg` 也更安全一些，因为 GoToSocial 将 `ffmpeg` 二进制文件封装在一个非常受限的文件系统中，该系统不允许 `ffmpeg` 二进制文件访问除正在解码和重新编码的文件以外的任何文件。换句话说，在受支持的平台上，GoToSocial 提供了 `ffmpeg` 等的大多数功能，而不存在一些麻烦。

然而，并不是所有的平台都能在速度更快的“编译器”模式下运行 Wazero，因此必须使用非常慢（且资源占用大的）“解释器”模式。有关符合性的详细信息，请参考 Wazero 的[此表](https://github.com/tetratelabs/wazero?tab=readme-ov-file#conformance)。

“解释器”模式的运行性能非常差，以至于在不是 64 位 Linux 或 64 位 FreeBSD 的平台上运行 GoToSocial 实例是不切实际的，因为所有的内存和 CPU 都被媒体处理消耗殆尽。

但是！为了让用户能够运行**实验性、不受支持的 GoToSocial 部署**，我们开放了 `nowasm` 构建标签，该标签可用于编译完全不使用 Wazero 或 WASM 的 GoToSocial 构建。

使用 `nowasm` 构建的 GoToSocial 二进制文件将使用 [modernc 版本的 SQLite](https://pkg.go.dev/modernc.org/sqlite) 而不是 WASM 版本，并将在系统上使用 `ffmpeg` 和 `ffprobe` 二进制文件进行媒体处理。

要使用 `nowasm` 标签构建 GoToSocial，可以像这样将标签传入我们的便利 `build.sh` 脚本：

```bash
GO_BUILDTAGS=nowasm ./scripts/build.sh
```

要运行以此方式构建的 GoToSocial 版本，你必须确保在主机上安装了 `ffmpeg` 和 `ffprobe`。这通常只需运行类似 `doas -u root pkg_add ffmpeg`（OpenBSD）或 `sudo apt install ffmpeg`（Debian 等）的命令即可。

!!! danger "确实不受支持"
    再次强调，如果在你的操作系统/架构组合上运行 `nowasm` 构建的 GoToSocial 有效，那很好，但我们不会为这样的构建提供支持，也无法帮助调试为何某些功能不起作用。
