package main

import (
	"bufio"
	"bytes"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/go-playground/validator.v9"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &formValidator{validator: validator.New()}

	e.POST("/resize", resize)

	e.Logger.Fatal(e.Start(":1323"))
}

type (
	param struct {
		URL    string `form:"url" validate:"required"`
		Width  int    `form:"width" validate:"required"`
		Height int    `form:"height" validate:"required"`
	}
	formValidator struct {
		validator *validator.Validate
	}
)

func (cv *formValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func resize(c echo.Context) error {
	p := new(param)
	if err := c.Bind(p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := c.Validate(p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	resp, err := http.Get(p.URL)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer resp.Body.Close()

	image, _, err := image.Decode(resp.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	newImage := imaging.Resize(image, p.Width, p.Height, imaging.Lanczos)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err = jpeg.Encode(w, newImage, nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Stream(http.StatusOK, "image/jpeg", bytes.NewReader(b.Bytes()))
}
