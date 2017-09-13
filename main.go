package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "net/http"
    "time"
    "strconv"
    "io"
    "os"
)

type Student struct {
  ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
  Firstname  string
  Lastname   string
  Age        float64
  Gender     string
  Current    bool
  Awesome    float64
  Timestamp  int32
}

func main() {

  gin.DisableConsoleColor()

  f,_ := os.Create("gin.log")
  gin.DefaultWriter = io.MultiWriter(f)

  session, err := mgo.Dial("localhost")
  students := session.DB("student").C("students")
  if err != nil {
    panic(err)
  }

  router := gin.Default()

  router.GET("/all", func(c *gin.Context) {
    var data []Student
    students.Find(nil).All(&data)
    c.JSON(http.StatusOK, gin.H{
      "students": data,
    })
  })

  router.POST("/new", func(c *gin.Context){
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
        Lastname: lastName,
        Age: age,
        Gender: gender,
        Awesome: awesome,
        Current: true,
        Timestamp: int32(time.Now().Unix()),
      })
    if err != nil {
      panic(err)
    }

    c.JSON(http.StatusOK, gin.H{
      "message": "Posted",
    })
  })

  router.DELETE("/delete/:id", func(c *gin.Context) {
    time, _ := strconv.Atoi(c.Param("id"))
    err := students.Remove(bson.M{"timestamp": time})
    if err != nil {
      panic(err)
    }

    c.JSON(http.StatusOK, gin.H{
      "message": "Deleted",
    })
  })

  router.PUT("/update/:id", func(c *gin.Context)  {
    firstName := c.PostForm("firstname")
    lastName := c.PostForm("lastname")
    age, _ := strconv.ParseFloat(c.PostForm("age"), 64)
    gender := c.DefaultPostForm("gender", "N/A")
    awesome, _ := strconv.ParseFloat(c.PostForm("awesome"), 64)
    current := c.PostForm("current")
    time, _ := strconv.Atoi(c.Param("id"))
    update := bson.M{
      "firstname": firstName,
      "lastname": lastName,
      "age": age,
      "gender": gender,
      "awesome": awesome,
      "current": current,
    }
    change := bson.M{"$set": update}
    students.Update(bson.M{"timestamp": time}, change)

    var data []Student
    students.Find(nil).All(&data)
    c.JSON(http.StatusOK, gin.H{
      "students": data,
    })
  })

  fmt.Println("server ready")
  router.Run(":8080")
}
