drop-in replacement for standard "bytes" library

contains alternative Buffer implementation that provides direct access to the
underlying byte-slice, with some interesting alternative struct methods. provides
no safety guards, if you pass bad values it will blow up in your face...

and alternative `ToUpper()` and `ToLower()` implementations that use lookup
tables for improved performance

provides direct call-throughs to most of the "bytes" library functions to facilitate
this being a direct drop-in. in some time, i may offer alternative implementations
for other functions too