package nape

import(
	"encoding/json"
	"encoding/xml"

	"github.com/muir/nvelope"
)

// DecodeJSON is is a pre-defined special nject.Provider created with 
// nvelope.GenerateDecoder for decoding JSON requests.  Use it with the
// other features of https://github.com/muir/nvelope
var DecodeJSON = nvelope.GenerateDecoder(
	nvelope.WithDecoder("application/json", json.Unmarshal),
	nvelope.WithDefaultContentType("application/json"),
	nvelope.WithPathVarsFunction(func(r *http.Request) nvelope.RouteVarLookup {
		vars := mux.Vars(r)
		return func(v string) string {
			return vars[v]
		}
	}),
)

// DecodeXML is is a pre-defined special nject.Provider created with 
// nvelope.GenerateDecoder for decoding XML requests.Use it with the
// other features of https://github.com/muir/nvelope
var DecodeXML = nvelope.GenerateDecoder(
	nvelope.WithDecoder("application/xml", xml.Unmarshal),
	nvelope.WithDefaultContentType("application/xml"),
	nvelope.WithPathVarsFunction(func(r *http.Request) nvelope.RouteVarLookup {
		vars := mux.Vars(r)
		return func(v string) string {
			return vars[v]
		}
	}),
)
