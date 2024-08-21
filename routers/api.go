package routers

import (
	"net/http"
	"time"

	"sirherobrine23.com.br/go-bds/bds/modules"
	"sirherobrine23.com.br/go-bds/bds/modules/versions"
	"sirherobrine23.com.br/go-bds/bds/routers/utils"
)

var APIv1 *http.ServeMux = http.NewServeMux()

func init() {
	versions.InitFetch()
	APIv1.HandleFunc("GET /fetchs", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, map[string]any{
			"bedrock": map[string]string{
				"mojang": versions.NextFetchBedrock.Format(time.RFC3339),
			},
			"java": map[string]string{
				"mojang": versions.NextFetchJava.Format(time.RFC3339),
			},
		})
	})

	APIv1.HandleFunc("GET /servers", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, map[string]any{
			"bedrock": map[string]any{
				"mojang": versions.BedrockVersions,
			},
			"java": map[string]any{
				"mojang": versions.JavaVersions,
				"purpur": versions.PurpurVersions,
			},
		})
	})

	APIv1.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, map[string]any{
			"version": modules.AppVersion,
			"started": modules.StartTime.Format(time.RFC1123Z),
			"uptime":  time.Now().UTC().Sub(modules.StartTime).String(),
		})
	})
}
