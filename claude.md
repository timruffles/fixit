- the web/layouts package already has header footers, so when writing new templates
  use that rather than defining headers etc
- use the web/handler Wrap function to define high level handlers rather than using the imperative response writer in http package

- use tailwind for css
- use cobra for clis
- use pkgerrors for returning errors not from other packages in this project (e.g. use `pkgerrors.WithStack(err)` on errors from stdlib or third-party libraries)
- use testify assert for test assertions, and use require to prevent nil pointer exceptions etc (i.e. places the test should halt)
- use table-driven tests tests with > 3 cases, where there's a lot of setup. use multiple calls to test fn and assert results if that's much shorter

- be very sparing with log calls. Use slog
- don't add comments unless they're ncessary. e.g. if intention is not obvious from code, or code has side-effects that should be called out. 