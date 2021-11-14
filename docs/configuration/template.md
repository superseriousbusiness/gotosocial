# Template

## Settings

```yaml
###############################
##### WEB TEMPLATE CONFIG #####
###############################

# Config pertaining to templating of web pages/email notifications and the like
template:

  # String. Directory from which gotosocial will attempt to load html templates (.tmpl files).
  # Examples: ["/some/absolute/path/", "./relative/path/", "../../some/weird/path/"]
  # Default: "./web/template/"
  baseDir: "./web/template/"

  # String. Directory from which gotosocial will attempt to serve static web assets (images, scripts).
  # Examples: ["/some/absolute/path/", "./relative/path/", "../../some/weird/path/"]
  # Default: "./web/assets/"
  assetBaseDir: "./web/assets/"
```
