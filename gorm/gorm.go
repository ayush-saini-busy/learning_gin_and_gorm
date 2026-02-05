package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Age       int    `gorm:"check:age>0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ConnectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=bipl dbname=gorm_demo port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("An error occured while connecting to Postgres")
		return nil, err
	}
	// Auto-migrate the user schema
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}
	return db, nil
}

// Create a new user in the database
func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// Retrieves the user with the specified id
func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Retrieves all users in the database
func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Updates an existing user's information
func UpdateUser(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}

func DeleteUser(db *gorm.DB, id uint) error {
	return db.Delete(&User{}, id).Error
}

func main() {
	// Performing CRUD operations
	db, err := ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	user := &User{
		Name:  "Jack",
		Email: "jack.sparrow@gmail.com",
		Age:   50,
	}
	// Creating user
	if err := CreateUser(db, user); err != nil {
		log.Fatalf("An error occured while creating new user : %v", err)
	}
	fmt.Printf("Created user: %+v\n", *user)

	// Fetching user
	fetchedUser, err := GetUserByID(db, user.ID)
	if err != nil {
		log.Fatalf("Failed to fetch user data: %+v\n", err)
	}
	fmt.Println("Fetched user details:", *fetchedUser)

	// Fetching all users
	all_users, err := GetAllUsers(db)
	if err != nil {
		log.Fatalf("Failed to fetch user information: %+v\n", err)
	}
	fmt.Println("Users in the DB : \n", all_users)

	// Updating user information
	fetchedUser.Email = "lord.commander@gmail.com"
	if err := UpdateUser(db, fetchedUser); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Println("Updated user age to : ", fetchedUser.Email)

	// Deleting user information
	if err := DeleteUser(db, user.ID); err != nil {
		log.Fatalf("Failed to delete user data")
	}
	fmt.Println("User deleted successfully")
}
