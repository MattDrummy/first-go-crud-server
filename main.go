package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
  "log"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
  "github.com/gin-contrib/cors"
  "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
)

type Student struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Firstname string
	Lastname  string
	Age       float64
	Gender    string
	Current   bool
	Awesome   float64
	Timestamp int32
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	gin.DisableConsoleColor()

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	session, err := mgo.Dial(os.Getenv("MONGOLAB_URI"))
	students := session.DB(os.Getenv("DATABASE_NAME")).C("student")
	fmt.Println("*printing students", students)
	if err != nil {
		fmt.Println(err)
	}

	router := gin.Default()
	fmt.Println(os.Getenv("SITE_URL"))
  router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("SITE_URL")},
		AllowMethods:     []string{"PUT", "PATCH", "DELETE", "GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))  // router.Use(cors.New(config)) // adding cors needs to be done before you add router modes

	router.GET("/all", func(c *gin.Context) {
		fmt.Println("hitting all route")
		var data []Student
		students.Find(nil).All(&data)
		c.JSON(http.StatusOK, gin.H{
			"students": data,
		})
	})

	router.POST("/new", func(c *gin.Context) {
		firstName := c.PostForm("firstname")
		lastName := c.PostForm("lastname")
		age, _ := strconv.ParseFloat(c.PostForm("age"), 64)
		gender := c.DefaultPostForm("gender", "N/A")
		awesome, _ := strconv.ParseFloat(c.PostForm("awesome"), 64)
		query := bson.M{"current": true}
		change := bson.M{"$set": bson.M{"current": false}}
		students.UpdateAll(query, change)
		err = students.Insert(&Student{
			Firstname: firstName,
			Lastname:  lastName,
			Age:       age,
			Gender:    gender,
			Awesome:   awesome,
			Current:   true,
			Timestamp: int32(time.Now().Unix()),
		})
		if err != nil {
			fmt.Println(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Posted",
		})
	})

	router.DELETE("/delete/:time", func(c *gin.Context) {
		time, _ := strconv.Atoi(c.Param("time"))
		err := students.Remove(bson.M{"timestamp": time})
		if err != nil {
			fmt.Println(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Deleted",
		})
	})

	router.PUT("/update/:id", func(c *gin.Context) {
		firstName := c.PostForm("firstname")
		lastName := c.PostForm("lastname")
		age, _ := strconv.ParseFloat(c.PostForm("age"), 64)
		gender := c.DefaultPostForm("gender", "N/A")
		awesome, _ := strconv.ParseFloat(c.PostForm("awesome"), 64)
		current := c.PostForm("current")
		time, _ := strconv.Atoi(c.Param("id"))
		update := bson.M{
			"firstname": firstName,
			"lastname":  lastName,
			"age":       age,
			"gender":    gender,
			"awesome":   awesome,
			"current":   current,
		}
		change := bson.M{"$set": update}
		students.Update(bson.M{"timestamp": time}, change)

		var data []Student
		students.Find(nil).All(&data)
		c.JSON(http.StatusOK, gin.H{
			"students": data,
		})
	})

  server, err := socketio.NewServer(nil)
  if err != nil {
    fmt.Println(err)
  }

  server.On("connection", func(so socketio.Socket) {
    so.On("open", func(user [2]string){
      fmt.Println(user[0] + " connected to " + user[1])
      so.BroadcastTo(user[1], "message", user[0] + " connected to " + user[1])
      so.Join(user[1])
    })
    so.On("message", func(msg [2]string) {
      fmt.Printf("%+v\n", msg)
      so.Emit("message", msg[1])
      so.BroadcastTo(msg[0], "message", msg[1])
    })
    so.On("disconnection", func() {
      fmt.Println("on disconnect")
    })
  })
  server.On("error", func(so socketio.Socket, err error) {
    log.Println("error:", err)
  })


  router.GET("/socket.io/", gin.WrapH(server))
	router.POST("/socket.io/", gin.WrapH(server))
  fmt.Println("server ready")
  router.Run(":" + os.Getenv("PORT"))

}
