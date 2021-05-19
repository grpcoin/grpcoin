{{ template "header.tpl" "Leaderboard" }}

<h1>grpcoin Real-time Leaderboard</h1>
<ol>
{{ range $i,$v := .users }}
<li>
	<a href="/user/{{$v.User.ID}}">{{$v.User.DisplayName}}</a> (USD {{fmtPrice $v.Valuation}})
</li>
{{- end }}
</ol>

{{ template "footer.tpl" }}
