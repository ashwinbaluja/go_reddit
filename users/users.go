package users

import (
	"errors"
	"strconv"

	"github.com/azbxx/fiber-reddit-test/database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type bioReq struct {
	Bio string
}

func GetUser(c *fiber.Ctx) error {
	user := new(database.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON("{'failed': 'invalid id'}")
	}
	err = database.DBConn.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON("{'failed': 'user does not exist'}")
	}
	return c.JSON(user)

}

func EditBio(c *fiber.Ctx) error {
	bio := new(bioReq)
	user := new(database.User)
	err := c.BodyParser(bio)
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON("{'failed': 'invalid id'}")
	}
	err = database.DBConn.First(&user, id).Error
	user.Bio = bio.Bio
	database.DBConn.Save(user)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON("{'failed': 'user does not exist'}")
	}
	return c.JSON(user.Bio)
}
