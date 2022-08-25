# simple_throttle implementing token-bucket  rate-limitting

    The most common algorithm for implementing rate-limiting behavior is the token bucket, which uses the analogy of a bucket that can hold some maximum number of tokens. When a function is called, a token is taken from the bucket, which then refills at some fixed rate.

    The way that a Throttle treats requests when there are insufficient tokens in the bucket to pay for it can vary depending according to the needs of the developer. Some common strategies are:

    Return an error
    This is the most basic strategy and is common when you’re only trying to restrict unreasonable or potentially abusive numbers of client requests. A RESTful service adopting this strategy might respond with a status 429 (Too Many Requests).

    Replay the response of the last successful function call
    This strategy can be useful when a service or expensive function call is likely to provide an identical result if called too soon. It’s commonly used in the JavaScript world.

    Enqueue the request for execution when sufficient tokens are available
    This approach can be useful when you want to eventually handle all requests, but it’s also more complex and may require care to be taken to ensure that memory isn’t exhausted.

# simple_throttle algorithm

    // ==========================================
    // simple_throttle rate-limitter algorithm:
    // ==========================================
    // - Throttle function wraps the effector function e with a closure 
    //   that contains the rate-limiting logic. 
    // - The bucket is initially allocated max tokens; 
    //   each time the closure is triggered it checks whether it has any remaining tokens. 
    // - If tokens are available, it decrements the token count by one and 
    //   triggers the effector function. 
    // - If not, an error is returned. 
    // - Tokens are added at a rate of refill tokens every duration d.