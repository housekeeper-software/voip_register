package controller

type UserRequest struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

type UserResponse struct {
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	UserId string `json:"userId"`
}
