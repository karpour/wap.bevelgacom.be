package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/bevelgacom/wap.wap.bevelgacom.be/pkg/barcode"
	"github.com/labstack/echo/v4"
)

type barcodeContent struct {
	Type    string
	Content string
}

func serveBarcode(c echo.Context) error {
	p := c.Request().URL.Path

	_, file := path.Split(p)
	if file == "" || file == "barcode" {
		file = "index.wml"
	}

	f, err := os.Open("./static/barcode/" + file)

	if err == os.ErrNotExist {
		return c.String(http.StatusNotFound, "")
	} else if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	mime := "text/vnd.wap.wml"
	if strings.HasSuffix(file, ".wbmp") {
		mime = "image/vnd.wap.wbmp"
	}

	return c.Stream(http.StatusOK, mime, f)
}

func serveBarcodePage(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("./static/barcode/barcode.wml"))
	content := base64.StdEncoding.EncodeToString([]byte(c.QueryParam("content")))

	pageContent := barcodeContent{
		Type:    c.QueryParam("type"),
		Content: content,
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")

	return tmpl.Execute(c.Response().Writer, pageContent)
}

func serveBarcodeImage(c echo.Context) error {
	content, err := base64.StdEncoding.DecodeString(c.QueryParam("content"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid content")
	}

	if c.QueryParam("type") == "qr" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateQR(string(content)))
	} else if c.QueryParam("type") == "aztec" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateAztec(string(content)))
	} else if c.QueryParam("type") == "code128" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateCode128(string(content)))
	}

	return nil
}
