package main

import (
	"log"

	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/shotpreviewservicev0_1"
)

func main() {
	handler := handlerv0_1.NewShotPreviewHandler(&service.ShotPreviewServiceImpl{})
	server := shotpreviewservicev0_1.NewServer(handler)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
