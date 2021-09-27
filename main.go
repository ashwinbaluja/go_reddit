package main

import (
	"github.com/azbxx/fiber-reddit-test/database"
	"github.com/azbxx/fiber-reddit-test/login"
	"github.com/azbxx/fiber-reddit-test/posts"
	"github.com/azbxx/fiber-reddit-test/users"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	//jwt "github.com/form3tech-oss/jwt-go"
	//jwtware "github.com/gofiber/jwt/v2"
)

func registerRoutes(app *fiber.App) {
	app.Get("post/:id", posts.GetPost)
	app.Get("user/:id", users.GetUser)
	app.Get("/", posts.GetPostHomepage)

	authGroup := app.Group("/a", login.JWTMiddleware)
	authGroup.Post("/post/create", posts.CreatePost)
	authGroup.Post("/post/:id/upvote", posts.UpvotePost)
	authGroup.Post("/post/:id/downvote", posts.DownvotePost)
	authGroup.Post("/user/edit/bio", users.EditBio)

	loginGroup := app.Group("/u")
	loginGroup.Post("/create", login.CreateAccount)
	loginGroup.Post("/login", login.AccountLogin)
	loginGroup.Get("/login", login.GetLoginPage)

}

func connectDB() {
	var err error
	database.DBConn, err = gorm.Open(sqlite.Open("stuff.db"))
	database.DBConn.AutoMigrate(&database.User{}, &database.Post{}, &database.Subspace{})

	if err != nil {
		panic("db connect failed")
	}
}

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(logger.New())

	registerRoutes(app)

	connectDB()

	err := app.Listen(":8000")

	if err != nil {
		panic(err)
	}
}
