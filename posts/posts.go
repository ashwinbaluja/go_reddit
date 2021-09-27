package posts

import (
	"errors"
	"log"
	"sort"
	"strconv"

	"github.com/azbxx/fiber-reddit-test/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PostReq struct {
	Title string
	Body  string
}

type PostBodyTemplate struct {
	ID          uint
	Title       string
	TitleAuthor string
	Body        string
	Points      int
	Percent     float32
}

func GetPostHomepage(c *fiber.Ctx) error {
	var posts []database.Post
	database.DBConn.Find(&posts)

	postsTemplate := make([]PostBodyTemplate, len(posts))
	log.Println(len(posts))

	for i, s := range posts {
		user := new(database.User)

		postsTemplate[i].ID = s.ID
		postsTemplate[i].Title = s.Title
		database.DBConn.First(&user, s.UserID)
		postsTemplate[i].TitleAuthor = s.Title + " by " + user.Username
		postsTemplate[i].Body = s.Body
		postsTemplate[i].Points = s.Upvotes - s.Downvotes
		var percent float32
		if s.Upvotes != 0 {
			percent = float32(int(float32(s.Upvotes)/float32(s.Upvotes+s.Downvotes)*10000)) / 100
		} else {
			percent = 0.
		}
		postsTemplate[i].Percent = percent
	}
	sort.Slice(postsTemplate, func(i, j int) bool {
		return postsTemplate[i].Points < postsTemplate[j].Points
	})
	for i, j := 0, len(postsTemplate)-1; i < j; i, j = i+1, j-1 {
		postsTemplate[i], postsTemplate[j] = postsTemplate[j], postsTemplate[i]
	}
	return c.Render("home", fiber.Map{
		"Posts": postsTemplate,
	})
}

func GetPost(c *fiber.Ctx) error {
	post := new(database.Post)
	user := new(database.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	err = database.DBConn.First(&post, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON("{'failed': 'post not found'}")
	}
	database.DBConn.First(&user, post.UserID)
	var percent float32
	if post.Upvotes != 0 {
		percent = float32(int(float32(post.Upvotes)/float32(post.Upvotes+post.Downvotes)*10000)) / 100
	} else {
		percent = 0.
	}
	return c.Render("post", fiber.Map{
		"ID":          post.ID,
		"Title":       post.Title,
		"TitleAuthor": post.Title + " by " + user.Username,
		"Body":        post.Body,
		"Points":      post.Upvotes - post.Downvotes,
		"Percent":     percent,
	})
}

func CreatePost(c *fiber.Ctx) error {
	postReq := new(PostReq)
	err := c.BodyParser(postReq)
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	id, _ := strconv.Atoi(c.Locals("claims").(jwt.MapClaims)["user_id"].(string))
	post := database.Post{
		Title:     postReq.Title,
		Body:      postReq.Body,
		Upvotes:   0,
		Downvotes: 0,
		UserID:    id,
	}
	database.DBConn.Create(&post)
	return c.JSON(post)
}

func UpvotePost(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	database.DBConn.Exec("UPDATE posts SET upvotes = upvotes + 1 WHERE id = ?", id)

	post := new(database.Post)
	database.DBConn.First(&post, id)

	return c.JSON(post.Upvotes)
}

func DownvotePost(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON("{'failed': 'invalid format'}")
	}
	database.DBConn.Exec("UPDATE posts SET downvotes = downvotes + 1 WHERE id = ?", id)

	post := new(database.Post)
	database.DBConn.First(&post, id)

	return c.JSON(post.Downvotes)
}
