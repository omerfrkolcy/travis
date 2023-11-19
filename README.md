# Travis

This is a Go project that provides a set of APIs for managing user profiles. It uses the Echo framework for handling HTTP requests and MongoDB for data storage.

## Libraries Used

- [Echo](https://echo.labstack.com/): A high performance, extensible, minimalist web framework for Go.
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver): The official MongoDB driver for Go.

## Functionality

- `getUserByPhoneNumber`: This function takes a phone number as input and returns the corresponding user profile. If the phone number is not associated with any user, it returns an error.
- `getUserByUUID`: This function takes a UUID as input and returns the corresponding user profile. If the UUID is not associated with any user, it returns an error.
- `deleteProfile`: This function deletes a user profile based on the provided user ID.
- `updateProfile`: This function updates a user profile based on the provided user data.

## Getting Started

To run this project, you will need to have Go and MongoDB installed on your machine. Once you have those installed, you can clone this repository and run `go run main.go` to start the server.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
