{{ template "header.tpl" "Leaderboard" }}

<main class="container">
    <div class="pt-5 text-center text-white">
        <h2>gRPCOIN</h2>
        <p class="lead">Leaderboard</p>
    </div>
    <div class="card mx-auto bg-color-black col-12 col-lg-6 p-0">
        <div class="card-body p-1 m-0">
            <table class="table table-borderless table-hover text-white m-0 leaderboard">
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
                        <td>
                        </td>
                        <td>
                            {{ with (profilePic .User.ProfileURL) }}
                            <img src="{{.}}" width=24 height=auto />
                            {{ end }}
                            <a href="/user/{{.User.ID}}">
                                {{.User.DisplayName}}</a>
                        </td>
                        <td>
                            USD {{fmtPrice .Valuation}}
                        </td>

                        {{ end }}
                </tbody>
            </table>
        </div>
    </div>

</main>
{{ template "footer.tpl" }}