gRPCOIN is an educational paper trading game with real-time cryptocurrency
prices.

It aims to teach you about [Protocol Buffers][pb] and [gRPC] technologies, but
also to actually trade by writing non-trivial high-frequency trading bots.

### How to play?

You have two options to play this game: Either use an example command-line tool
to trade, or write your own bot that uses the trading API.

### How to sign up?

You don't need to create an account as the first time you make an authenticated
API request (using your GitHub token), your account will be initialized, with
$100,000 cash available to trade. So all you need is a GitHub account.

### Option 1: Trade manually with command-line tool

1. Make sure you have a GitHub account.
1. [Create a GitHub personal access token][pat]
    that does not have any permissions!
1.  Install the command line tool [written in Node.js](https://github.com/grpcoin/example-cli-node/)
    (may require `sudo`)

    ```sh
    npm install -g grpcoin
    ```

1. Set your GitHub token to `TOKEN` environment variable, and start trading:

   ```text
   export TOKEN=...
   grpcoin buy ETH 2    # buy 2 ETH
   grpcoin sell ETH 1.5 # sell 1.5 ETH
   grpcoin watch DOGE   # watch DOGE prices
   ```

### Option 2: Fork from an existing bot implementation

There are some very simple bot implementations on GitHub. Fork them and start
writing your trading logic.

* [Go example bot](https://github.com/grpcoin/grpcoin/tree/main/example-bot)
* [C# example bot](https://github.com/grpcoin/example-bot-csharp)
* [Node.js command-line tool](https://github.com/grpcoin/example-cli-node)
* [Python example bot](https://github.com/grpcoin/example-grpcoin-bot-python)


### Option 3: Writing a bot using the API

First, make sure to read the [`grpcoin.proto` API file on GitHub][api].
This file defines the API endpoints and how they work. It's worth noting:

* This is **not a REST API** you can call with `curl` or an HTTP library.
* This API is defined using [Protobuf][pb] and [gRPC][grpc].

**To get started with the API:**

1. You will use Protobuf/gRPC to generate a client library that can call
   the API from the language of your choice.

   * Copy the `grpcoin.proto` to your program

   * Follow instructions on [grpc.io][grpc] to learn how to generate client
     libraries and use gRPC in your language. (Instructions are different for
     each language.)

1. Point your API client to `api.grpco.in:443` (TLS).

1. Authenticate to the API by providing your permissionless [GitHub personal
   access token][pat] by adding `authorization` header (gRPC calls headers
   "metadata"). Prefix the value with string `"Bearer "`, e.g.:

       authorization: Bearer YOUR_PERSONAL_ACCESS_TOKEN

1. Make API requests using the client libraries you generated with gRPC.

1. Enjoy trading!

## Game rules

1. All players start with $100,000 cash.
1. All trades are "market orders" (no limit or stop orders, this is by design), and
   they execute with the real-time prices at the time.
1. You can't trade coins with other coins (e.g., you can buy BTC with cash, or sell
   BTC for cash).
1. Supported coin symbols (e.g., BTC, DOGE, ETH) are documented [in the API][api].

[pb]: https://developers.google.com/protocol-buffers/
[grpc]: https://grpc.io
[api]: https://github.com/grpcoin/grpcoin/blob/main/api/grpcoin.proto
[pat]: https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-token
