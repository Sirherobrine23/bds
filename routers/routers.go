package routers

import (
	"fmt"
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/modules/config"
	web "sirherobrine23.com.br/go-bds/bds/web_src"
)

var Router *http.ServeMux = http.NewServeMux() // Server Handler

type ServerConfig struct {
	Port         int    `ini:"PORT" json:"port"`
	PortRedirect int    `ini:"PORT_REDIRECT" json:"portRedirect"`
	ListenHTTPs  bool   `ini:"HTTPS" json:"listenHTTPs"`
	CertFile     string `ini:"CERT" json:"-"`
	KeyFile      string `ini:"KEY" json:"-"`
}

func Listen() error {
	server, err := config.ConfigProvider.GetSection("server")
	if err != nil {
		if server, err = config.ConfigProvider.NewSection("server"); err != nil {
			return err
		}
		server.NewKey("PORT", "3000")
		server.NewKey("HTTPS", "false")
	}
	server.Key("PORT").MustInt(3000)
	var dataConfig ServerConfig
	if err = server.MapTo(&dataConfig); err != nil {
		return err
	}

	listAddr, redirectAddr := fmt.Sprintf(":%d", dataConfig.Port), fmt.Sprintf(":%d", dataConfig.PortRedirect)
	fmt.Printf("Listen on %s\n", listAddr)
	if dataConfig.ListenHTTPs {
		// Redirect handler
		redirect := http.NewServeMux()
		redirect.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			head := w.Header()
			head.Set("Location", "https://"+r.Host+r.RequestURI)
			w.WriteHeader(http.StatusMovedPermanently)
		})

		fmt.Printf("Listen redirect on %s\n", listAddr)
		go http.ListenAndServe(redirectAddr, redirect)
		return http.ListenAndServeTLS(listAddr, dataConfig.CertFile, dataConfig.KeyFile, Router)
	}

	return http.ListenAndServe(listAddr, Router)
}

func init() {
	// API Path
	Router.Handle("/api/v1/", http.StripPrefix("/api/v1", APIv1))

	// Static files
	Router.Handle("GET /js/", http.FileServer(http.FS(web.StatisFiles)))
	Router.Handle("GET /css/", http.FileServer(http.FS(web.StatisFiles)))
	Router.Handle("/", WebRoute)
}
