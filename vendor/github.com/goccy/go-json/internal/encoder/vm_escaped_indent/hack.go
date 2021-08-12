package vm_escaped_indent

// HACK: compile order
// `vm`, `vm_escaped`, `vm_indent`, `vm_escaped_indent` packages uses a lot of memory to compile,
// so forcibly make dependencies and avoid compiling in concurrent.
// dependency order: vm => vm_escaped => vm_indent => vm_escaped_indent
