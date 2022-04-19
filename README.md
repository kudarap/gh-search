# GH search
Github username search

## Problem
We need to create service that provides user details that depends on external resource and can scale. External resource comes with limitation and sometimes unrealiable.

Since we are dealing with external resource, its important to know how we can work with its limitation, handling inevitable failure and timeouts. Luckily, Github's API rate limits are clear https://docs.github.com/en/rest/reference/rate-limit and with scalability in mind, we will use caching like redis to share between service instances for horizontal scaling.

## Running Locally

#### Requirements
- Go 1.18
- Redis 6.2
- Docker

#### Running Locally
- generate github access token if you don't have one on https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
- copy `.env.sample` file to `.env` and change the values accordingly, you must use your own access token
- spin up redis using docker `docker run --rm -p 6379:6379 -e REDIS_PASSWORD=password bitnami/redis:6.2` and dont forget to change `REDIS_PASWORD`.
- finally `go run ./cmd/serverd`

#### Sample Search Request
- `curl http://localhost:8080/users?usernames=kudarap,octocat`

#### Running Unit Test
- `go test -v -cover -race ./...`


## Architecture
- concurrent user data retrieval
- internal rate limit check - prefetched rate limits on init
- request grouping to prevent duplicate in-flight requests
- caching user data using redis to share between service instances when scaling


## Future Plan
- integration test
- stress test
- server level rate limit
- tracing and metrics
- automatic deployment
- more unit test


## Proud Code
https://bitbucket.org/javinc/tasio