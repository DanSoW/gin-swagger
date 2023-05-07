package main

import (
	"context"
	"fmt"
	mainserver "main-server"
	"main-server/config"
	initConfigure "main-server/config"
	handler "main-server/pkg/handler"
	repository "main-server/pkg/repository"
	"main-server/pkg/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Основной сервис
// @version 1.0
// description Основной сервис

// @host localhost:5000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {

	// Инициализация конфигурации сервера
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	// Инициализация переменных внешней среды
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variable: %s", err.Error())
	}

	// Инициализация логгера
	openLogFiles, err := initConfigure.InitLogrus()
	if err != nil {
		logrus.Error("Ошибка при настройке параметров логгера. Вывод всех ошибок будет осуществлён в консоль")
	}

	// Закрытие всех открытых файлов в результате настройки логгера
	defer func() {
		for _, item := range openLogFiles {
			item.Close()
		}
	}()

	// Создание нового подключения к БД
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})

	// Создание строки DNS
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		viper.GetString("db.host"),
		viper.GetString("db.username"),
		os.Getenv("DB_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.port"),
		viper.GetString("db.sslmode"),
	)

	// Получение адаптера после открытия подключения к базе данных через gorm
	dbAdapter, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dns,
	}), &gorm.Config{})

	// Создание нового адаптера c кастомной таблицей
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(dbAdapter, &config.AcRule{}, viper.GetString("rules_table_name"))

	if err != nil {
		logrus.Fatalf("failed to initialize adapter by db with custom table: %s", err.Error())
	}

	// Определение нового объекта enforcer, по модели PERM
	enforcer, err := casbin.NewEnforcer(viper.GetString("paths.perm_model"), adapter)

	if err != nil {
		logrus.Fatalf("failed to initialize new enforcer: %s", err.Error())
	}

	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	// Инициализация OAuth2 сервисов
	config.InitOAuth2Config()
	config.InitVKAuthConfig()

	// Dependency Injection
	repos := repository.NewRepository(db, enforcer)
	service := service.NewService(repos)
	handlers := handler.NewHandler(service)

	srv := new(mainserver.Server)

	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("Main Server Started")

	// Реализация Graceful Shutdown
	// Блокировка функции main с помощью канала os.Signal
	quit := make(chan os.Signal, 1)

	// Запись в канал, если процесс, в котором выполняется приложение
	// получит сигнал SIGTERM или SIGINT
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Чтение из канала, блокирующая выполнение функции main
	<-quit

	logrus.Print("Rental Housing Main Server Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
}

/* Инициализация файлов конфигурации */
func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
