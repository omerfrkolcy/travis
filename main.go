package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"os"
	"regexp"
	"sync"
)

type User struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	ImageURL    string `json:"image_url"`
	Status      string `json:"status"`
	UUID        string `json:"uuid"`
}

var _ = godotenv.Load()
var opt, _ = redis.ParseURL(os.Getenv("REDIS_HOST"))
var redisClient = redis.NewClient(opt)

func main() {
	instance := echo.New()

	instance.Use(middleware.Logger())
	instance.Use(middleware.Recover())

	instance.POST("/users/register", saveProfile)
	instance.POST("/users/update", updateProfile)

	instance.GET("/users/profile/:id", getProfile)
	instance.GET("/users/get-all/id", getAllIds)
	instance.GET("/users/get-all/profile", getAllProfiles)

	instance.GET("/users/delete/:id", deleteProfile)
	instance.GET("/users/flush", flushAll)

	instance.Logger.Fatal(instance.Start(":1001"))
}

func saveProfile(cnt echo.Context) error {
	user := new(User)

	if err := cnt.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if user.UUID == "" {
		user.UUID = uuid.NewString()
	} else {
		_, err := uuid.Parse(user.UUID)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	jsonValue, err := json.Marshal(user)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	redisClient.Set(generateUserRedisKey(user.UUID), jsonValue, 0)

	return cnt.JSON(http.StatusCreated, user)
}

func getProfile(cnt echo.Context) error {
	user := new(User)
	userId := cnt.Param("id")

	_, err := uuid.Parse(userId)

	if err != nil {
		return echo.NewHTTPError(http.StatusNoContent, err.Error())
	}

	userData, _ := redisClient.Get(generateUserRedisKey(userId)).Bytes()

	_ = json.Unmarshal(userData, &user)

	return cnt.JSON(http.StatusOK, user)
}

func getAllIds(cnt echo.Context) error {
	users := redisClient.Keys(generateUserRedisKey("*")).Val()

	return cnt.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func getAllProfiles(cnt echo.Context) error {
	users := make([]interface{}, 0)
	user := new(User)
	iter := redisClient.Keys(generateUserRedisKey("*")).Val()

	wg := new(sync.WaitGroup)
	wg.Add(len(iter))

	for index := range iter {
		go func(userIndex int) {
			defer func() {
				wg.Done()
			}()
			userData, _ := redisClient.Get(generateUserRedisKey(iter[userIndex])).Bytes()
			_ = json.Unmarshal(userData, &user)
			users = append(users, user)
		}(index)
	}

	wg.Wait()

	return cnt.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func updateProfile(cnt echo.Context) error {
	user := new(User)

	if err := cnt.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userRedisKey := generateUserRedisKey(user.UUID)
	uuidQuery := redisClient.Keys(userRedisKey).Val()

	if len(uuidQuery) > 0 && uuidQuery[0] == userRedisKey {
		jsonValue, err := json.Marshal(user)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		redisClient.Set(userRedisKey, jsonValue, 0)

		return cnt.JSON(http.StatusAccepted, map[string]bool{
			"status": true,
		})
	}

	return cnt.JSON(http.StatusAccepted, map[string]bool{
		"status": false,
	})
}

func deleteProfile(cnt echo.Context) error {
	userId := cnt.Param("id")

	redisClient.Del(generateUserRedisKey(userId))

	return cnt.JSON(http.StatusResetContent, true)
}

func flushAll(cnt echo.Context) error {
	redisClient.FlushAll()

	return cnt.JSON(http.StatusNoContent, map[string]bool{
		"status": true,
	})
}

/*func getUserByPhoneNumber(string phoneNumber) error {

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return cnt.JSON(http.StatusOK, userUUID) //get sql here
}*/

func generateUserRedisKey(userId string) string {
	res, err := regexp.MatchString(`^user-`, userId)

	if err != nil {
		return ""
	}

	if res == true {
		return userId
	}

	return fmt.Sprint("user-", userId)
}
