package main

import (
	"log"
	"time"

	"github.com/pdcgo/pdc_orc/pkg/golang_proj"
	"github.com/pdcgo/pdc_orc/pkg/node_proj"
)

func main() {
	frontend := node_proj.NewNodeProject(
		&node_proj.GitConfig{
			Uri: "https://github.com/PDC-Repository/warehois.git",
		},
		"/tmp/warehouse_frontend",
	)
	backend := golang_proj.
		NewGolangProject(
			&golang_proj.GitConfig{
				Uri:    "https://github.com/pdcgo/gudang.git",
				Branch: "dev",
			},
			"/pdcgo/gudang",
		).
		Check().
		SetEnvs(golang_proj.MapEnvs{
			"DB_DRIVER":       "sqlite",
			"DB_SQLITE_FNAME": "/tmp/dev_db.sqlite",
			"WAREHOUSE_HOST":  "localhost",
			"DEV_MODE":        "true",
		})

	client := frontend.
		Check().
		Pulling().
		Folder("client").
		InstallPackage().
		DotEnv(map[string]string{
			"VITE_BASE_URL":            "http://localhost:8080",
			"VITE_RECAPTCHA_TOKEN":     "",
			"VITE_CLIENT_URL":          "http://localhost:5173",
			"VITE_ADMIN_WAREHOUSE_URL": "http://localhost:5174",
		})

	admin := frontend.
		Check().
		Pulling().
		Folder("admin").
		InstallPackage().
		DotEnv(map[string]string{
			"VITE_BASE_URL":  "http://localhost:8080",
			"VITE_IMAGE_CDN": "http://localhost:8080/v1/assets/get?id=",
		})

	if frontend.Err != nil {
		log.Panicln(frontend.Err)
	}

	go client.PackageRunScript("dev", "--port", "5173")
	go admin.PackageRunScript("dev", "--port", "5174")
	backend.Run("go", "run", ".\\cmd\\app_server", "dev")
	// backend.Run(".\\shell\\run_devel.bat")

	for {
		time.Sleep(time.Minute)
	}
}
