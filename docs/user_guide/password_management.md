# Password Management

GoToSocial stores hashes of user passwords in its database using the secure [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) function in the [Go standard libraries](https://pkg.go.dev/golang.org/x/crypto/bcrypt).

This means that the plaintext value of your password is safe even if the database of your GoToSocial instance is compromised. It also means that your instance admin does not have access to your password.

To check whether a password is sufficiently secure before accepting it, GoToSocial uses [this library](https://github.com/wagslane/go-password-validator) with entropy set to 60. This means that passwords like `password` are rejected, but something like `verylongandsecurepasswordhahaha` would be accepted, even without special characters/upper+lowercase etc.

We recommend following the EFF's guidelines on [creating strong passwords](https://ssd.eff.org/en/module/creating-strong-passwords).

## Change Your Password

### API method

If you are logged in (ie., you have a valid oauth token), you can change your password by making a POST request to `/api/v1/user/password_change`, using your token as authentication, and giving your old password and desired new password as parameters. Check the [API documentation](../api/swagger.md) for more details.

## Reset Your Password

todo
