package response

// Error модель ответа в случае ошибки
// type Error struct {
// 	Error string `json:"error"`
// }

// OK модель ответа в случае успеха
// type OK struct {
// 	Result any `json:"result"`
// }

const (
	StatusOK  = "OK"
	StatusErr = "Error"
)

// Response модель ответа сервера
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Result any    `json:"result,omitempty"`
}

func Error(msg string) Response {
	return Response{
		Status: StatusErr,
		Error:  msg,
	}
}

func OK(result any) Response {
	return Response{
		Status: StatusOK,
		Result: result,
	}
}
