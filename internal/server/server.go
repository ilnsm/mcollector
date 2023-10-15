package server

import (
	"fmt"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/storage"
	"log"
	"net/http"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>

	<title>Metric's' Data</title>

</head>
<body>

	   <h1>Data</h1>
	   <ul>
	   {{range $key, $value := .}}
	       <li>{{ $key }}: {{ $value }}</li>
	   {{end}}
	   </ul>


</body>
</html>
`

func Run(s storage.Storager) error {

	cfg, err := config.New()
	if err != nil {
		log.Fatal("could not get config")
	}

	fmt.Println("Start server on", cfg.Endpoint)
	return http.ListenAndServe(cfg.Endpoint, MetrRouter(s))
}
