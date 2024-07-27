package db

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Setup() Connection {
	fmt.Println("initialising database connection...")
	db, err := gorm.Open(sqlite.Open("todos.db"), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Todo{})
	return NewConnection(db)
}

func Close(db *gorm.DB) {
	fmt.Println("closing database connection...")
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to close database")
	}
	sqlDB.Close()
}

func NewConnection(db *gorm.DB) Connection {
	return Connection{db}
}

type Connection struct {
	DB *gorm.DB
}

func (d Connection) GetAllUsers(users *[]User) error {
	return d.DB.Find(users).Error
}

func (d Connection) GetUserByID(id string, user *User) error {
	err := d.DB.Where("id=?", id).Find(user)
	return err.Error
}

func (d Connection) CreateUser(user *User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return d.DB.Create(user).Error
}

func (d Connection) PaginateUsers(users *[]User, query *string, page, perPage *int) error {
	q := d.DB
	if query != nil && *query != "" {
		q = q.Where("text like ?", "%"+*query+"%")
	}
	if perPage != nil {
		q = q.Limit(*perPage)
	}
	if page != nil {
		q = q.Offset(*perPage * (*page - 1))
	}
	return q.Find(users).Error

}

func (d Connection) UpdateUser(id string, name, email, username *string) error {
	var user User
	err := d.DB.Where("id=?", id).First(&user)
	if err.Error != nil {
		return err.Error
	}
	if name != nil {
		user.Name = *name
	}
	if email != nil {
		user.Email = *email
	}
	if username != nil {
		user.Username = *username
	}
	user.UpdatedAt = time.Now()
	return d.DB.Save(&user).Error
}

func (d Connection) DeleteUser(id string) error {
	var user User
	d.DB.Where("id=?", id).First(&user)
	return d.DB.Delete(&user).Error
}

func (d Connection) GetAllTodos(todos *[]Todo) error {
	return d.DB.Find(todos).Error
}

func (d Connection) GetTodosByUserID(todos *[]Todo, userID string) error {
	return d.DB.Where("user_id=?", userID).Find(todos).Error
}

func (d Connection) CreateTodo(todo *Todo) error {
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	return d.DB.Create(todo).Error
}

func (d Connection) PaginateTodos(todos *[]Todo, query *string, page, perPage *int) error {
	q := d.DB
	if query != nil && *query != "" {
		q = q.Where("text like ?", "%"+*query+"%")
	}
	if perPage != nil {
		q = q.Limit(*perPage)
	}
	if page != nil {
		q = q.Offset(*perPage * (*page - 1))
	}
	return q.Find(todos).Error
}

func (d Connection) UpdateTodo(id, text string) error {
	var todo Todo
	d.DB.Where("id=?", id).First(&todo)
	if text != "" {
		todo.Text = text
	}
	todo.UpdatedAt = time.Now()
	return d.DB.Save(&todo).Error
}

func (d Connection) UpdateTodoAsDone(id string) error {
	var todo Todo
	d.DB.Where("id=?", id).First(&todo)
	todo.Done = true
	todo.UpdatedAt = time.Now()
	return d.DB.Save(&todo).Error
}

func (d Connection) UpdateTodoAsIncomplete(id string) error {
	var todo Todo
	d.DB.Where("id=?", id).First(&todo)
	todo.Done = false
	todo.UpdatedAt = time.Now()
	return d.DB.Save(&todo).Error
}

func (d Connection) GetTodoByID(id string, todo *Todo) error {
	err := d.DB.Where("id=?", id).First(todo)
	return err.Error
}

func (d Connection) DeleteTodo(id string) error {
	var todo Todo
	d.DB.Where("id=?", id).First(&todo)
	return d.DB.Delete(&todo).Error
}

func (d Connection) DeleteEverything() error {
	todos := []Todo{}
	err := d.DB.Where("id is not null").Delete(&todos)
	if err.Error != nil {
		return err.Error
	}
	users := []User{}
	err = d.DB.Where("id is not null").Delete(&users)
	return err.Error
}
