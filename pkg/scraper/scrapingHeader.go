package scraper

import "math/rand"

// httpHeader es el tipo de dato que contieee las cabeceras de las peticiones http de los sitios web.
type httpHeaders map[string]string

// headers lista de httpHeader
type headers []httpHeaders

var headersPool = headers{
	{
		"DNT":             "1",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "es-US,es-419;q=0.9,es;q=0.8,en;q=0.7",
		"Cache-Control":   "max-age=0",
		"Connection":      "keep-alive",
	},
}

func GetHeaders() httpHeaders {
	return headersPool[rand.Intn(len(headersPool))]
}
