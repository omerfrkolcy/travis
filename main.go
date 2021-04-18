package main

import (
	"context"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var (
	_ = godotenv.Load()
	ctx = context.Background()
	mongoClient, _ = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DB_HOST")))
	mongoDatabase = mongoClient.Database(os.Getenv("USER_DB"))
	mongoCollection = mongoDatabase.Collection(os.Getenv("USER_COLLECTION"))
)

type User struct {
	Name        string `json:"name" bson:"name"`
	PhoneNumber string `json:"phone_number" bson:"phone_number"`
	ImageURL    string `json:"image_url" bson:"image_url"`
	Status      string `json:"status" bson:"status"`
	UUID        string `json:"uuid" bson:"_id"`
}

func main() {
	instance := echo.New()

	instance.Use(middleware.Logger())
	instance.Use(middleware.Recover())

	instance.POST("/users/register", saveProfile)
	instance.POST("/users/update", updateProfile)

	instance.GET("/users/profile/:id", getProfile)
	instance.GET("/users/get-all/profile", getAllProfiles)

	instance.GET("/users/delete/:id", deleteProfile)

	instance.Logger.Fatal(instance.Start(":1001"))
}

func saveProfile(cnt echo.Context) error {
	user := new(User)

	if err := cnt.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if user.PhoneNumber == "" {
		return echo.NewHTTPError(http.StatusBadRequest, false)
	}

	userByPhone := getUserByPhoneNumber(user.PhoneNumber)

	if userByPhone.UUID != "" {
		user.UUID = userByPhone.UUID
		bsonUser, err := bson.Marshal(&user)
		bson.Unmarshal(bsonUser, &user)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		mongoCollection.UpdateByID(ctx, userByPhone.UUID, bson.M{
			"$set": user,
		})

		return cnt.JSON(http.StatusOK, user)
	}

	if  user.UUID == "" {
		user.UUID = uuid.NewString()
	} else {
		_, err := uuid.Parse(user.UUID)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	userByUUID := getUserByUUID(user.UUID)

	if userByUUID.PhoneNumber != "" {
		user.PhoneNumber = userByUUID.PhoneNumber
	}

	mongoCollection.InsertOne(ctx, user)

	return cnt.JSON(http.StatusOK, user)
}

func getProfile(cnt echo.Context) error {
	userId := cnt.Param("id")
	_, err := uuid.Parse(userId)

	if err != nil {
		return echo.NewHTTPError(http.StatusOK, err.Error())
	}
	
	return cnt.JSON(http.StatusOK, getUserByUUID(userId))
}

func getAllProfiles(cnt echo.Context) error {
	users := make([]*User, 0)
	user := new(User)

	userList, err := mongoCollection.Find(ctx, bson.D{})

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} 

	for userList.Next(ctx) {
		userList.Decode(&user)
		users = append(users, user)
	}

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

	if user.UUID == "" && user.PhoneNumber == "" {
		return echo.NewHTTPError(http.StatusBadRequest, false)
	}

	existingUserByPhoneNumber := getUserByPhoneNumber(user.PhoneNumber)
	existingUserByUUID := getUserByUUID(user.UUID)
	
	if existingUserByPhoneNumber.PhoneNumber == "" && existingUserByUUID.UUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, false)
	}

	if existingUserByUUID.UUID != "" {
		user.PhoneNumber = existingUserByUUID.PhoneNumber
	} else if existingUserByPhoneNumber.PhoneNumber != "" {
		user.UUID = existingUserByPhoneNumber.UUID
	}

	bsonUser, err := bson.Marshal(&user)
	bson.Unmarshal(bsonUser, &user)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mongoCollection.UpdateByID(ctx, user.UUID, bson.M{
		"$set": user,
	})

	return cnt.JSON(http.StatusOK, user)
}

func deleteProfile(cnt echo.Context) error {
	userId := cnt.Param("id")

	_, err := mongoCollection.DeleteOne(ctx, bson.M{"_id": userId})

	return cnt.JSON(http.StatusOK, err)
}

func getUserByPhoneNumber(phoneNumber string) *User {
	user := new(User)
	
	if phoneNumber == "" {
		return user
	}

	mongoCollection.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&user)

	return user
}

func getUserByUUID(uuid string) *User {
	user := new(User)

	if uuid == "" {
		return user
	}
	
	mongoCollection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&user)

	return user
}