# gRPCoin: Programmatic paper trading platform using gRPC and Protobuf

This is an educational server project that allows you to trade
cryptocurrencies in a competitive setting using a gRPC API with
real-time prices.

Using the [gRPCoin API][api], you can **write a bot** and **start trading**
using 100,000$ cash assigned to your account when you sign up.

[api]: ./api/grpcoin.proto

## Get started

1. Create a [GitHub personal access token](https://github.com/settings/tokens).
   You don't need to give any permissions.

1. Learn about API types and endpoints in [grpcoin.proto][api].

1. Choose a programming language, and compile the `.proto` file ([learn
   how](https://grpc.io/docs/languages/)) or you can fork an
   [example bot implementation](#example-bot-implementations).

1. Implement a bot! To authenticate to the API, you need to add an
   `authorization` header (called "metadata" in gRPC) to your requests, e.g.

       authorization: Bearer MY_TOKEN

1. Endpoints you can use:
    - API PROD endpoint `grpcoin-main-kafjc7sboa-wl.a.run.app:443` (TLS)
    - Website: https://grpcoin-main-kafjc7sboa-wl.a.run.app/

1. First time you make an authenticated request, your account will be created
   with $100,000 cash to start buying coins.

1. Start tracking your progress on the [leaderboard].

[leaderboard]: https://grpcoin-main-kafjc7sboa-wl.a.run.app/

## Example bot implementations

To learn gRPC and Protocol Buffers, you should learn the [API][api],
generate the code from the proto, and write your own bot.

Just to get started, you can also check out some example bot implementations:

1. [Go example bot](./example-bot/)
1. [C# example bot](https://github.com/grpcoin/example-bot-csharp)
1. Please contribute more! 🚀
