package httpServer

import (
	"OnlineExams/src/core/appValues"
	"OnlineExams/src/core/utils/logging"
	"OnlineExams/src/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func RunServer() error {
	err := LoadDatabase()
	if err != nil {
		logging.Error("RunMasterServer: failed to load database: ", err)
		return err
	}

	appValues.ServerEngine = fiber.New(fiber.Config{
		ProxyHeader:   ProxyHeader,
		CaseSensitive: CaseSensitive,
	})

	LoadMiddlewares(appValues.ServerEngine)
	LoadHandlersV1(appValues.ServerEngine)

	return nil
}

func LoadHandlersV1(app *fiber.App) {
}

func LoadMiddlewares(app *fiber.App) {
	app.Use(cors.New())

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Server", "Microsoft-IIS/10.0")
		c.Set("X-Powered-By", "PHP/8.2.8")

		return c.Next()
	})
}

func LoadDatabase() error {
	return database.StartDatabase()
}
