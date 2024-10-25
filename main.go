package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	port int    = 9090
	dir  string = "./data"
	DB   *gorm.DB
)

func main() {
	mux := &http.ServeMux{}
	log.Printf("running on port %d, save data in directory %s", port, dir)
	mux.HandleFunc("/image/{name}", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

type Image struct {
	Name      string    `json:"name" gorm:"index:idx"`
	Tag       string    `json:"tag" gorm:"index:idx"`
	ID        uint      `gorm:"primarykey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time
}

func init() {
	// prase flag
	flag.IntVar(&port, "port", 9090, fmt.Sprintf("server port, default %d", port))
	flag.StringVar(&dir, "dir", "./data", fmt.Sprintf("server data dir, default %s", dir))
	flag.Parse()

	// check dir exist, if not exist, create it.
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			if mkerr := os.MkdirAll(dir, 0755); mkerr != nil {
				log.Fatal(mkerr)
			}
		}
	}
	// init sqlite db
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "data.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
		var limitStr = r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = -1
		}
		var images []Image
		if err := DB.Limit(limit).Order("updated_at desc").Where("name = ?", name).Find(&images).Error; err != nil {
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
	if len(arr) != 2 {
		w.WriteHeader(400)
		w.Write([]byte("format error"))
		return
	}

	g := Image{
		Name: arr[0],
		Tag:  arr[1],
	}

	if err := DB.Where("name = ? and tag = ?", arr[0], arr[1]).First(&g).Error; err == nil {
		g.UpdatedAt = time.Now()
		DB.Save(&g)
		return
	}
	if err := DB.Save(&g).Error; err != nil {
		w.WriteHeader(500)
		w.Write([]byte("server error"))
		return
	}
	log.Printf("successfully create image: %s", image)
}
