package main

import (
	"fmt"
	"stream_hub/internal/admin"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/config"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	adminConf, err := config.NewAdminConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	auth := security.NewAuth(commonConf)
	
	userController := admin.NewUserController(base)
	if err := userController.InitRootAdmin(); err != nil {
		fmt.Println("init root admin failed:", err)
		return
	}

	router := admin.NewAdminRouter(base, auth, adminConf)

	fmt.Printf("Admin service starting on port %d...\n", adminConf.Port)
	if err := router.Run(); err != nil {
		fmt.Println("err:", err)
	}
}
