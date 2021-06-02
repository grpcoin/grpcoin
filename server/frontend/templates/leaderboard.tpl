{{ template "header.tpl" "Leaderboard" }}

<main class="container">
	<div class="pt-5 text-center">
		<h2>gRPCOIN</h2>
		<p class="lead">Leaderboard</p>
	</div>
	<div class="col-12 col-lg-6 mx-auto">
		<table class="table table-striped table-hover">
			<thead class="table-dark">
				<tr>
					<th scope="col">#</th>
					<th scope="col">Name</th>
					<th scope="col">Valuation</th>
				</tr>
			</thead>
			<tbody>
				{{ range $i,$v := .users }}
				<tr class="position-relative">
					<td>
						{{$i}}
					</td>
					<td>
						{{ with (profilePic $v.User.ProfileURL) }}
						<img src="{{.}}" width=24 height=auto />
						{{ end }}
						<a href="/user/{{$v.User.ID}}">
							{{$v.User.DisplayName}}</a>
					</td>
					<td>
						USD {{fmtPrice $v.Valuation}}
					</td>

				{{ end }}
			</tbody>
		</table>
	</div>
</main>
{{ template "footer.tpl" }}
