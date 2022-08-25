# throttle_token_bucket for implementing user quotas

    Throttles are also frequently applied on a per-user basis to provide something like a usage quota, so that no one caller can consume too much of a service’s resources.

# USER QUOTA IMPL 

    User Quota Throttle (throttle_token_bucket) is a throttle implementation that, while still using a token bucket algorithm, is otherwise quite different from (simple_throttle) in several ways:

    1 First, instead of having a single bucket that’s used to gate all incoming requests, the following implementation throttles on a per-user basis, returning a function that accepts a “key” parameter, that’s meant to represent a username or some other unique identifier.

    2 Second, rather than attempting to “replay” a cached value when imposing a throttle limit, the returned function returns a Boolean that indicates when a throttle has been imposed. Note that the throttle doesn’t return an error when it’s activated: throttling isn’t an error condition, so we don’t treat it as one.

    3 Finally, and perhaps most interestingly, it doesn’t actually use a timer (a time.Ticker) to explicitly add tokens to buckets on some regular cadence. Rather, it refills buckets on demand, based on the time elapsed between requests. This strategy means that we don’t have to dedicate background processes to filling buckets until they’re actually used, which will scale much more effectively: