# GoLeague

___

This is a project with main purpose of learning Golang, gRPC, Redis, Docker and other technologies.

The application will be a complete Back-End for a web app with League Of Legends Data.

At first, only the API and Fetcher will be developed, but later a web app or mobile app can be developed.

---

## Services

- #### Fetcher
  - Queue for fetching data at a constant rate from the Riot API and filling the database, handling multiple constraints and cases to guarantee data integrity. 
  - Logging to files that will be sent to a bucket.
  - gRPC endpoint for getting data needed on demand (For example, a player match list and it's data)
  - Shared Rate Limit between the On Demand and the Queue, with priority for the On Demand requests, creating a optimized use of the rate limits.
- #### API
  - Receives requests from a FrontEnd and get the data from the Database or the Fetcher.
  - gRPC client for force fetching requests on the Fetcher.
  - Multiple endpoints for data fetching.
  - InMemory cache for data that doesn't change often (Tierlists) 
- #### Scheduler
  - Simple scheduler that runs on background, revalidating cache keys and recalculating matches average ratings.
- #### Redis
  - Shared cache for holding champion and item assets.
  - Cache layer for matches previews and tierlist data.
- #### PostgreSQL
  - Shared database.
