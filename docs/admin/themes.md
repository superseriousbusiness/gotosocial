# Themes

Users on your instance can select a theme for their profile from any css files present in the `web/assets/themes` directory.

GoToSocial comes with some theme files already, but you can add more yourself by doing the following:

1. Create a file in `web/assets/themes` called (for example) `new-theme.css`.
2. (Optional) Include the following comment at the top of your theme file to title and describe your theme:
  ```css
  /*
    theme-title: My New Theme
    theme-description: This is an example theme
  */
  ```
  You can use any text you like for these fields, but bear in mind whatever you write here will appear in the settings panel to help users when selecting a theme, so keep it short and sweet.
3. Fill out your custom CSS in the rest of the file. You can use one of the existing CSS files to guide you. Also see [this page](../user_guide/custom_css.md) for some rough guidelines about how to write accessible CSS.
4. Restart your instance so that the new CSS file is picked up.

!!! info
    If you're using Docker for your deployment, you can mount theme files from the host machine into your GoToSocial `web/assets/themes` directory instead, by including entries for them in the `volumes` section of your Docker configuration.
    
    For example, say you've created a theme on your host machine at `~/gotosocial/my-themes/new-theme.css`, you could mount that theme into the GoToSocial Docker container in the following way:
    
    ```yaml
    volumes:
      [.... some other volume entries ...]
      - "~/gotosocial/my-themes/new-theme.css:/gotosocial/web/assets/themes/new-theme.css"
    ```
    
    Bear in mind if you mount an entire directory to `/gotosocial/web/assets/themes` instead of mounting individual theme files, you'll override the default themes.
