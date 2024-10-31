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
	"unicode"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	port   int    = 9090
	dir    string = "./data"
	prefix string = "api"
	DB     *gorm.DB
)

func main() {
	mux := &http.ServeMux{}
	log.Printf("Running on port %d, save data in directory %s", port, dir)
	mux.HandleFunc(fmt.Sprintf("/%s/{name}", prefix), handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

type Key struct {
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
	flag.StringVar(&prefix, "prefix", prefix, fmt.Sprintf("server api prefix, default %s", prefix))
	flag.Parse()

	prefix = strings.TrimFunc(prefix, func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		}
		if r == '/' {
			return true
		}
		return false
	})
	log.Printf("server start up with api  prefix : %s", prefix)
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
	var dbpath = filepath.Join(dir, "data.db")
	log.Printf("sqlite db path : %s", dbpath)
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	DB.AutoMigrate(&Key{})
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getHanlder(w, r)
		return
	case http.MethodPost, http.MethodPut:
		postHandler(w, r)
		return
	default:
		log.Printf("method not support: %s", r.Method)
		w.WriteHeader(405)
		return
	}
}

func getHanlder(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue("name")
	if strings.Contains(name, ":") {
		name = strings.Split(name, ":")[0]
	}
	log.Printf("get key : %s value list", name)
	var limitStr = r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = -1
	}
	var keys []Key
	if err := DB.Limit(limit).Order("updated_at desc").Where("name = ?", name).Find(&keys).Error; err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		w.Write([]byte("server database error"))
		return
	}
	bytes, err := json.Marshal(keys)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("name")
	arr := strings.Split(key, ":")
	if len(arr) != 2 {
		w.WriteHeader(400)
		w.Write([]byte("format error"))
		return
	}

	g := Key{
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
	log.Printf("successfully create key: %s", key)
}
