# Access Control

GoToSocial uses access control restrictions to protect users and resources from unwanted interactions with remote accounts and instances.

As shown in the [http signatures](http_signatures.md) document, GoToSocial requires all incoming `GET` and `POST` requests from remote servers to be signed. Unsigned requests will be denied with http code `401 Unauthorized`.

Access control restrictions are implemented by checking the `keyId` of the signature (who owns the public/private key pair making the request).

First, the host value of the `keyId` uri is checked against the GoToSocial instance's list of blocked (defederated) domains. If the host is recognized as a blocked domain, then the http request will immediately be aborted with http code `403 Forbidden`.

Next, GoToSocial will check for the existence of a block (in either direction) between the owner of the public key making the http request, and the owner of the resource that the request is targeting. If the GoToSocial user blocks the remote account making the request, then the request will be aborted with http code `403 Forbidden`.
