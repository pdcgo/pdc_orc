package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/pdcgo/pdc_orc/pkg/golang_proj"
	"github.com/pdcgo/pdc_orc/pkg/node_proj"
)

func main() {
	frontend := node_proj.NewNodeProject(
		&node_proj.GitConfig{
			Uri:    "https://github.com/PDC-Repository/warehois.git",
			Branch: "master",
		},
		"/tmp/warehouse_frontend",
	).
		Check().
		Pulling()
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
			"DB_DRIVER":           "sqlite",
			"DB_SQLITE_FNAME":     "/tmp/dev_db.sqlite",
			"WAREHOUSE_HOST":      "localhost",
			"DEV_MODE":            "true",
			"SDK_FILE":            filepath.Join(frontend.Dir, "client", "src/api/api-sdk.ts"),
			"SDK_FILE_PERMISSION": filepath.Join(frontend.Dir, "client", "src/api/permission.ts"),
			"SDK_TEMPLATE":        filepath.Join(frontend.Dir, "client", "src/api/api-sdk.template"),
		})

	client := frontend.
		Folder("client").
		InstallPackage().
		DotEnv(map[string]string{
			"VITE_BASE_URL":            "http://localhost:8080",
			"VITE_RECAPTCHA_TOKEN":     "",
			"VITE_CLIENT_URL":          "http://localhost:5173",
			"VITE_ADMIN_WAREHOUSE_URL": "http://localhost:5174",
		})

	admin := frontend.
		Folder("client").
		InstallPackage().
		DotEnv(map[string]string{
			"VITE_BASE_URL":            "http://localhost:8080",
			"VITE_RECAPTCHA_TOKEN":     "",
			"VITE_CLIENT_URL":          "http://localhost:5173",
			"VITE_ADMIN_WAREHOUSE_URL": "http://localhost:5174",
		})

	warehouse := frontend.
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
	go warehouse.PackageRunScript("dev", "--port", "5174")
	go admin.PackageRunScript("dev-admin", "--port", "5175")
	backend.Run("go", "run", ".\\cmd\\app_server", "dev")
	// backend.Run(".\\shell\\run_devel.bat")

	for {
		time.Sleep(time.Minute)
	}
}
