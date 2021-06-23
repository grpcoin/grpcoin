{{ template "header.tpl" (printf "User: %s" .u.DisplayName) }}

{{ $tv := pv .u.Portfolio .quotes }}
<div class="container">
    <div class="row py-5">
        <div class="col-lg-3">
            <div class="card bg-color-black mb-3">
                {{ with (profilePic .u.ProfileURL) }}
                    <img src="{{.}}" class="card-image-top img-thumbnail
                    d-none d-lg-block"/>
                {{ end }}
                <div class="card-body">
                    <h3 class="card-title display-6">
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
                            {{fmtPrice .u.Portfolio.CashUSD -}}<br/>
                            <span class="text-white-50">
                                {{ fmtPercent (toPercent ( div .u.Portfolio.CashUSD $tv ) ) }}
                            </span>
                        </div>
                    </li>
                    {{ range $tick, $amount := .u.Portfolio.Positions }}
                        <li class="list-group-item
                        justify-content-between d-flex bg-color-black bg-hover">
                            <div>
                                <b class="text-white">{{$tick}}</b><br/>
                                {{ if not (isZero $amount) }}
                                    x{{ fmtAmount $amount }}
                                    <small>
                                        at
                                        <span class="text-muted"
                                              id="price-{{$tick}}">{{ fmtPrice (index $.quotes $tick ) }}</span>
                                    </small>
                                {{end}}
                            </div>
                            <div class="text-end text-white">
                                <span id="posValue-{{$tick}}">{{ fmtPrice (mul $amount (index $.quotes $tick )) }}</span><br/>
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
                    <div class="text-end text-white" id="total">
                        {{fmtPrice $tv }}
                    </div>
                </div>
            </div>

            <div class="mt-3">
                <a type="button" href="/" class="btn bg-color-black bg-hover btn-lg text-white
                d-none d-lg-block" style="width: 100%;">
                    &larr; Leaderboard
                </a>
            </div>
        </div>
        <div class="col-lg-9 order-lg-last">
            <div class="card bg-color-black">
                <h4 class="card-header">
                    <span class="text-white">Returns</span>
                </h4>
                <div class="card-body text-center p-0">
                    <div class="table-responsive">
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
                                    <td
                                        class="{{ if isNegative .Percent }}gh-bg-color-red{{ else }}gh-bg-color-green{{end}} returns-tbl">
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
            </div>

            <div class="card bg-color-black mt-3">
                <h4 class="card-header">
                    <span class="text-white">Portfolio Value</span>
                </h4>
                <div class="card-body p-2">
                    <script src="https://cdn.jsdelivr.net/npm/apexcharts"></script>
                    <style>
                        #chart {
                            width: 100%;
                        }

                        #chart-timeline {
                            height: 400px;
                        }
                    </style>
                    <script>
                        var formatUSD = val => Intl.NumberFormat('en-US', {
                            style: 'currency',
                            currency: 'USD'
                        }).format(val)
                        var options = {
                            chart: {
                                id: 'area-datetime',
                                animations: {
                                    enabled: false
                                },
                                type: 'area',
                                fontFamily: null,
                                zoom: {
                                    autoScaleYaxis: true
                                }
                            },
                            series: [],
                            dataLabels: {
                                enabled: false,
                            },
                            markers: {
                                size: 0,
                                colors: '#D4D9DE',
                                hover: {
                                    sizeOffset: 5,
                                }
                            },
                            xaxis: {
                                type: 'datetime',
                                tickAmount: 6,
                                borderColor: '#000',
                                labels: {
                                    datetimeUTC: false,
                                    style: {colors: "#ffffff"},
                                }
                            },
                            yaxis: {
                                labels: {
                                    formatter: formatUSD,
                                    style: {colors: "#ffffff"},
                                },
                            },
                            tooltip: {
                                x: {
                                    fillSeriesColor: false,
                                    format: 'dd MMM yyyy HH:mm',
                                },
                                y: {
                                    formatter: formatUSD,
                                }
                            },
                            fill: {
                                type: 'solid',
                                colors: '#8C949C'
                            },
                            stroke: {
                                curve: 'straight',
                                colors: '#848894'
                            },
                        };


                        document.addEventListener('DOMContentLoaded', async () => {
                            await fetch('/api/portfolioValuation/{{.u.ID}}')
                                .then(resp => {
                                    if (!resp.ok) {
                                        throw new Error(`http status code: ${resp.status}`)
                                    }
                                    return resp.json()
                                })
                                .then(data => {
                                    data.push([new Date().getTime(), {{ fmtAmountRaw $tv }} ]);
                                    options.series = [{
                                        name: 'Portfolio',
                                        data: data,
                                    }];
                                }).catch(e => console.log(e));

                            var tl = document.getElementById("chart-timeline");
                            options.chart.height = tl.offsetHeight;
                            var chart = new ApexCharts(tl, options);
                            chart.render();

                            Date.prototype.subDays = function (days) {
                                var date = new Date(this.valueOf());
                                date.setDate(date.getDate() - days);
                                return date;
                            }
                            var resetButtonStyles = function (activeEl) {
                                document.querySelectorAll('#chart button').forEach(el => el.classList.remove('btn-primary'));
                                document.querySelectorAll('#chart button').forEach(el => el.classList.add('btn-secondary'));
                                activeEl.target.classList.remove('btn-secondary');
                                activeEl.target.classList.add('btn-primary');
                            }
                            document.getElementById('one_month').addEventListener('click', function (e) {
                                resetButtonStyles(e);
                                chart.zoomX(new Date().subDays(31).getTime(), new Date().getTime());
                            })
                            document.getElementById('one_week').addEventListener('click', function (e) {
                                resetButtonStyles(e);
                                chart.zoomX(new Date().subDays(7).getTime(), new Date().getTime());
                            })
                            document.getElementById('one_day').addEventListener('click', function (e) {
                                resetButtonStyles(e);
                                chart.zoomX(new Date().subDays(1).getTime(), new Date().getTime());
                            })
                        });
                    </script>
                    <div id="chart">
                        <div class="toolbar text-end">
                            <button type="button" class="btn btn-sm btn-primary" id="one_month">1 month</button>
                            <button type="button" class="btn btn-sm btn-secondary" id="one_week">1 week</button>
                            <button type="button" class="btn btn-sm btn-secondary" id="one_day">1 day</button>
                        </div>
                        <div id="chart-timeline"></div>
                    </div>
                </div>
            </div>

            {{ with .orders }}
            <div class="card mt-3 bg-color-black">
                <h4 class="card-header">
                    <span class="text-white">Orders</span>
                </h4>
                <div class="card-body table-responsive" style="max-height: 500px;">
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
                                    <td>
                                        <time
                                                datetime="{{fmtDateISO .Date}}"
                                                alt="{{fmtDateISO .Date}}">
                                            {{fmtDuration (since .Date) 2}}
                                        </time>
                                    </td>
                                </tr>
                            {{ end }}
                            </tbody>
                        </table>
                    </div>
                </div>
            {{ end }}
        </div>
    </div>
    <hr/>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/decimal.js/9.0.0/decimal.min.js"></script>
    <script>
        const socket = new WebSocket("ws://localhost:8081/ws/tickers");

        socket.addEventListener("open", function (event) {
            socket.send("Hi grpco.in websocket server!");
        });

        const toDecimal = v => new Decimal(v.Units || v.units || 0).plus(new Decimal(v.Nanos || v.nanos || 0).times(new Decimal(10).pow(-9)))
        const fmtPrice = v => v.toNumber().toLocaleString('en-US',{
            maximumFractionDigits: 2,
            minimumFractionDigits: 2,
        });

        const initialQuotes = {{.quotes}};
        const quotes = {};
        const portfolio = {{.u.Portfolio.Positions }};
        const cash = toDecimal({{.u.Portfolio.CashUSD}});

        socket.onmessage = function (evt) {
            const data = JSON.parse(evt.data)
            const key = data.t
            quotes[key] = toDecimal(data.p)

            for (const [symbol, amount] of Object.entries(portfolio)) {
                if (!quotes[symbol]) {
                    continue
                }
                document.getElementById(`price-${symbol}`).innerHTML = fmtPrice(quotes[symbol])
                const posValue = '$' + fmtPrice(toDecimal(amount).times(quotes[symbol]))
                document.getElementById(`posValue-${symbol}`).innerHTML = posValue;
            }

            const posValues = Object.entries(portfolio).map(pos => {
                const symbol = pos[0]
                const amount = toDecimal(pos[1])
                const price = quotes[symbol] || toDecimal(initialQuotes[symbol])
                const posValue = amount.times(price)
                return posValue
            })
            const total = cash.plus(posValues.reduce((a, b) => a.plus(b)) || 0);
            document.getElementById("total").innerHTML = '$' + fmtPrice(total)
        }
    </script>
    {{ template "footer.tpl" }}
