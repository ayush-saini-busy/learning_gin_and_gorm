package main

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Data Models
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Posts     []Post `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Post struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	UserID    uint   `gorm:"not null;index"`
	User      User
	Tags      []Tag `gorm:"many2many:post_tags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique;not null"`
	Posts []Post `gorm:"many2many:post_tags;"`
}

// Connecting to the database
func ConnectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=bipl dbname=gorm_demo port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&User{}, &Post{}, &Tag{}); err != nil {
		return nil, err
	}
	return db, nil
}

// Create user with posts
func CreateUserWithPost(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// Get user with posts
func GetUserWithPost(db *gorm.DB, userId uint) (*User, error) {
	var user User
	if err := db.Preload("Posts").
		First(&user, userId).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create post with tag
func CreatePostWithTag(db *gorm.DB, post *Post, tagNames []string) error {
	var tags []Tag
	for _, name := range tagNames {
		var tag Tag
		if err := db.FirstOrCreate(&tag, Tag{Name: name}).Error; err != nil {
			return err
		}
		tags = append(tags, tag)
	}
	post.Tags = tags
	return db.Create(post).Error
}

func GetPostWithTag(db *gorm.DB, tagName string) ([]Post, error) {
	var posts []Post

	err := db.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Joins("JOIN tags ON tags.id = post_tags.tag_id").
		Where("tags.name = ?", tagName).
		Preload("User").
		Preload("Tags").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func AddTagToPost(db *gorm.DB, postID uint, tagNames []string) error {
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		return err
	}

	var tags []Tag
	for _, name := range tagNames {
		var tag Tag
		if err := db.FirstOrCreate(&tag, Tag{Name: name}).Error; err != nil {
			return err
		}
		tags = append(tags, tag)
	}

	return db.Model(&post).Association("Tags").Append(&tags)
}

func GetPostWithUserAndTags(db *gorm.DB, postID uint) (*Post, error) {
	var post Post
	if err := db.Preload("User").
		Preload("Tags").
		First(&post, postID).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func main() {

}
