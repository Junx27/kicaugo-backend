// Command api menjalankan server KicauGo.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Junx27/kicaugo-backend/cmd/server"
	"github.com/Junx27/kicaugo-backend/internal/config"
	"github.com/Junx27/kicaugo-backend/internal/database"
	"github.com/Junx27/kicaugo-backend/internal/features/auth"
	"github.com/Junx27/kicaugo-backend/internal/features/birdcategory"
	"github.com/Junx27/kicaugo-backend/internal/features/city"
	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/home"
	"github.com/Junx27/kicaugo-backend/internal/features/organizer"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
	"github.com/Junx27/kicaugo-backend/internal/features/province"
	"github.com/Junx27/kicaugo-backend/internal/features/weather"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	// scaffold:imports
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := database.Connect(ctx, &cfg)
	if err != nil {
		log.Fatalf("database connection: %v", err)
	}
	defer pool.Close()

	if err := database.Setup(ctx, pool); err != nil {
		log.Fatalf("database setup: %v", err)
	}

	tokens := jwt.NewManager([]byte(cfg.JWTSecret), cfg.TokenTTL)
	ProvinceService := province.NewService(province.NewRepository(pool))
	CityService := city.NewService(city.NewRepository(pool))
	profileService := profile.NewService(profile.NewRepository(pool))
	eventService := event.NewService(event.NewRepository(pool))
	birdCategoryService := birdcategory.NewService(birdcategory.NewRepository(pool))
	organizerService := organizer.NewService(organizer.NewRepository(pool))
	homeService := home.NewService(profileService, eventService, birdCategoryService, organizerService)
	weatherService := weather.NewService(weather.NewClient(cfg.WeatherAPIKey, cfg.WeatherAPIBase))
	// scaffold:services
	authService := auth.NewService(auth.NewRepository(pool), tokens)

	if err := server.Run(ctx, server.Dependencies{
		Config:   &cfg,
		Tokens:   tokens,
		Province: ProvinceService,
		City:     CityService,
		// scaffold:deps
		Profile:        profileService,
		Event:          eventService,
		Home:           homeService,
		Organizer:      organizerService,
		Weather:        weatherService,
		Auth:           authService,
		BirdCategories: birdCategoryService,
	}); err != nil {
		log.Fatalf("server: %v", err)
	}
}
