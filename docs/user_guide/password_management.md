# Password Management

## Change Your Password

You can use the [User Settings Panel](./settings.md) to change your password. Just log in to the user panel, scroll to the bottom of the page, and input your old password and desired new password.

If the new password you provide is not long/complicated enough, you will see an error and be prompted to try again with a different password.

If your instance uses OIDC (ie., you log in via Google or some other external provider), you will have to change your password via your OIDC provider, not through the user settings panel.

## Password Storage

GoToSocial stores hashes of user passwords in its database using the secure [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) function in the [Go standard libraries](https://pkg.go.dev/golang.org/x/crypto/bcrypt).

This means that the plaintext value of your password is safe even if the database of your GoToSocial instance is compromised. It also means that your instance admin does not have access to your password.

To check whether a password is sufficiently secure before accepting it, GoToSocial uses [this library](https://github.com/wagslane/go-password-validator) with entropy set to 60. This means that passwords like `password` are rejected, but something like `verylongandsecurepasswordhahaha` would be accepted, even without special characters/upper+lowercase etc.

We recommend following the EFF's guidelines on [creating strong passwords](https://ssd.eff.org/en/module/creating-strong-passwords).
