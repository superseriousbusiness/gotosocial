# GoToSocial Admin Panel

Standalone web admin panel for [GoToSocial](https://github.com/superseriousbusiness/gotosocial).

A public hosted instance is also available at https://gts.superseriousbusiness.org/admin/, so you can fill your own instance URL in there.

## Installation
Build requirements: some version of Node.js with npm,
```
git clone https://github.com/superseriousbusiness/gotosocial-admin.git && cd gotosocial-admin
npm install
node index.js
```
All processed build output will now be in `public/`, which you can copy over to a folder in your GoToSocial installation like `web/assets/admin`, or serve elsewhere.
No further configuration is required, authentication happens through normal OAUTH flow.

## Development
Follow the installation steps, but run `NODE_ENV=development node index.js` to start the livereloading dev server instead.

## License, donations
[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html). If you want to support my work, you can: <a href="https://liberapay.com/f0x/donate"><img alt="Donate using Liberapay" src="https://liberapay.com/assets/widgets/donate.svg"></a>