{{ template "header.tpl" (printf "User: %s" .u.DisplayName) }}

{{ $tv := pv .u.Portfolio .quotes }}
<div class="container">
    <div class="row py-5">
        <div class="col-md-5 col-lg-3">
            <div class="card mb-3">
                {{ with (profilePic .u.ProfileURL) }}
                <img src="{{.}}" class="card-image-top img-thumbnail
                    d-none d-md-block" />
                {{ end }}
                <div class="card-body">
                    <h3 class="card-title display-5">
                        <a href="{{.u.ProfileURL}}" class="card-link
                        stretched-link">
                            {{.u.DisplayName}}
                        </a>
                    </h3>
                    <p class="card-text">
                        Joined {{ fmtDuration (since .u.CreatedAt) 1 }} ago.
                    </p>
                </div>
            </div>

            <div class="card">
                <h5 class="card-header">
                    Positions
                </h5>
                <ul class="list-group list-group-flush">
                    <li class="list-group-item
                        justify-content-between d-flex">
                        <div>
                            <b>CASH</b>
                        </div>
                        <div class="text-end">
                            {{fmtPrice .u.Portfolio.CashUSD -}}<br />
                            <span class="text-muted">
                                {{ fmtPercent (toPercent ( div .u.Portfolio.CashUSD $tv ) ) }}
                            </span>
                        </div>
                    </li>
                    {{ range $tick, $amount := .u.Portfolio.Positions }}
                    <li class="list-group-item
                        justify-content-between d-flex">
                        <div>
                            <b>{{$tick}}</b><br />
                            <span class="text-muted">
                                {{ fmtPrice (index $.quotes $tick ) }}
                            </span>
                        </div>
                        <div class="text-end">
                            {{ fmtPrice (mul $amount (index $.quotes $tick )) }}<br />
                            <span class="text-muted">
                                {{ fmtPercent ( toPercent (div (mul $amount (index $.quotes $tick )) $tv )) }}
                            </span>
                        </div>
                    </li>
                    {{ end }}
                </ul>
                <div class="card-footer justify-content-between d-flex">
                    <div>
                        <b>Total value</b>
                    </div>
                    <div class="text-end">
                        {{fmtPrice $tv }}
                    </div>
                </div>
            </div>

            <div class="mt-3">
                <a href="/" class="btn btn-lg text-primary">
                    &larr; Leaderboard
                </a>
            </div>
        </div>
        <div class="col-md-7 col-lg-9 order-md-last">
            <div class=card>
                <h4 class="card-header">
                    <span>Returns</span>
                </h4>
                <div class="card-body text-center p-0">
                    <table class="table table-striped mb-0">
                        <thead>
                            <tr>
                                {{ range .returns }}
                                <th scope="col">{{.Label}}</th>
                                {{ end }}
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                {{ range .returns }}
                                <td class="{{ if isNegative .Percent }}table-danger{{ else }}table-success
                                    {{end}}">
                                    <b>
                                        {{ fmtPercent .Percent }}
                                    </b>
                                </td>
                                {{ end }}
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            {{ with .orders }}
            <div class="card mt-3">
                <h4 class="card-header">
                    <span>Orders</span>
                </h4>
                <div class="card-body overflow-scroll">
                    <table class="table table-striped table-hover">
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
                                <td>{{fmtPriceFull (mul .Price .Size)}}</td>
                                <td>{{fmtDuration (since .Date) 2}}</td>
                            </tr>
                            {{ end }}
                        </tbody>
                    </table>
                </div>
            </div>
            {{ end }}
        </div>
    </div>

    {{ with .orders }}
    <h2>Order History</h2>


    {{ end }}

    <hr />
    <p>
        <a href="/">&larr; Back to leaderboard</a>
    </p>
    {{ template "footer.tpl" }}
