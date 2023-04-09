package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	go func() {
		app := iris.Default()

		// However, this one will match /user/john/ and also /user/john/send
		// If no other routers match /user/john, it will redirect to /user/john/
		app.Get("/", func(ctx iris.Context) {

			ctx.WriteString("Helo")
		})

		app.Listen(":8080")
	}()
	// inisialisasi objek Echo
	url1, err := url.Parse("http://localhost:8080")
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	// url2, err := url.Parse("http://localhost:8082")
	// if err != nil {
	// 	fmt.Printf("err: %v\n", err)
	// }
	e := echo.New()
	e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: url1,
		},
	})))

	// handler untuk proxy request
	e.Any("/*", func(c echo.Context) error {
		// parsing alamat URL dari request
		targetURL := c.Scheme() + "://" + c.Request().Host + c.Request().RequestURI
		fmt.Printf("targetURL: %v\n", targetURL)
		// membuat objek http.Request dengan alamat URL dari request
		req, err := http.NewRequest(c.Request().Method, targetURL, c.Request().Body)
		if err != nil {
			return err
		}

		// mengatur header dari request
		for k, v := range c.Request().Header {
			req.Header.Add(k, strings.Join(v, ","))
		}

		// inisialisasi objek url.URL untuk merepresentasikan alamat URL dari server proxy
		proxyURL, _ := url.Parse("http://localhost:8080")

		// membuat objek http.Transport dengan konfigurasi custom
		transport := &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		// membuat objek http.Client dengan menggunakan objek http.Transport yang sudah diatur konfigurasinya
		client := &http.Client{Transport: transport}

		// melakukan request ke server tujuan dengan menggunakan objek http.Client yang sudah diatur konfigurasinya
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return err
		}
		defer resp.Body.Close()

		// menyalin header dan isi response dari server tujuan ke response server proxy
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Response().Header().Add(k, v)
			}
		}
		c.Response().WriteHeader(resp.StatusCode)
		io.Copy(c.Response(), resp.Body)

		return nil
	})

	// menjalankan server
	e.Logger.Fatal(e.StartTLS(":84", "/etc/letsencrypt/live/aerogpstrack.com/cert.pem", "/etc/letsencrypt/live/aerogpstrack.com/privkey.pem"))
}
