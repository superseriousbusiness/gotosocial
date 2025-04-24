# Scraper Deterrence

GoToSocial provides an optional proof-of-work based scraper and automated HTTP client deterrence that can be enabled on profile and status web views. The way
it works is that it generates a unique but deterministic challenge for each incoming HTTP request based on client information and current time, that-is a hex encoded SHA256 hash, and asks the client to find an addition to a portion of this that will generate a hex encoded SHA256 hash with at least 4 leading '0' characters. This is served to the client as a minimal holding page with a single JavaScript worker that computes a solution to this.

Once a solution to this challenge has been provided, by refreshing the page with the solution in the query parameter, GoToSocial will verify this solution and on success will return the expected profile / status page with a cookie that provides challenge-less access to the instance for up-to the next hour.

The outcomes of this, (when enabled), is that it should make scraping of your instance's profile / status pages economically unviable for automated data gathering (e.g. by AI companies, search engines). The only negative, is that it places a requirement on JavaScript being enabled for people to access your profile / status web views.

This was heavily inspired by the great project that is [anubis](https://github.com/TecharoHQ/anubis), but ultimately we determined we could implement it ourselves with only the features we require, minimal code, and more granularity with our existing authorization / authentication procedures.