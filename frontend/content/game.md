See [this page](/join) to learn how to create an account and start coding a
bot and use the API to play the game.

### Game mechanics

1. All players start with $100,000 cash (USDT) to buy coins.

1. All trades are "market orders" and execute with the real-time prices at the
   time of receiving the order.

   * There are no "limit/stop orders", this is by design.
   * If you want limit/stop orders, write a bot running continuously and
     watching the prices.
   * We offer an API to track prices of supported coins in real-time (or you
     can use other APIs to find coin prices).

1. You can't trade coins with other coins: You need to sell coin position to
   get cash, and then buy another coin with available cash balance.

1. Supported coin symbols (e.g., BTC, DOGE, ETH) are documented [in the API][api].

### API Rate limits

* Authenticated API calls (e.g. `Trade`): 100 calls per minute.
* Unauthenticated API calls (e.g. `Watch`): 50 calls per minute.

Rate limits reset at the beginning of each minute.

It's not recommended to make trades concurrently. To protect against data
inconsistency, all trades are serialized and executed one by one. This means
if you issue `Trade()` requests in parallel, some will fail.

### Game Resets

We are probably going to reset the game (accounts won't be deleted) periodically
so that new users have a chance. We still have not decided when and how
frequently this will happen. Stay tuned (**and don't be frustrated when the game
resets**).

[api]: https://github.com/grpcoin/grpcoin/blob/main/api/grpcoin.proto
