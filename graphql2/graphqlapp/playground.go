package graphqlapp

import "html/template"

const playVersion = "1.7.23"

const playHTML = `
<!DOCTYPE html>
<html>
<head>
	<meta charset=utf-8/>
	<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
	<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react@{{ .Version }}/build/static/css/index.css"/>
	<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react@{{ .Version }}/build/favicon.png"/>
	<script src="//cdn.jsdelivr.net/npm/graphql-playground-react@{{ .Version }}/build/static/js/middleware.js"></script>
	<title>GoAlert - GraphQL API</title>
	<style type="text/css">
		html { font-family: "Open Sans", sans-serif; overflow: hidden; }
		body { margin: 0; background: #172a3a; }
		.CodeMirror-cursor { background-color: white !important; width: 1px !important; }
	</style>
</head>
<body>

<div id="root"/>
<script type="text/javascript">
	window.addEventListener('load', function (event) {
		var root = document.getElementById('root');
		var path = location.host + location.pathname.replace(/\/explore.*$/, '');

		GraphQLPlayground.init(root, {
			endpoint: location.protocol + '//' + path,
			settings: {
				'request.credentials': 'same-origin'
			}
		})
	})
</script>
</body>
</html>
`

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(playHTML))
