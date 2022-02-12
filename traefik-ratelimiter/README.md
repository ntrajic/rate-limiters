# Abhinav Singh 

    https://medium.com/geekculture/system-design-basics-rate-limiter-351c09a57d14
    Oct 30, 2021

    e.g. open source traefik api gateway implements as midleware package in go token bucket algorithm:
    traefik/pkg/middlewares/ratelimiter (Ref 0.)

## System Design Basics: Rate Limiter

    What is Rate Limiter?
    Rate limiting refers to preventing the frequency of an operation from exceeding a defined limit. In large-scale systems, rate limiting is commonly used to protect underlying services and resources. Rate limiting is generally used as a defensive mechanism in distributed systems, so that shared resources can maintain availability.
    Rate limiting protects your APIs from unintended or malicious overuse by limiting the number of requests that can reach your API in a given period of time. Without rate limiting any user can bombard your server with requests leading to spikes that starves other users.

## Rate limiting at work

    Why Rate limiting?
    Preventing Resource Starvation: The most common reason for rate limiting is to improve the availability of API-based services by avoiding resource starvation. Load based denial of service (doS) attacks can be prevented if rate limiting is applied. Other users are not starved even when one user bombards the API with loads of requests.

    ##Security: Rate limiting prevents brute forcing of security intensive functionalities like login, promo code etc. Number of requests to these features is limited on a user level so brute force algorithms don’t work in these scenarios.

    ##Preventing Operational Costs: In case of auto-scaling resources on a pay per use model, Rate Limiting helps in controlling operational costs by putting a virtual cap on scaling of resources. Resources might scale out of proportion leading to exponential bills if rate limiting is not employed.

# Rate Limiting Strategies

    Rate limiting can be applied on the following parameters:

    1. User: A limit is applied on the number of requests allowed for a user in a given period of time. User based rate limiting is one of the most common & intuitive forms of rate limiting.

    2. Concurrency: Here the limit is employed on the number of parallel sessions that can be allowed for a user in a given timeframe. A limit on the number of parallel connections helps mitigate DDOS attacks as well.
    3. Location/ID: This helps in running location based or demography centric campaigns. Requests not from the target demography can be rate limited so as to increase availability in the target regions
    4. Server: Server based rate limiting is a niche strategy. This is employed generally when specific servers need most of the requests, i.e. servers are strongly coupled to specific functions

# Rate Limiting Algorithms

    1. Leaky Bucket: Leaky Bucket is a simple intuitive algorithm. It creates a queue with a finite capacity. All requests in a given time frame beyond the capacity of the queue are spilled off.
    The advantage of this algorithm is that it smoothens out bursts of requests and processes them at a constant rate. It’s also easy to implement on a load balancer and is memory efficient for each user. A constant near uniform f
    low is maintained to the server irrespective of the number of requests.

    Leaky Bucket
    The downside of this algorithm is that a burst of requests can fill up the bucket leading to starving of new requests. It also provides no guarantee that requests get completed in a given amount of time.

    2. Token Bucket: Token Bucket is similar to leaky bucket. Here we assign tokens on a user level. For a given time duration d, the number of request r packets that a user can receive is defined. Every time a new request arrives at a server, there are two operations that happen:
    Fetch token: The current number of tokens for that user is fetched. If it is greater than the limit defined then the request is dropped.
    Update token: If the fetched token is less than the limit for the time duration d, then the request is accepted and the token is appended.
    This algorithm is memory efficient as we are saving less amount of data per user for our application. The problem here is that it can cause race condition in a distributed environment. This happens when there are two requests from two different application servers trying to fetch the token at the same time.

    Token Bucket Algorithm
    3. Fixed Window Counter: Fixed window is one of the most basic rate limiting mechanisms. We keep a counter for a given duration of time, and keep incrementing it for every request we get. Once the limit is reached, we drop all further requests till the time duration is reset.
    The advantage here is that it ensures that most recent requests are served without being starved by old requests. However, a single burst of traffic right at the edge of the limit might hoard all the available slots for both the current and next time slot. Consumers might bombard the server at the edge in an attempt to maximise number of requests served.

    Fixed Window Counter
    4. Sliding Log : Sliding log algorithm involves maintaining a time stamped log of requests at the user level. The system keeps these requests time sorted in a Set or a Table. It discards all requests with timestamps beyond a threshold. Every minute we look out for older requests and filter them out. Then we calculate the sum of logs to determine the request rate. If the request would exceed the threshold rate, then it is held, else it is served.
    The advantage of this algorithm is that it does not suffer from the boundary conditions of fixed windows. Enforcement of the rate limit will remain precise. Since the system tracks the sliding log for each consumer, you don’t have the stampede effect that challenges fixed windows.
    However, it can be costly to store an unlimited number of logs for every request. It’s also expensive to compute because each request requires calculating a summation over the consumer’s prior requests, potentially across a cluster of servers. As a result, it does not scale well to handle large bursts of traffic or denial of service attacks.

    5. Sliding Window: This is similar to the Sliding Log algorithm, but memory efficient. It combines the fixed window algorithm’s low processing cost and the sliding log’s improved boundary conditions.
    We keep a list/table of time sorted entries, with each entries being a hybrid and containing the timestamp and the number of requests at that point. We keep a sliding window of our time duration and only service requests in our window for the given rate. If the sum of counters is more than the given rate of the limiter, then we take only the first sum of entries equal to the rate limit.
    The Sliding Window approach is the best of the lot because it gives the flexibility to scale rate limiting with good performance. The rate windows are an intuitive way to present rate limit data to API consumers. It also avoids the starvation problem of the leaky bucket and the bursting problems of fixed window implementations

