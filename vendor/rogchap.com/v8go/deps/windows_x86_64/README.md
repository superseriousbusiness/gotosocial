
The MINGW patches (`0000`, `0001`, `zlib.gn`) are originally from the [MSYS2 V8 package](https://packages.msys2.org/package/mingw-w64-x86_64-v8?repo=mingw64)
([sources](https://github.com/msys2/MINGW-packages/tree/master/mingw-w64-v8), [LICENSE](https://github.com/msys2/MINGW-packages/blob/master/LICENSE)).

As v8go was already using V8 9.0.257.18 while the MSYS2 package was still at
8.8.278.14-1, they have been slightly updated to work with the newer version.

To create a new version, apply the existing patches as far as possible:

    cd ${v8go_project}/deps/v8
    git apply --reject --whitespace=fix 0000-add-mingw-main-code-changes.patch
    cd build
    git apply --reject --whitespace=fix 0001-add-mingw-toolchain.patch

Check the git output for files that could not be patched cleanly:

    Applying patch config/win/BUILD.gn with 1 reject...

There will be a corresponding .rej file (`config/win/BUILD.gn.rej`) that
contains the rejected parts of the patch. Apply the necessary changes
manually and delete the .rej file afterwards.

Once all changes are complete, create the new patches from the modified working
tree:

    cd ${v8go_project}/deps/v8
    git diff --relative >../windows_x86_64/0000-add-mingw-main-code-changes.patch
    cd build
    git diff --relative >../../windows_x86_64/0001-add-mingw-toolchain.patch

Optional: To minimize whitespace/filemode diffs to the original MINGW patches,
run:

    sed -i 's/[ \t]*$//' <patch file>
    sed -i 's/100755/100644/' <patch file>

Note: `git reset --hard` the `v8` and `v8/build` submodules if you want to run
`build.py` after this. Otherwise the build will complain about the unstaged
changes.
