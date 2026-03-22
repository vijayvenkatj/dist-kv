package api

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type putRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type deleteRequest struct {
	Key string `json:"key"`
}
