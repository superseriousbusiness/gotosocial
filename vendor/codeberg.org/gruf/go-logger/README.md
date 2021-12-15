Fast levelled logging package with customizable formatting.

Supports logging in 2 modes:
- no locks, fastest possible logging, no guarantees for io.Writer thread safety
- mutex locks during writes, still far faster than standard library logger

Running without locks isn't likely to cause you any issues*, but if it does, you can wrap your `io.Writer` using `AddSafety()` when instantiating your new Logger. Even when running the benchmarks, this library has no printing issues without locks, so in most cases you'll be fine, but the safety is there if you need it.

*most logging libraries advertising high speeds are likely not performing mutex locks, which is why with this library you have the option to opt-in/out of them.

Note there are 2 uses of the unsafe package:
- safer interface nil value checks, uses similar logic to reflect package to check if the value in the internal fat pointer is nil
- casting a byte slice to string to allow sharing of similar byte and string methods, performs same logic as `strings.Builder{}.String()`