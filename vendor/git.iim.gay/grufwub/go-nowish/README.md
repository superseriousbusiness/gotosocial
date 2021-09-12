nowish is a very simple library for creating Go clocks that give a good (ish)
estimate of the "now" time, "ish" depending on the precision you request

similar to fastime, but more bare bones and using unsafe pointers instead of
atomic value since we don't need to worry about type changes