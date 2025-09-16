package jsonutil

import "encoding/json"

type Marshaller func(v any) ([]byte, error)

var DefaultMarshaller Marshaller = json.Marshal