# Rate Limiting in Distributed Systems

    The above algorithms works very well for single server applications. But the problem becomes very complicated when there is a distributed system involved with multiple nodes or app servers.It becomes more complicated if there are multiple rate limited services distributed across different server regions. The two broad problems that comes across in these situations are Inconsistency and Race Conditions.

    1. Inconsistency
    In case of complex systems with multiple app servers distributed across different regions and having their own rate limiters, we need to define a global rate limiter.
    A consumer could surpass the global rate limiter individually if it receives a lot of requests in a small time frame. The greater the number of nodes, the more likely the user will exceed the global limit.
    There are two ways to solve for these problems:

    2. Sticky Session: Have a sticky session in your load balancers so that each consumer gets sent to exactly one node. The downsides include lack of fault tolerance & scaling problems when nodes get overloaded. You can read more about sticky sessions here

    3. Centralized Data Store: Use a centralized data store like Redis or Cassandra to handle counts for each window and consumer. The added latency is a problem, but the flexibility provided makes it an elegant solution.

# Race Conditions

    Race conditions happen in a get-then-set approach with high concurrency. Each request gets the value of counter then tries to increment it. But by the time that write operation is completed, several other requests have read the value of the counter(which is not correct). Thus a very large number of requests are sent than what was intended. This can be mitigated using locks on the read-write operation, thus making it atomic. But this comes at a performance cost as it becomes a bottleneck causing more latency.

# Throttling

    Throttling is the process of controlling the usage of the APIs by customers during a given period. Throttling can be defined at the application level and/or API level. When a throttle limit is crossed, the server returns HTTP status “429 — Too many requests”.

    Types of Throttling:
    ##Hard Throttling: The number of API requests cannot exceed the throttle limit.
    ##Soft Throttling: In this type, we can set the API request limit to exceed a certain percentage. For example, if we have rate-limit of 100 messages a minute and 10% exceed-limit, our rate limiter will allow up to 110 messages per minute.
    ###Elastic or Dynamic Throttling: Under Elastic throttling, the number of requests can go beyond the threshold if the system has some resources available. For example, if a user is allowed only 100 messages a minute, we can let the user send more than 100 messages a minute when there are free resources available in the system.
    Congratulations on making it to the end! Feel free to talk tech or any cool projects on Twitter, Github, Medium, LinkedIn or Instagram.

References
0. Traefik Labs: Traefik Gateway's Rate Limiter https://github.com/traefik/traefik/blob/master/pkg/middlewares/ratelimiter/rate_limiter.go 
1. Sticky sessions: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/sticky-sessions.html
2. Narendra L: https://www.youtube.com/watch?v=mhUQe4BKZXs
3. Google Developers: https://cloud.google.com/architecture/rate-limiting-strategies-techniques
4. Kong API Gateway: https://konghq.com/blog/
5. how-to-design-a-scalable-rate-limiting-algorithm/
Token Bucket: https://en.wikipedia.org/wiki/Token_bucket