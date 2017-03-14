package htmlGen

const DefaultHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <title>DST API Explorer</title>
  <style>
  	div.row { margin-bottom: 0.75em; padding-top: .25em; padding-bottom: .25em; }
  	div.row:nth-child(odd) { background-color: whitesmoke; }
  	div.row:nth-child(even) { background-color: white; }

  	.unavailable_status { color: gray; }
  	.unavailable_desc_status::after { content: " (unavailable)"; }
  	.active_status { color: blue; }
  	.deprecated_status { color: gray; }
  	.deprecated_desc_status::after { content: " (deprecated)"; }
  	.name_col { display: inline-block; width: 7em; }
  	.version_col { display: inline-block; width: 5em; }
  	.desc_col { display: inline-block; width: 60em; }
  	.apiEndpoint { display: block; margin-left: 4em; }
  	.apiSet { font-size: 110%; font-weight: bold; }
  </style>
</head>
<body>
<h1>API Explorer</h1>
<div>
{{range .APIList}}
<div class="row">
	<div class="apiSet">
		<div class="name_col">{{.Name}}</div>
		<div class="desc_col">{{.Desc}}</div>
	</div>
	{{range .Endpoints}}
	<div class="apiEndpoint {{.Status}}_status">
		<div class="name_col">{{if not .SwaggerURL}}{{.Name}}{{else}}<a href="{{.SwaggerURL}}">{{.Name}}</a>{{end}}</div>
		<div class="version_col">{{.Version}}</div>
		<div class="desc_col {{ .Status }}_desc_status">{{ .Desc }}</div>
	</div>
	{{end}}
</div>
{{end}}
</div>
</body>
</html>`
