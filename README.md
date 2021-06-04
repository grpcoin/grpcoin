# gRPCoin: Programmatic paper trading platform using gRPC and Protobuf

This is an educational project that allows you to trade cryptocurrencies
competitively by writing a bot on top of a [gRPC](https://grpc.io)-based API
with real-time prices.

Using the [gRPCoin API][api], you can **write a bot** and **start trading**
using 100,000$ cash assigned to your account when you sign up.

[api]: ./api/grpcoin.proto

## Get started

1. Create a [GitHub personal access token](https://github.com/settings/tokens).
   You don't need to give any permissions.

1. Learn about API types and trading/portfolio endpoints in
   [grpcoin.proto][api].

1. Choose a programming language, and compile the `.proto` file ([learn
   how](https://grpc.io/docs/languages/)) or you can fork an
   [example bot implementation](#example-bot-implementations).

1. Implement a bot! To authenticate to the API, you need to add an
   `authorization` header (called "metadata" in gRPC) to your requests, e.g.

       authorization: Bearer MY_TOKEN

1. Endpoints you can use:
    - API PROD endpoint `api.grpco.in:443` (TLS)
      - Rate limits:
        - **100 per minute** for authenticated calls.
        - **50 per minute** for unauthenticated calls.
    - Website: https://grpco.in/

1. First time you make an authenticated request, your account will be created
   with $100,000 cash to start buying coins.

1. Start tracking your progress on the [leaderboard][home].

[home]: https://grpco.in/

## Example bot implementations

To learn gRPC and Protocol Buffers, you should learn the [API][api],
generate the code from the proto, and write your own bot.

Just to get started, you can also check out some example bot implementations:

1. [Go example bot](./example-bot/)
1. [C# example bot](https://github.com/grpcoin/example-bot-csharp)
1. [Node.js command-line tool](https://github.com/grpcoin/example-cli-node)

## Development

To run the server locally, follow these steps:

1. Clone the repository and `cd` into it.
1. Install Google Cloud SDK (`gcloud` tool) https://cloud.google.com/sdk/docs/quickstart
1. Install Firestore Emulator to run database locally (this
   may require you to install Java Runtime Environment).

      ```sh
      gcloud components install cloud-firestore-emulator
      ```

1. Make sure the emulator works by running it once:

      ```sh
      gcloud beta emulators firestore start
      ```

1. From repository root, run:

      ```sh
      LISTEN_ADDR=localhost PORT=8080 go run ./apiserver
      ```
      to start the gRPC API server, or run

      ```sh
      LISTEN_ADDR=localhost PORT=8080 go run ./frontend
      ```
      to start the web frontend and navigate to http://localhost:8080 to explore.
