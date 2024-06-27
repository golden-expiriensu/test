package main

import (
	"arkis_test/controllers"
	"arkis_test/stream"
	"context"
)

func main() {
	ctx := context.Background()
	stream.New("A", ctx)
	stream.New("B", ctx)
	controllers.StartGinServer()
}
