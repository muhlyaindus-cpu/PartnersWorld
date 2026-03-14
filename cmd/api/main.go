package main

import (
	"log"
	"net/http"
	"os"
	"partnersale/internal/auth"
	"partnersale/internal/db"
	"partnersale/internal/partnership"
	"partnersale/internal/user"
)

func main() {
	// Подключаем БД
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./partnersale.db"
	}
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "./web"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database := db.InitDB(dbPath)
	defer database.Close()
	db.RunMigrations(database)

	// Репозитории и хендлеры
	userRepo := user.NewRepository(database)
	userHandler := user.NewHandler(userRepo)

	partnershipRepo := partnership.NewRepository(database)
	partnershipHandler := partnership.NewHandler(partnershipRepo)

	// Публичные эндпоинты (без токена)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/api/register", userHandler.Register)
	http.HandleFunc("/api/login", userHandler.Login)
	http.HandleFunc("/api/users/count", userHandler.GetUserCount) // Новый эндпоинт для счётчика
	http.HandleFunc("/api/user/profile", auth.Middleware(userHandler.GetMyProfile))
	http.HandleFunc("/api/user/profile/update", auth.Middleware(userHandler.UpdateProfile)) // PUT
	http.HandleFunc("/api/user/", auth.Middleware(userHandler.GetPublicProfile))            // GET /api/user/123

	// Защищённые эндпоинты (с JWT)
	http.HandleFunc("/api/requests", auth.Middleware(partnershipHandler.ListRequests))         // GET для всех авторизованных
	http.HandleFunc("/api/requests/create", auth.Middleware(partnershipHandler.CreateRequest)) // POST
	// Внимание: следующие два эндпоинта используют нестандартный синтаксис с методом и параметрами.
	// Для их корректной работы потребуется роутер вроде chi или ручное извлечение параметров.
	http.HandleFunc("POST /api/requests/{id}/respond", auth.Middleware(partnershipHandler.RespondToRequest))
	http.HandleFunc("GET /api/requests/{id}/responses", auth.Middleware(partnershipHandler.ListResponses))
	// Раздача статических файлов (фронтенд)
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)
	log.Printf("DB: %s", dbPath)
	log.Printf("WEB: %s", webDir)
	log.Printf("Сервер запущен на http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
