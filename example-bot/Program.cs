/*
 * Required NUGET Packages:
 * Google.Protobuf
 * Grpc.NET.ClientFactory
 * Grpc.Tools
 */

using Grpc.Core;
using Grpc.Net.Client;
using System;
using System.Threading.Tasks;

var token = GetToken();
var channel = CreateAuthenticatedChannel(GetUrl(), token); ;
var ticker = new GrpCoin.QuoteTicker { Ticker = "BTC-USD" };

var authClient = new GrpCoin.Account.AccountClient(channel);
var authResponse = await authClient.TestAuthAsync(new GrpCoin.TestAuthRequest());
Console.WriteLine($"You are user {authResponse.UserId}");

var paperTradeClient = new GrpCoin.PaperTrade.PaperTradeClient(channel);
var portfolioResponse = await paperTradeClient.PortfolioAsync(new GrpCoin.PortfolioRequest(), null);
Console.WriteLine($"Cash Position:{portfolioResponse.CashUsd}");
foreach (var position in portfolioResponse.Positions)
{
    Console.WriteLine($"Coin amount: {position.Amount}");
}
var orderReponse = await paperTradeClient.TradeAsync(new GrpCoin.TradeRequest
{
    Action = GrpCoin.TradeAction.Buy,
    Ticker = new GrpCoin.TradeRequest.Types.Ticker { Ticker_ = "BTC" },
    Quantity = new GrpCoin.Amount { Units = 0, Nanos = 99_990_000 }
});
Console.WriteLine($"ORDER EXECUTED: {orderReponse.Action} [{orderReponse.Quantity}] coins at USD[{orderReponse.ExecutedPrice}]");

var tickerClient = new GrpCoin.TickerInfo.TickerInfoClient(channel);
await foreach (var item in tickerClient.Watch(ticker, null).ResponseStream.ReadAllAsync())
{
    Console.Write(item.Price);
    Console.Write("---");
    Console.Write(item.T);
    Console.WriteLine();
}

string GetToken()
{
    var token = Environment.GetEnvironmentVariable("TOKEN");
    if (string.IsNullOrEmpty(token))
    {
        throw new Exception("Create a permissionless Personal Access Token on GitHub and set it to TOKEN environment variable");
    }
    return token;
}
string GetUrl()
{
    const string prod = "https://grpcoin-main-kafjc7sboa-wl.a.run.app:443";
    const string local = "localhost:8080";
    return string.IsNullOrEmpty(Environment.GetEnvironmentVariable("LOCAL")) ? local : prod;
}
GrpcChannel CreateAuthenticatedChannel(string address, string token)
{
    var credentials = CallCredentials.FromInterceptor((context, metadata) =>
    {
        if (!string.IsNullOrEmpty(token))
        {
            metadata.Add("Authorization", $"Bearer {token}");
        }
        return Task.CompletedTask;
    });
    // SslCredentials is used here because this channel is using TLS.
    // CallCredentials can't be used with ChannelCredentials.Insecure on non-TLS channels.
    var channel = GrpcChannel.ForAddress(address, new GrpcChannelOptions
    {
        Credentials = ChannelCredentials.Create(new SslCredentials(), credentials)
    });
    return channel;
}
