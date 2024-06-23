# Split-domain deployments

This guide explains how to have usernames like `@me@example.org` but run the GoToSocial instance itself on a subdomain like `social.example.org`. Configuring this type of deployment layout **must** be done before starting GoToSocial for the first time.

!!! danger
    You cannot change your domain layout after you've federated with someone. Servers are going to get confused and you'll need to convince the admin of every instance that's federated with you before to mess with their database to resolve it. It also requires regenerating the database on your side to create a new instance account and pair of encryption keys.

## Background

The way ActivityPub implementations discover how to map your account domain to your host domain is through a protocol called [webfinger](https://www.rfc-editor.org/rfc/rfc7033). This mapping is typically cached by servers and hence why you can't change it after the fact.

It works by doing a request to `https://<account domain>/.well-known/webfinger?resource=acct:@me@example.org`. At this point, a server can return a redirect to where the actual webfinger endpoint is, `https://<host domain>/.well-known/webfinger?resource=acct:@me@example.org` or may respond directly. The JSON document that is returned informs you what the endpoint to query is for the user:

```json
{
  "subject": "acct:me@example.org",
  "aliases": [
    "https://social.example.org/users/me",
    "https://social.example.org/@me"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://social.example.org/@me"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://social.example.org/users/me"
    }
  ]
}
```

ActivityPub clients and servers will now use the entry from the `links` array with `rel` `self` and `type` `application/activity+json` to query for further information, like where the `inbox` is located to federated messages to.

## Configuration

There are 2 configuration settings you'll need to concern yourself with:

* `host`, the domain the API will be served on and what clients and servers will end up using when talking to your instance
* `account-domain`, the domain user accounts will be created on

In order to achieve the setup as described in the introduction, you'll need to set these two configuration options accordingly:

```yaml
host: social.example.org
account-domain: example.org
```

!!! info
    The `host` must always be the DNS name that your GoToSocial instance runs on. It does not affect the IP address the GoToSocial instance binds to. That is controlled with `bind-address`.

## Reverse proxy

When using a [reverse proxy](../getting_started/reverse_proxy/index.md) you'll need to ensure you're set up to handle traffic on both of those domains. You'll need to redirect a few endpoints from the account domain to the host domain.

Redirects are typically used so that the change of domain can be detected client side. The endpoints to redirect from the account domain to the host domain are:

* `/.well-known/webfinger`
* `/.well-known/host-meta`
* `/.well-known/nodeinfo`

!!! tip
    Do not proxy or redirect requests to the API endpoints, `/api/...`, from the account domain to the host domain. This will confuse heuristics some clients use to detect a split-domain deployment resulting in broken login flows and other weird behaviour.

### nginx

In order to configure the redirect, you'll need to configure it on the account domain. Assuming the account domain is `example.org` and the host domain is `social.example.org`, the following configuration snippet showcases how to do this:

```nginx
server {
  server_name example.org;                                                      # account-domain

  location /.well-known/webfinger {
    rewrite ^.*$ https://social.example.org/.well-known/webfinger permanent;    # host
  }

  location /.well-known/host-meta {
      rewrite ^.*$ https://social.example.org/.well-known/host-meta permanent;  # host
  }

  location /.well-known/nodeinfo {
      rewrite ^.*$ https://social.example.org/.well-known/nodeinfo permanent;   # host
  }
}
```

### Traefik

If `example.org` is running on [Traefik](https://doc.traefik.io/traefik/), we could use labels similar to the following to setup the redirect.

```yaml
myservice:
  image: foo
  # Other stuff
  labels:
    - 'traefik.http.routers.myservice.rule=Host(`example.org`)'                                                                # account-domain
    - 'traefik.http.middlewares.myservice-gts.redirectregex.permanent=true'
    - 'traefik.http.middlewares.myservice-gts.redirectregex.regex=^https://(.*)/.well-known/(webfinger|nodeinfo|host-meta)(\?.*)?$'  # host
    - 'traefik.http.middlewares.myservice-gts.redirectregex.replacement=https://social.$${1}/.well-known/$${2}$${3}'                # host
    - 'traefik.http.routers.myservice.middlewares=myservice-gts@docker'
```

### Caddy 2

Ensure that the redirect is configured on the account domain in your `Caddyfile`. The following example assumes the account domain as `example.com`, and host domain as `social.example.com`.

```
example.com {                                                                    # account-domain
        redir /.well-known/host-meta* https://social.example.com{uri} permanent  # host
        redir /.well-known/webfinger* https://social.example.com{uri} permanent  # host
        redir /.well-known/nodeinfo* https://social.example.com{uri} permanent   # host
}
```


