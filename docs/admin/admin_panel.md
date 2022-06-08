# Admin Control Panel

The GoToSocial admin panel is a simple webclient that uses the [admin api routes](https://docs.gotosocial.org/en/latest/api/swagger/#operations-tag-admin) to manage your instance. It uses the same OAUTH mechanism as normal clients (with scope: admin), and as such can be hosted anywhere, separately from your instance, or run locally. A public installation is available here: [https://gts.superseriousbusiness.org/admin](https://gts.superseriousbusiness.org/admin).

## Using the panel
To use the Admin API your account has to be promoted as such:
```
./gotosocial --config-path ./config.yaml admin account promote --username YOUR_USERNAME
```
After this, you can enter your instance domain in the login field (auto-filled if you run GoToSocial on the same domain), and login like you would with
any other client.

<p align="middle">
	<img src="../../assets/admin-panel.png">Screenshot of the GoToSocial admin panel, showing the fields to change an instance's settings</img>
</p>

You can change the instance's settings like the title and descriptions, and add/remove/change domain blocks including a bulk import/export.

## Building the panel
Build requirements: some version of [Node.js](https://nodejs.org) and yarn.
```
yarn install --cwd web/source
BUDO_BUILD=1 node web/source 
```

See also: [Contributing.md Stylesheet / Web dev](https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#stylesheet--web-dev)