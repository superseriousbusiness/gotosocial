# Web

Starting with v0.7.0, gotosocial embeds all assets for the web frontend inside the executable.
The configuration defines directories where assets additionally can be read from to allow for customization of the frontend.
Head over to [Frontend Customization](TODO) to learn more about this use case.

## Settings

```yaml
######################
##### WEB CONFIG #####
######################

# Config pertaining to templating and serving of web pages/email notifications and the like

# String. Directory from which gotosocial will attempt to load html templates (.tmpl files),
#   overriding the embedded templates.
# Examples: ["/some/absolute/path/", "./relative/path/", "../../some/weird/path/"]
# Default: "./web/template/"
web-template-base-dir: "./web/template/"

# String. Directory from which gotosocial will attempt to serve static web assets (images, scripts).
#   overriding the embedded assets.
# Examples: ["/some/absolute/path/", "./relative/path/", "../../some/weird/path/"]
# Default: "./web/assets/"
web-asset-base-dir: "./web/assets/"
```
