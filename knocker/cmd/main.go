package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/saneknapenek/clog"
)

func main() {

	log := clog.InitLogger(clog.EnvDev, nil)

	proxyUrl := "http://localhost:1080"

	knocking(*log, proxyUrl)
}

func knocking(log slog.Logger, strProxyUrl string) {

	proxyUrl, err := url.Parse(strProxyUrl)
	if err != nil {
		log.Error(
			"error parse proxy url",
			slog.String("url:", proxyUrl.String()),
			slog.String("error:", err.Error()),
		)
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Transport: transport,
	}

	for {
		response, err := client.Get("http://ifconfig.me")
		if err != nil {
			log.Error(
				"error requesting to destination host",
				slog.String("error:", err.Error()),
			)
		} else {
			bodyBuf, err := io.ReadAll(response.Body)
			if err != nil {
				log.Error(
					"error reading response body",
					slog.String("error", err.Error()),
				)
			}
			log.Info(
				"success requesting to destination host",
				slog.Int("status", response.StatusCode),
				slog.String("body", string(bodyBuf)),
			)
		}

		time.Sleep(5 * time.Second)
	}
}
