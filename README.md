# whcrawler

Crawler features
----------------
* Efficient DNS resolver based on https://idea.popcount.org/2013-11-28-how-to-resolve-a-million-domains/
* Load balancing DNS requests to process from concurrent goroutines
* Stepwise cache of DNS responses in GroupCache and Redis cache
