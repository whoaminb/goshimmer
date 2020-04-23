package utils

import "github.com/labstack/echo"

const prettyJSON = true

type SimpleResponse struct {
	Error string `json:"err"`
}

func ToJSON(c echo.Context, code int, data interface{}) error {
	if prettyJSON {
		return c.JSONPretty(code, data, " ")
	} else {
		return c.JSON(code, data)
	}
}
