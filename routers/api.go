package routers

import (
	"net/http"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/bds/modules"
	db "sirherobrine23.com.br/go-bds/bds/modules/database"
	"sirherobrine23.com.br/go-bds/bds/modules/versions"
	"sirherobrine23.com.br/go-bds/bds/routers/utils"
	"sirherobrine23.com.br/go-bds/go-bds/bedrock"
	"sirherobrine23.com.br/go-bds/go-bds/java"
)

var (
	APIv1     *http.ServeMux = http.NewServeMux()
	APIv1Auth *http.ServeMux = http.NewServeMux()
)

type ServerVersions struct {
	Times struct {
		Bedrock  time.Time `json:"bedrock"`
		Java     time.Time `json:"java"`
		Spigot   time.Time `json:"spigot"`
		Paper    time.Time `json:"paper"`
		Purpur   time.Time `json:"purpur"`
		Folia    time.Time `json:"folia"`
		Velocity time.Time `json:"velocity"`
	} `json:"cache_time"`
	Bedrock  bedrock.Versions `json:"bedrock"`
	Java     java.Versions    `json:"java"`
	Spigot   java.Versions    `json:"spigot"`
	Paper    java.Versions    `json:"paper"`
	Purpur   java.Versions    `json:"purpur"`
	Folia    java.Versions    `json:"folia"`
	Velocity java.Versions    `json:"velocity"`
}

func init() {
	APIv1.HandleFunc("GET /servers", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, ServerVersions{
			Bedrock:  versions.BedrockVersions,
			Java:     versions.JavaVersions,
			Spigot:   versions.SpigotVersions,
			Paper:    versions.PaperVersions,
			Purpur:   versions.PurpurVersions,
			Folia:    versions.FoliaVersions,
			Velocity: versions.VelocityVersions,
			Times: struct {
				Bedrock  time.Time "json:\"bedrock\""
				Java     time.Time "json:\"java\""
				Spigot   time.Time "json:\"spigot\""
				Paper    time.Time "json:\"paper\""
				Purpur   time.Time "json:\"purpur\""
				Folia    time.Time "json:\"folia\""
				Velocity time.Time "json:\"velocity\""
			}{
				Bedrock:  versions.BedrockTime,
				Java:     versions.JavaTime,
				Spigot:   versions.SpigotTime,
				Paper:    versions.PaperTime,
				Purpur:   versions.PurpurTime,
				Folia:    versions.FoliaTime,
				Velocity: versions.VelocityTime,
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

	APIv1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authStr := r.Header.Get("Authentication")
		switch {
		case strings.HasPrefix(authStr, "token"):
			authStr = strings.TrimSpace(strings.TrimPrefix(authStr, "token"))
			var Token []db.Token
			if err := db.DatabaseConnection.Join("INNER", "User", "User.ID = token.user_id").Where("token.ispass == 0 AND token.token == ?", authStr).Find(&Token); err != nil || len(Token) == 0 {
				utils.JsonResponse(w, http.StatusForbidden, map[string]string{"error": "authentication", "message": "token not exists"})
				return
			}
		case strings.HasPrefix(authStr, "basic"):
			username, password, ok := r.BasicAuth()
			if !ok {
				utils.JsonResponse(w, http.StatusForbidden, map[string]string{"error": "authentication", "message": "Require Authentication header"})
				return
			}

			var Auth []db.Token
			if err := db.DatabaseConnection.Join("INNER", "User", "User.ID = token.user_id").Where("token.ispass == 1 AND User.Username == ?", username).Find(&Auth); err != nil || len(Auth) == 0 {
				utils.JsonResponse(w, http.StatusForbidden, map[string]string{"error": "authentication", "message": "cannot auth user"})
				return
			}

			if Auth[0].Compare(password) != nil {
				utils.JsonResponse(w, http.StatusForbidden, map[string]string{"error": "authentication", "message": "cannot auth user"})
				return
			}
		default:
			utils.JsonResponse(w, http.StatusForbidden, map[string]string{"error": "authentication", "message": "Require Authentication header"})
			return
		}
		APIv1Auth.ServeHTTP(w, r)
	})
}
