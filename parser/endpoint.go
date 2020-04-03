package parser

import "net/http"

type Parser func([]byte) ([]*http.Request, error)
