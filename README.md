# gRPCOIN: Programmatic paper trading platform using gRPC and Protobuf

This is an educational project that allows you to trade cryptocurrencies
competitively by writing a bot on top of a [gRPC](https://grpc.io)-based API
with real-time prices.

Using the [gRPCOIN API][api], you can **write a bot** and **start trading**
using 100,000$ cash assigned to your account when you sign up.

## How to play on gRPCOIN?

Visit the https://grpco.in/ website to learn more about how to
[join the game](https://grpco.in/join).

[api]: ./api/grpcoin.proto

## What is on this repository?

This repository contains the server implementations and the [API][api] for the
platform.

## How to implement a bot?

To learn gRPC and Protocol Buffers, you should learn the [API][api],
generate the code from the proto file, and write your own bot on top.

Just to get started, you can also check out some example bot implementations:

1. [Go example bot](./example-bot/)
1. [C# example bot](https://github.com/grpcoin/example-bot-csharp)
1. [Node.js command-line tool](https://github.com/grpcoin/example-cli-node)
1. [Python example bot](https://github.com/grpcoin/example-grpcoin-bot-python)

## Develop gRPCOIN API/Frontend Servers

To run the servers locally, follow these steps:

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
