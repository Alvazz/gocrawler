package scraper

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gocolly/colly"
	"github.com/hako/durafmt"
)

// httpHeader es el tipo de dato que contieee las cabeceras de las peticiones http de los sitios web.
type httpHeaders map[string]string

// headers lista de httpHeader
type headers []httpHeaders

//env es un map que contiene las variable de ambiente del archivo .env
type enVars map[string]string

// requestTracker es la estructura que tiene los datos de toda la petición que realiza el Scraper
type requestTracker struct {
	id           string
	absoluteURL  string
	callback     string
	requestError string
	request      *colly.Request
	response     *colly.Response
	startAt      time.Time
	endAt        time.Time
	duration     string
}

// scrapingRequests es una lista a punteros de requestTracker
type scrapingRequests []*requestTracker

// newRequestTracker retorna un puntero a una nueva instancia de requestTracker
func newRequestTracker(id, url, cb string, req *colly.Request, res *colly.Response, startAt, endAt time.Time, err error) *requestTracker {
	reqDuration := endAt.Sub(startAt)
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	return &requestTracker{
		id:           id,
		absoluteURL:  url,
		callback:     cb,
		request:      req,
		response:     res,
		startAt:      startAt,
		endAt:        endAt,
		duration:     durafmt.Parse(reqDuration).String(),
		requestError: errorMsg,
	}
}

func (sr scrapingRequests) MarshalJSON() ([]byte, error) {
	newList := make([]interface{}, 0)
	for _, rt := range sr {
		tmpStruct := struct {
			ID           string `json:"id"`
			AbsoluteURL  string `json:"absolute_url"`
			Callback     string `json:"callback"`
			RequestError string `json:"error_msg,omitempty"`
			Request      struct {
				URL                       *url.URL       `json:"url"`
				Headers                   *http.Header   `json:"headers"`
				Ctx                       *colly.Context `json:"context,omitempty"`
				Depth                     int            `json:"depth"`
				Method                    string         `json:"method"`
				ResponseCharacterEncoding string         `json:"response_char_encoding,omitempty"`
				ID                        uint32         `json:"request_id"`
			} `json:"request"`
			Response struct {
				StatusCode int            `json:"status_code"`
				Ctx        *colly.Context `json:"context,omitempty"`
				Headers    *http.Header   `json:"headers"`
			} `json:"response"`
			StartAt  time.Time `json:"start_at"`
			EndAt    time.Time `json:"end_at"`
			Duration string    `json:"request_duration"`
		}{
			ID:           rt.id,
			AbsoluteURL:  rt.absoluteURL,
			Callback:     rt.callback,
			RequestError: rt.requestError,
			Request: struct {
				URL                       *url.URL       `json:"url"`
				Headers                   *http.Header   `json:"headers"`
				Ctx                       *colly.Context `json:"context,omitempty"`
				Depth                     int            `json:"depth"`
				Method                    string         `json:"method"`
				ResponseCharacterEncoding string         `json:"response_char_encoding,omitempty"`
				ID                        uint32         `json:"request_id"`
			}{
				URL:                       rt.request.URL,
				Headers:                   rt.request.Headers,
				Ctx:                       rt.request.Ctx,
				Depth:                     rt.request.Depth,
				Method:                    rt.request.Method,
				ResponseCharacterEncoding: rt.request.ResponseCharacterEncoding,
				ID:                        rt.request.ID,
			},
			Response: struct {
				StatusCode int            `json:"status_code"`
				Ctx        *colly.Context `json:"context,omitempty"`
				Headers    *http.Header   `json:"headers"`
			}{
				StatusCode: rt.response.StatusCode,
				Ctx:        rt.response.Ctx,
				Headers:    rt.response.Headers,
			},
			StartAt:  rt.startAt,
			EndAt:    rt.endAt,
			Duration: rt.duration,
		}
		newList = append(newList, tmpStruct)
	}
	return json.MarshalIndent(newList, "", "\t")
}
