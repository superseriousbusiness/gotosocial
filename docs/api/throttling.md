# HTTP Request Throttling

To prevent your instance from being accidentally DDOS'd (aka [the hug of death](https://en.wikipedia.org/wiki/Slashdot_effect)) when interacting with an account with thousands of followers, the ActivityPub API of GoToSocial uses request throttling to limit the number of open connections to your instance.
