package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uint
	Username string
	Password string
}

type ToDo struct {
	gorm.Model
	ID     uint
	UserID uint
	Name   string
}

func initDb() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&ToDo{}, &User{})

	return db
}

func getTodosForUser(db *gorm.DB, c *fiber.Ctx) []ToDo {
	userId := c.Cookies("userId")

	intUserId, err := strconv.Atoi(userId)
	if err != nil {
		c.Render("login", fiber.Map{})
	}

	uintUserId := uint(intUserId)

	var todos []ToDo
	db.Where("user_id = ?", uintUserId).Find(&todos)

	return todos
}

func main() {
	db := initDb()

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		user := new(User)
		if err := c.BodyParser(user); err != nil {
			return err
		}

		var u User
		db.Where("username = ?", user.Username).First(&u)

		if u.Password == user.Password {
			c.Cookie(&fiber.Cookie{
				Name:     "userId",
				Value:    fmt.Sprintf("%d", u.ID),
				Expires:  time.Now().Add(24 * time.Hour),
				HTTPOnly: true,
			})

			return c.Render("home", fiber.Map{})
		}

		return c.Render("login", fiber.Map{})
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("register", fiber.Map{})
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		user := new(User)
		if err := c.BodyParser(user); err != nil {
			return err
		}

		db.Create(&user)

		return c.Render("login", fiber.Map{})
	})

	app.Get("/home", func(c *fiber.Ctx) error {
		userId := c.Cookies("userId")

		if userId == "" {
			return c.Redirect("/login")
		}

		return c.Render("home", fiber.Map{})
	})

	app.Get("/todos", func(c *fiber.Ctx) error {
		todos := getTodosForUser(db, c)

		return c.Render("todos", fiber.Map{
			"Todos": todos,
		})
	})

	app.Post("/todos", func(c *fiber.Ctx) error {
		todo := new(ToDo)
		if err := c.BodyParser(todo); err != nil {
			return err
		}

		userId := c.Cookies("userId")

		intUserId, err := strconv.Atoi(userId)
		if err != nil {
			c.Render("login", fiber.Map{})
		}

		uintUserId := uint(intUserId)

		todo.UserID = uintUserId

		db.Create(&todo)

		todos := getTodosForUser(db, c)
		return c.Render("todos", fiber.Map{"Todos": todos})
	})

	app.Delete("/todos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		db.Delete(&ToDo{}, id)

		todos := getTodosForUser(db, c)
		return c.Render("todos", fiber.Map{"Todos": todos})
	})

	app.Listen(":3000")
}
