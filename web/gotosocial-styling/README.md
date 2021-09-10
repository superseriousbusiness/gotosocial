# GoToSocial Styling

Common package for the PostCSS styling of GoToSocial (related) pages.

## Bundle
Source in `src/style.css` is bundled by running `node index.js`. Output appears in `build/bundle.css`, and can be required from other packages with `require("gotosocial-styling/build/bundle.css")`.

## Development
You can run `NODE_ENV=development node index.js` to start a livereloading setup that automatically re-bundles on file changes in `src/`.

## License, donations
[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html). If you want to support my work, you can:  
<a href="https://liberapay.com/f0x/donate"><img alt="Donate using Liberapay" src="https://liberapay.com/assets/widgets/donate.svg"></a>


## Changelog
### v0.0.1 (August 29th, 2021)
initial release
