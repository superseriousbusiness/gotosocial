# Login

## Credentials

_admin_ user at your instance will inform you about your credentials:

* **username**: user's email (NOT your _@handle@gts.instance_).
* **password**: set at user's creation (you can change it later).

## Clients 

At this moment both [Pinafore](https://pinafore.social/) (web client) and [Tusky](https://tusky.app/) (android) are the _goto-clients_ 😉 of reference.

[Mastodon](https://joinmastodon.org/apps)'s Official Apps can also be used, but it lacks (ATM) some features _tusky_ has.

In any of these apps the procedure is transparent for the user: Enter credentials and **grant permission** to access your account.

### Toot CLI 

[toot cli](https://toot.readthedocs.io/en/latest/index.html) is 

> a CLI and TUI tool for interacting with Mastodon instances from the command line.

so you can use it in programs, scripts, alias, etc. 

In order to use _toot_ you need a _auth key_ from your instance.

After [installing toot](https://toot.readthedocs.io/en/latest/install.html) open a terminal and write (as regular user) and execute `toot login` to authorize _toot_ to use your account

```
[user@host]$ toot login
Creating config file at /home/user/.config/toot/config.json
Choose an instance [mastodon.social]: gts.tld.org
Looking up instance info...
Found instance GoToSocial' Social Instance running Mastodon version 0.5.2
Registering application...
Application tokens saved.

This authentication method requires you to log into your Mastodon instance
in your browser, where you will be asked to authorize toot to access
your account. When you do, you will be given an authorization code
which you need to paste here.

This is the login URL:
https://gts.tld.org/oauth/authorize/?response_type=code&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&scope=read+write+follow&client_id=CLIENTKEYVERYLONGHEXVALUE

Open link in default browser? [Y/n]
Authorization code: 
```

if you hit _Enter_ (Y default), then your system's default browser opens a **login page in your instance**. 

if you select _n_ (press _n_ then _Enter_), browser is not opened and prompt keeps awaiting for your _authorization code_. You can _copy/paste_ in any browser the complete **login URL** to **get** _auth code_:

In both cases the result is the same. **Keep** this terminal opened and go to your browser.

**Before** even writing your credentials, **press F12** and select _tab_ **Console** in the development tools of your browser. Keep it opened so you can see what's happening and see _auth code_.

**Now**, enter credentials and **Allow** your user to request **authorization code**

Back in _Console_, _Headers_ tab, there's a **POST** element with a dropdown arrow; click on it to open this element. You've to look for `location urn:ietf:wg:oauth:2.0:oob?code=YMQ2N2E3-VERYLONGAUTHKEY-NINTQ0MTA3MWI0` info.

**This is the Auth Code you need** to paste in `toot` prompt

![firefox's DevTools screenshot](https://blog.xmgz.eu/assets/imaxes/toot_login.png "firefox screenshot with auth code")


```
Authorization code: YMQ2N2E3-VERYLONGAUTHKEY-NINTQ0MTA3MWI0

Requesting access token...
Access token saved to config at: /home/user/.config/toot/config.json

✓ Successfully logged in.
```

press _Enter_ and then you should be able to post from **toot cli**.
