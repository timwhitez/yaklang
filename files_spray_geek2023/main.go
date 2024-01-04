package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"embed"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Cache represents the cache in the system
type Cache struct {
	sync.Mutex
	files map[string]string
}

func NewCache() *Cache {
	return &Cache{
		files: make(map[string]string),
	}
}

var jwtKey = []byte("123456")

// JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

//go:embed html/login.html
var loginPage []byte

//go:embed html/upload.html
var uploadPage []byte

// Number of files in the cache
const cacheSize = 4

// Number of hits needed within hitDuration for a successful spray
const hitDuration = 30 * time.Second

const FLAGSTR = "SYC{75ec9b17-2284-447f-9faa-babccc8f159c}"

//go:embed static/*
var staticFiles embed.FS

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)
		writer.Write(loginPage)
	}).Methods(http.MethodGet)

	r.HandleFunc("/flag", handleFlag).Methods(http.MethodGet)

	r.HandleFunc("/login", handleLogin).Methods(http.MethodPost)

	CSRFKey := []byte("251e79cd5d1a994c51fd316f7040f13d") // 修改为你自己的密钥，建议使用随机生成的
	CSRFMiddleware := csrf.Protect(
		CSRFKey,
		csrf.Secure(false),
		csrf.CookieName("yak_csrf"),
		csrf.FieldName("yak-token"),
		csrf.MaxAge(10),
	)

	r.Handle("/new-csrf-token", CSRFMiddleware(authenticateMiddleware(http.HandlerFunc(newToken)))).Methods(http.MethodGet)

	r.Handle("/upload", CSRFMiddleware(authenticateMiddleware(http.HandlerFunc(formHandler)))).Methods(http.MethodGet)

	r.Handle("/upload", CSRFMiddleware(authenticateMiddleware(http.HandlerFunc(uploadHandler)))).Methods(http.MethodPost)

	r.Handle("/file-list", CSRFMiddleware(authenticateMiddleware(http.HandlerFunc(fileListHandler))))

	cache.Lock()
	generateRandomFiles(cacheSize)
	fmt.Println(fmt.Sprintf("%v", cache.files))

	cache.Unlock()

	go func() {
		for range time.Tick(hitDuration) {
			dirs, err := ioutil.ReadDir("./tmp")
			if err != nil {
				fmt.Println("Failed to read directory:", err)
			} else {
				for _, d := range dirs {
					if d.IsDir() {
						err = os.RemoveAll("./tmp/" + d.Name())
						if err != nil {
							fmt.Println("Failed to delete directory:", err)
						}
					}
				}
			}
			cache.Lock()
			// Clear the cache
			for k := range cache.files {
				delete(cache.files, k)
			}
			// Generate new cache files
			generateRandomFiles(cacheSize)
			fmt.Println(fmt.Sprintf("%v", cache.files))

			currentFileNumber = 0
			cache.Unlock()
		}
	}()

	fmt.Println("Server started " + "http://127.0.0.1:8089")

	log.Fatal(http.ListenAndServe(":8089", r))
}

func handleFlag(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数 "pass"
	pass := r.URL.Query().Get("pass")

	// 拼接 cache 中的所有文件内容
	cache.Lock()
	var fileContents string
	for _, content := range cache.files {
		fileContents += content
	}
	cache.Unlock()

	// 检查 pass 是否与拼接的文件内容匹配
	if pass == fileContents {
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(fmt.Sprintf("找到 flag 啦! %s", FLAGSTR)))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("验证失败呀！"))
	}
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("jwt-token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized token"))
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr := c.Value
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized token"))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized token"))
			return
		}
		ctx := context.WithValue(r.Context(), "username", claims.Username)

		// If everything is OK, call the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	// 完整的 HTML 表单内容
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(string(uploadPage), csrf.Token(r))))
}

func newToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(csrf.Token(r)))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Check if the password is weak
	if user.Password == "123456" && user.Username != "admin" {
		expirationTime := time.Now().Add(15 * time.Minute)
		claims := &Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "jwt-token",
			Value:   tokenString,
			Expires: expirationTime,
		})
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username")
	if username == nil {
		http.Error(w, "No username in context", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("filename")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 创建一个限制读取字节数的 reader
	limitedReader := &io.LimitedReader{
		R: file,
		N: 512,
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(limitedReader)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// 创建一个临时缓存文件
	cacheFileName := fmt.Sprintf("./cache/cache-%03d.key", currentFileNumber)
	err = ioutil.WriteFile(cacheFileName, buf.Bytes(), 0644)
	if err != nil {
		http.Error(w, "Failed to create cache file", http.StatusInternalServerError)
		return
	}

	c, err := r.Cookie("jwt-token")
	if err != nil {
		http.Error(w, "Failed to get JWT token", http.StatusInternalServerError)
		return
	}

	// 计算 JWT token 的 MD5 哈希
	hasher := md5.New()
	hasher.Write([]byte(c.Value))
	md5Hash := hex.EncodeToString(hasher.Sum(nil))
	cache.Lock()
	defer cache.Unlock()
	// 使用哈希作为文件存储的目录
	finalDir := "./tmp/" + md5Hash + "/"
	os.MkdirAll(finalDir, os.ModePerm) // 确保目录存在
	// 所有检查都通过后，将文件移动到最终的位置
	finalFileName := fmt.Sprintf("%s-%03d.key", username, currentFileNumber)
	err = os.Rename(cacheFileName, finalDir+finalFileName)
	if err != nil {
		http.Error(w, "Failed to move file", http.StatusInternalServerError)
		return
	}

	currentFileNumber = (currentFileNumber + 1) % 100 // Wrap around to 0 after reaching 999

	if _, ok := cache.files[finalFileName]; ok {
		cache.files[finalFileName] = buf.String()
	}
	w.Write([]byte("上传成功"))
}

func fileListHandler(w http.ResponseWriter, r *http.Request) {
	// 从请求中获取 JWT token
	c, err := r.Cookie("jwt-token")
	if err != nil {
		http.Error(w, "Failed to get JWT token", http.StatusInternalServerError)
		return
	}

	// 计算 JWT token 的 MD5 哈希
	hasher := md5.New()
	hasher.Write([]byte(c.Value))
	md5Hash := hex.EncodeToString(hasher.Sum(nil))

	// 使用哈希来获取文件列表
	dir := "./tmp/" + md5Hash + "/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		http.Error(w, "Failed to read cache directory", http.StatusInternalServerError)
		return
	}

	// 获取所有的文件名、大小和创建时间
	filesInfo := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			fileInfo := map[string]interface{}{
				"cname": strings.SplitN(file.Name(), "-", 2)[1],
				"name":  file.Name(),
				"size":  file.Size(),                         // 文件大小，单位为字节
				"time":  file.ModTime().Format(time.RFC3339), // 文件创建时间，格式为 RFC3339
			}
			filesInfo = append(filesInfo, fileInfo)
		}
	}

	// 将文件信息数组转换为 JSON
	jsonData, err := json.Marshal(filesInfo)
	if err != nil {
		http.Error(w, "Failed to create JSON", http.StatusInternalServerError)
		return
	}

	// 设置响应头并发送响应
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

var currentFileNumber int

var cache = NewCache()

func generateRandomFiles(n int) {
	for i := 0; i < n; i++ {
		filename := fmt.Sprintf("admin-%03d.key", rand.Intn(100))
		content := uuid.New().String() // Generate a new UUID
		cache.files[filename] = content
	}
}
