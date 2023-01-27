# HTTP Signatures

GoToSocial requires all `GET` and `POST` requests to ActivityPub s2s endpoints to be accompanied by a valid http signature.

GoToSocial will also sign all outgoing `GET` and `POST` requests that it makes to other servers.

This behavior is the equivalent of Mastodon's [AUTHORIZED_FETCH / "secure mode"](https://docs.joinmastodon.org/admin/config/#authorized_fetch).

GoToSocial uses the [go-fed/httpsig](https://github.com/go-fed/httpsig) library for signing outgoing requests, and for parsing and validating the signatures of incoming requests. This library strictly follows the [Cavage http signature RFC](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures), which is the same RFC used by other implementations like Mastodon, Pixelfed, Akkoma/Pleroma, etc. (This RFC has since been superceded by the [httpbis http signature RFC](https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures), but this is not yet widely implemented.)

## Incoming Requests

GoToSocial request signature validation is implemented in [internal/federation](https://github.com/superseriousbusiness/gotosocial/blob/main/internal/federation/authenticate.go).

GoToSocial will attempt to parse the signature using the following algorithms (in order), stopping at the first success:

```text
RSA_SHA256
RSA_SHA512
ED25519
```

## Outgoing Requests

GoToSocial request signing is implemented in [internal/transport](https://github.com/superseriousbusiness/gotosocial/blob/main/internal/transport/signing.go).

When assembling signatures:

- outgoing `GET` requests use `(request-target) host date`
- outgoing `POST` requests use `(request-target) host date digest` 

GoToSocial uses the `RSA_SHA256` algorithm for signing requests, which is in line with other ActivityPub implementations.

## Quirks

The `keyId` used by GoToSocial in the `Signature` header will look something like the following:

```text
https://example.org/users/example_user/main-key
```

This is different from most other implementations, which usually use a fragment (`#`) in the `keyId` uri. For example, on Mastodon the user's key would instead be found at:

```text
https://example.org/users/example_user#main-key
```

For Mastodon, the public key of a user is served as part of that user's Actor representation. GoToSocial mimics this behavior when serving the public key of a user, but instead of returning the entire Actor at the `main-key` endpoint (which may contain sensitive fields), will return only a partial stub of the actor. This looks like the following:

```json
{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams"
  ],
  "id": "https://example.org/users/example_user",
  "preferredUsername": "example_user",
  "publicKey": {
    "id": "https://example.org/users/example_user/main-key",
    "owner": "https://example.org/users/example_user",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzGB3yDvMl+8p+ViutVRG\nVDl9FO7ZURYXnwB3TedSfG13jyskoiMDNvsbLoUQM9ajZPB0zxJPZUlB/W3BWHRC\nNFQglE5DkB30GjTClNZoOrx64vLRT5wAEwIOjklKVNk9GJi1hFFxrgj931WtxyML\nBvo+TdEblBcoru6MKAov8IU4JjQj5KUmjnW12Rox8dj/rfGtdaH8uJ14vLgvlrAb\neQbN5Ghaxh9DGTo1337O9a9qOsir8YQqazl8ahzS2gvYleV+ou09RDhS75q9hdF2\nLI+1IvFEQ2ZO2tLk3umUP1ioa+5CWKsWD0GAXbQu9uunAV0VoExP4+/9WYOuP0ei\nKwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "type": "Person"
}
```

Remote servers federating with GoToSocial should extract the public key from the `publicKey` field. Then, they should use the `owner` field of the public key to further dereference the full version of the Actor, using a signed `GET` request.

This behavior was introduced as a way of avoiding having remote servers make unsigned `GET` requests to the full Actor endpoint. However, this may change in future as it is not compliant and causes issues. Tracked in [this issue](https://github.com/superseriousbusiness/gotosocial/issues/1186).
