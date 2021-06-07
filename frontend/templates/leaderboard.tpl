{{ template "header.tpl" "Leaderboard" }}
<main class="container">
    <div class="pt-5 text-center text-white">
        <h2>gRPCOIN</h2>
        <p class="lead">Leaderboard</p>
    </div>
    <div class="card mx-auto bg-color-black col-12 col-lg-6 p-0">
        <div class="card-body p-1 m-0">
            <table
                class="table table-borderless table-hover text-white m-0 leaderboard"
                id="LeaderTable"
                >
                <thead class="card-header">
                    <tr class="text-white">
                        <th scope="col" class="p-3 fs-5">#</th>
                        <th scope="col" class="p-3 fs-5">Name</th>
                        <th scope="col" class="p-3 fs-5">Valuation</th>
                    </tr>
                </thead>
                <tbody>
                    {{ range .users }}
                    <tr class="position-relative">
                        <td></td>
                        <td>
                            {{ with (profilePic .User.ProfileURL) }}
                            <img src="{{.}}" width="24" height="auto" />
                            {{ end }}
                            <a href="/user/{{.User.ID}}" id="{{.User.ID}}">
                            {{.User.DisplayName}}</a
                                >
                        </td>
                        <td id="price-{{.User.ID}}">USD {{fmtPrice .Valuation}}</td>
                        {{ end }}
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</main>
<script>
    const table = document.getElementById("LeaderTable");
    const users = {{.users}}
    const currencies = {};
    const userCash = {};
    
    const socket = new WebSocket("ws://localhost:8081/ws/tickers");
    
    // Connection opened
    socket.addEventListener("open", function(event) {
      socket.send("Hi grpco.in websocket server!");
    });
    
    // Listen for messages
    socket.onmessage = function(evt) {
      const data = JSON.parse(evt.data)
      if (data.product === "BTC-USD" || data.product === "ETH-USD") {
        currencies[data.product === "BTC-USD" ? "BTC" : "ETH"] = data.price.units + data.price.nanos * Math.pow(10, -9);
        for (item of users) {
          const portfolio = item.User.Portfolio;
    
          const cash = portfolio.CashUSD !== undefined ? portfolio.CashUSD.Units + portfolio.CashUSD.Nanos * Math.pow(10, -9) : 0
          const BtcCash = portfolio.Positions.BTC !== undefined ? portfolio.Positions.BTC.Units + portfolio.Positions.BTC.Nanos * Math.pow(10, -9) : 0
          const EthCash = portfolio.Positions.ETH !== undefined ? portfolio.Positions.ETH.Units + portfolio.Positions.ETH.Nanos * Math.pow(10, -9) : 0
    
          const total = cash + (BtcCash > 0 ? BtcCash * currencies["BTC"] : 0) + (EthCash > 0 ? EthCash * currencies["ETH"] : 0)
    
          userCash[item.User.ID] = total;
          document.getElementById(`price-${item.User.ID}`).innerHTML = " USD " + parseFloat(total.toFixed(2)).toLocaleString('en-US', {
            style: 'currency',
            currency: 'USD',
          })
        }
    
        let rows, switching, i, x, y, shouldSwitch;
    
        switching = true;
    
        while (switching) {
          switching = false;
          rows = table.rows;
          for (i = 1; i < (rows.length - 1); i++) {
            shouldSwitch = false;
            x = rows[i].getElementsByTagName("a")[0];
            y = rows[i + 1].getElementsByTagName("a")[0]
            if (userCash[x.id] < userCash[y.id]) {
              shouldSwitch = true;
              break;
            }
          }
          if (shouldSwitch) {
            rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
            switching = true;
          }
        }
      }
    }
    
</script>
{{ template "footer.tpl" }}