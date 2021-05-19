{{ template "header.tpl" "Leaderboard" }}

<h1>grpcoin Real-time Leaderboard</h1>
<ol>
{{ range $i,$v := .users }}
<li>
	{{ with (profilePic $v.User.ProfileURL) }}
		<img src="{{.}}" width=16 height=auto/>
	{{ end }} 
	<a href="/user/{{$v.User.ID}}">{{$v.User.DisplayName}}</a> (USD {{fmtPrice $v.Valuation}})
</li>
{{- end }}
</ol>

{{ template "footer.tpl" }}
