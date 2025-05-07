# Scraper Deterrence

GoToSocial provides an optional proof-of-work based scraper and automated HTTP client deterrence that can be enabled on profile and status web views. The way
it works is that it generates a unique but deterministic challenge for each incoming HTTP request based on client information and current time, that-is a hex encoded SHA256 hash, and asks the client to find an addition to a portion of this that will generate a hex encoded SHA256 hash with a pre-determined number of leading '0' characters. This is served to the client as a minimal holding page with a single JavaScript worker that computes a solution to this.

The number of required leading '0' characters can be configured to your liking, where higher values take longer to solve, and lower values take less. But this is not exact, as the challenges themselves are random, so you can only effect the **average amount of time** it may take. If your challenges take too long to solve, you may deter users from accessing your web pages. And conversely, the longer it takes for a solution to be found, the more you'll be incurring costs for scrapers (and in some cases, causing their operation to time-out). That balance is up to you to configure, hence why this is an advanced feature.

Once a solution to this challenge has been provided, by refreshing the page with the solution in the query parameter, GoToSocial will verify this solution and on success will return the expected profile / status page with a cookie that provides challenge-less access to the instance for up-to the next hour.

The outcomes of this, (when enabled), is that it should make scraping of your instance's profile / status pages economically unviable for automated data gathering (e.g. by AI companies, search engines). The only negative, is that it places a requirement on JavaScript being enabled for people to access your profile / status web views.

This was heavily inspired by the great project that is [anubis], but ultimately we determined we could implement it ourselves with only the features we require, minimal code, and more granularity with our existing authorization / authentication procedures.

The GoToSocial implementation of this scraper deterrence is still incredibly minimal, so if you're looking for more features or fine-grained control over your deterrence measures then by all means keep ours disabled and stand-up a service like [anubis] in front of your instance!

!!! warning
    This proof-of-work scraper deterrence does not protect user profile RSS feeds due to the extra complexity involved. If you rely on your RSS feed being exposed, this is one such case where [anubis] may be a better fit!

[anubis]: https://github.com/TecharoHQ/anubis