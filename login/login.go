package login

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/azbxx/fiber-reddit-test/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Login struct {
	User string `json:"user" xml:"user" form:"user"`
	Pass string `json:"pass" xml:"pass" form:"pass"`
}

func GetLoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
}
func GenerateToken(authedUser *database.User) (string, error) {
	accessUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	godotenv.Load(".env")
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["token_uuid"] = accessUUID.String()
	atClaims["user_id"] = strconv.Itoa(int(authedUser.ID))
	atClaims["username"] = authedUser.Username
	atClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

func JWTMiddleware(c *fiber.Ctx) error {
	godotenv.Load(".env")
	header := c.Get(fiber.HeaderAuthorization)
	if len(header) > 0 && strings.ToLower(header[:6]) == "bearer" {
		splitted := strings.Split(header, " ")
		if len(splitted) == 2 {
			tokenStr := splitted[1]
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Println(token.Header["alg"])
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("ACCESS_SECRET")), nil
			})
			if err != nil {
				log.Println(err)
				return c.JSON("{'failed': 'token parse failed'}")
			}
			if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
				return err
			}
			claims, ok := token.Claims.(jwt.MapClaims)

			if ok && token.Valid {
				//c.Locals("id", claims["user_id"].(string))
				c.Locals("claims", claims)
				return c.Next()
			}
		}
	}
	return c.JSON("{'failed': 'token missing'}")
}

func CreateAccount(c *fiber.Ctx) error {
	user := new(database.User)
	db := database.DBConn
	u := new(Login)
	err := c.BodyParser(u)
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	err = db.Where("Username = ?", u.User).First(&user).Error
	newUser := database.User{Username: u.User}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Pass), bcrypt.MinCost)
		if err != nil {
			return c.JSON("{'failed' : 'invalid password'}")
		}
		newUser.HashedPassword = hash
		log.Println(err)
		database.DBConn.Create(&newUser)
		return c.JSON(newUser)
	} else {
		return c.JSON("{'failed': 'user exists already'}")
	}
}

func AccountLogin(c *fiber.Ctx) error {
	db := database.DBConn
	u := new(Login)
	err := c.BodyParser(u)
	user := new(database.User)
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	err = db.Where("Username = ?", u.User).First(&user).Error
	if err != nil {
		return c.JSON("{'failed' : 'user does not exist'}")
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(u.Pass))
	if err != nil {
		return c.JSON("{'failed' : 'invalid password'}")
	}
	var token string
	token, err = GenerateToken(user)
	if err != nil {
		return c.JSON("{'failed' : 'token generation failed'}")
	}
	return c.SendString(token)
}
