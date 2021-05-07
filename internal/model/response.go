package model

import "github.com/labstack/echo/v4"

type BaseResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	*BaseResponse
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	*BaseResponse
	Errors []string `json:"errors"`
}

func NewResponse(status int, message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		&BaseResponse{status, message}, data,
	}
}

func (r *SuccessResponse) SendJSON(c echo.Context) error {
	return c.JSON(r.Status, r)
}
