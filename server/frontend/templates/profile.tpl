{{ template "header.tpl" (printf "User: %s" .u.DisplayName) }}

    <a href="{{.u.ProfileURL}}" rel="nofollow">
        <h1>{{.u.DisplayName}}</h1>
    </a>

    <small>
        Joined on: {{fmtDate .u.CreatedAt}} ({{ fmtDuration (since .u.CreatedAt) 2 }} ago)
    </small>

<h2>Positions</h2>

<table>
    <thead>
        <tr>
            <th>Ticker</th>
            <th>Amount</th>
            <th>Price</th>
            <th>Value</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>CASH</td>
            <td>{{fmtAmount .u.Portfolio.CashUSD -}}</td>
            <td>$1.00</td>
            <td>${{fmtAmount .u.Portfolio.CashUSD -}}</td>
        </tr>
        {{ range $tick, $amount := .u.Portfolio.Positions }}
        <tr>
            <td>{{$tick}}</td>
            <td>{{fmtAmount $amount -}}</td>
            <td>{{ fmtPrice (index $.quotes $tick ) }} </td>
            <td>{{ fmtPrice (mul $amount (index $.quotes $tick )) }} </td>
        </tr>
        {{ end }}
         <tr>
            <td></td>
            <td></td>
            <td>TOTAL:</td>
            <td>{{ fmtPrice (pv .u.Portfolio .quotes ) }} </td>
        </tr>
    </tbody>
</table>

{{ with .orders }}
<h2>Order History</h2>

<table>
    <thead>
        <tr>
            <th>Action</th>
            <th>Ticker</th>
            <th>Size</th>
            <th>Price</th>
            <th>Total Cost</th>
            <th>Date</th>
        </tr>
    </thead>
    <tbody>
        {{ range . }}
        <tr>
            <td>{{.Action}}</td>
            <td>{{.Ticker}}</td>
            <td>{{fmtAmount .Size}}</td>
            <td>{{fmtPrice .Price}}</td>
            <td>{{fmtPrice (mul .Price .Size)}}</td>
            <td>{{fmtDuration (since .Date) 2}}</td>
        </tr>
        {{ end }}
    </tbody>
</table>
{{ end }}

<hr/>
<p>
    <a href="/">&larr; Back to leaderboard</a>
</p>
{{ template "footer.tpl" }}
