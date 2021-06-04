{{ template "header.tpl" (printf "User: %s" .u.DisplayName) }}

{{ $tv := pv .u.Portfolio .quotes }}
<div class="container">
    <div class="row py-5">
        <div class="col-md-5 col-lg-3">
            <div class="card bg-color-black mb-3">
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
                    <p class="card-text text-white">
                        Joined {{ fmtDuration (since .u.CreatedAt) 1 }} ago.
                    </p>
                </div>
            </div>

            <div class="card bg-color-black">
                <h5 class="card-header text-white">
                    Positions
                </h5>
                <ul class="list-group list-group-flush">
                    <li class="list-group-item
                        justify-content-between d-flex bg-color-black bg-hover">
                        <div>
                            <b class="text-white">CASH</b>
                        </div>
                        <div class="text-end text-white">
                            {{fmtPrice .u.Portfolio.CashUSD -}}<br />
                            <span class="text-white-50">
                                {{ fmtPercent (toPercent ( div .u.Portfolio.CashUSD $tv ) ) }}
                            </span>
                        </div>
                    </li>
                    {{ range $tick, $amount := .u.Portfolio.Positions }}
                    <li class="list-group-item
                        justify-content-between d-flex bg-color-black bg-hover">
                        <div>
                            <b class="text-white">{{$tick}}</b><br />
                            <span class="text-muted">
                                {{ if not (isZero $amount) }}
                                x{{ fmtAmount $amount }}
                                @ {{ fmtPrice (index $.quotes $tick ) }} </span>
                            {{end}}
                        </div>
                        <div class="text-end text-white">
                            {{ fmtPrice (mul $amount (index $.quotes $tick )) }}<br />
                            <span class="text-white-50">
                                {{ fmtPercent ( toPercent (div (mul $amount (index $.quotes $tick )) $tv )) }}
                            </span>
                        </div>
                    </li>
                    {{ end }}
                </ul>
                <div class="card-footer justify-content-between d-flex bg-hover">
                    <div>
                        <b class="text-white">Total value</b>
                    </div>
                    <div class="text-end text-white">
                        {{fmtPrice $tv }}
                    </div>
                </div>
            </div>

            <div class="mt-3">
                <a type="button" href="/" class="btn bg-color-black bg-hover btn-lg text-white" style="width: 100%;">
                    &larr; Leaderboard
                </a>

            </div>
        </div>
        <div class="col-md-7 col-lg-9 order-md-last">
            <div class="card bg-color-black">
                <h4 class="card-header">
                    <span class="text-white">Returns</span>
                </h4>
                <div class="card-body text-center p-0">
                    <table class="table table-striped mb-0">
                        <thead>
                            <tr>
                                {{ range .returns }}
                                <th scope="col" class="text-white returns-tbl">{{.Label}}</th>
                                {{ end }}
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                {{ range .returns }}
                                <td class="{{ if isNegative .Percent }}gh-bg-color-red{{ else }}gh-bg-color-green{{end}} returns-tbl">
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
            <div class="card mt-3 bg-color-black">
                <h4 class="card-header">
                    <span class="text-white">Orders</span>
                </h4>
                <div class="card-body overflow-scroll" style="height: 500px;">
                    <table class="table table-striped table-hover text-white">
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
    <hr />

    {{ template "footer.tpl" }}