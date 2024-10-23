package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	port int = 9090
	DB   *gorm.DB
)

func main() {
	mux := &http.ServeMux{}

	// set up the endpoint with a "greeting" wildcard.
	mux.HandleFunc("/image/{name}", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

type Image struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
	gorm.Model
}

func init() {
	// prase flag
	flag.IntVar(&port, "p", 9090, fmt.Sprintf("server port, default %d", port))
	flag.Parse()

	// init sqlite db
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	DB.AutoMigrate(&Image{})
}
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var name = r.PathValue("name")
		log.Printf("get image tags: %s", name)
		var images []Image
		if err := DB.Order("updated_at desc").Where("name = ?", name).Find(&images).Error; err != nil {
			log.Println(err.Error())
			w.WriteHeader(500)
			w.Write([]byte("server database error"))
			return
		}
		bytes, err := json.Marshal(images)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
		return
	}
	image := r.PathValue("name")
	arr := strings.Split(image, ":")
	var name = arr[0]
	var tag = arr[1]
	g := Image{
		Name: name,
		Tag:  tag,
	}

	if err := DB.Where("name = ? and tag = ?", name, tag).First(&g).Error; err == nil {
		g.UpdatedAt = time.Now()
		DB.Save(&g)
		return
	}
	if err := DB.Save(&g).Error; err != nil {
		w.WriteHeader(500)
		w.Write([]byte("server error"))
		return
	}
	log.Printf("successfully create image: %s:%s", name, tag)
}
