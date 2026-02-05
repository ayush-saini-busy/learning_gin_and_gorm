// Learning the basics of routing in Gin
package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// This struct defines a user in the system
type User struct {
	ID    int    `json:"int"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// This struct represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// List of users
var users = []User{
	{1, "John Doe", "john.doe@gmail.com", 30},
	{2, "Jane Smith", "jane.smith@gmail.com", 30},
	{3, "Max Williams", "max.williams@gmail.com", 30},
}

var nextId int = 4

func main() {
	// Utilsing default router provided by Go
	router := gin.Default()
	// Defining the routes
	router.GET("/users", getAllUsers)
	router.GET("/users/:id", getUserById)
	router.POST("/users", createUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	// router.GET("/users/:id", searchUser)

	router.Run(":8080")
}

// Handler for retrieving all the users
func getAllUsers(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    users,
	})
}

// Handler for retrieving specific user by Id
func getUserById(c *gin.Context) {
	// Used to retrieve the id parameter from the URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}
	user, _ := findUserById(id)
	if user == nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    user,
		Message: "User successfully returned",
	})
}

func createUser(c *gin.Context) {
	var newUser User
	// Checking whether JSON binding is implemented
	if err := c.ShouldBindBodyWithJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid JSON body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Checking whether passed credentials are valid or not according to format
	if err := validateUser(newUser); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	// Passing the nextId for the new user and then updating the variable
	newUser.ID = nextId
	nextId++
	users = append(users, newUser)
	// Returning
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    newUser,
		Message: "new user created",
	})
}

func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid user ID was passed",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var updatedUser User
	if err := c.ShouldBindBodyWithJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := validateUser(updatedUser); err != nil {
		c.JSON(http.StatusNotAcceptable, Response{
			Success: false,
			Error:   err.Error(),
			Code:    http.StatusNotAcceptable,
		})
		return
	}

	user, index := findUserById(id)
	if user == nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "user not found",
			Code:    http.StatusBadRequest,
		})
		return
	}
	updatedUser.ID = id
	users[index] = updatedUser

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    updatedUser,
		Message: "User data updated successfully",
	})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid user ID was passed",
			Code:    http.StatusBadRequest,
		})
		return
	}
	_, index := findUserById(id)
	if index == -1 {
		c.JSON(http.StatusNotAcceptable, Response{
			Success: false,
			Error:   "user not found",
			Code:    http.StatusNotAcceptable,
		})
		return
	}

	users = append(users[:index], users[index+1:]...)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "user deleted successfully",
	})
}

// func searchUser(c *gin.Context) {

// }

// Helper function to find users by ID
func findUserById(id int) (*User, int) {
	for i, user := range users {
		if user.ID == id {
			return &user, i
		}
	}
	return nil, -1
}

// Helper function for validating user input
func validateUser(user User) error {
	if strings.TrimSpace(user.Name) == "" {
		return gin.Error{
			Err:  http.ErrMissingFile,
			Type: gin.ErrorTypeBind,
			Meta: "name is a required field",
		}
	}

	if strings.TrimSpace(user.Email) == "" || !strings.Contains(user.Email, "@") {
		return gin.Error{
			Err:  http.ErrMissingFile,
			Type: gin.ErrorTypeBind,
			Meta: "valid email is required",
		}
	}
	return nil
}
