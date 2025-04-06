# GoLeague

___

This is a project with main purpose of learning Golang, gRPC, Redis, Docker and other technologies.

The application will be a complete Back-End for a web app with League Of Legends Data.

At first, only the API and Fetcher will be developed, but later a web app or mobile app can be developed.

---

## Services

- #### Fetcher
  - gRPC server that revalidates a given cache key. (For example, if a champion id is not set)
  - Queue for fetching data at a constant rate from the Riot API and filling the database, handling multiple constraints and cases to guarantee data integrity. 
  - Logging to files that will be sent to buckets.
  - gRPC endpoint for getting data needed on demand (For example, a player match list and it's data)
  - Shared Rate Limit between the On Demand and the Queue, with priority for the On Demand requests, creating a optimized use of the rate limits.
- #### Revalidator
  - Revalidate every asset key, should be run sometimes to  assure that it got the most recent cache for the champion or item.
- #### API
  - TODO
  - Receives requests from a FrontEnd and get the data from the Database or the Fetcher.
- #### Redis
  - Shared cache for holding champion and item assets.
- #### PostgreSQL
  - Shared database
