# GoLeague

___

This is a project which the main purpouse is learn Golang, gRPC and Redis.
It consists of a fetcher for the RiotAPI.

---

## Services

- #### Fetcher
  - gRPC server that revalidates a given cache key. (For example, if a champion id is not set)
  - Queue for fetching data at a constant rate from the Riot API and filling the database.
  - gRPC endpoint for getting data needed on demand (For example, a player match list and it's data)
  - Shared Rate Limit between the On Demand and the Queue, with priority for the On Demand requests.
- #### Revalidator
  - Revalidate every asset key, should be run sometimes to  assure that it get's the most recent cache for the champion or item.
- #### API
  - TODO
  - Receives requests from a FrontEnd and get the data from the Database or the Fetcher.
- #### Redis
  - Shared cache for holding champion and item assets.
- #### PostgreSQL
  - Shared database
