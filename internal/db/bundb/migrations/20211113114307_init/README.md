A note on when we should set data structures linked to objects in the database to use the
bun `nullzero` tag -- this should only be done if the member type is a pointer, or if the
this primitive type is literally invalid with an empty value (e.g. media IDs which when
empty signifies a null database value, compared to say an account note which when empty
could mean either an empty note OR null database value).

Obviously it is a little more complex than this in practice, but keep it in mind!