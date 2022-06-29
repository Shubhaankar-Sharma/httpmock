package httpmock

import (
	"net"
	"net/http"
)

type Option string

const OverrideHeaderMatchOption Option = "override_header_match"

func (opt *Option) String() string {
	return string(*opt)
}

type MockResponse struct {
	Request  http.Request
	Response Response
}

type MockHTTPServer struct {
	Listener    net.Listener
	ResponseMap map[string]Response
	Opts        []Option
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       string
}

func NewMockHTTPServer(b ...string) *MockHTTPServer {
	var err error
	m := &MockHTTPServer{
		ResponseMap: make(map[string]Response),
		Opts:        make([]Option, 0),
	}

	if len(b) == 0 {
		m.Listener, err = net.Listen("tcp", ":9001")
	} else {
		m.Listener, err = net.Listen("tcp", b[0])
	}
	if len(b) > 1 {
		for _, option := range b[1:] {
			switch Option(option) {
			case OverrideHeaderMatchOption:
				m.Opts = append(m.Opts, OverrideHeaderMatchOption)
			}
		}
	}

	if err != nil {
		panic(err)
	}

	go http.Serve(m.Listener, m)
	return m
}

func (m *MockHTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqString, err := request2string(*req, m.Opts)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("invalid request"))
	} else {
		resp, ok := m.ResponseMap[reqString]
		if ok {
			if resp.Header != nil {
				h := w.Header()
				for k, v := range resp.Header {
					for _, val := range v {
						h.Add(k, val)
					}
				}
			}
			if resp.StatusCode != 0 {
				w.WriteHeader(resp.StatusCode)
			}
			w.Write([]byte(resp.Body))
		} else {
			w.WriteHeader(404)
			w.Write([]byte("route not mocked"))
		}
	}
}
