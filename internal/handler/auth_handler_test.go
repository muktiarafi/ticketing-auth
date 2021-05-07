package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
)

func TestAuthHandlerNew(t *testing.T) {
	t.Run("create user with valid request body", func(t *testing.T) {
		user := &model.UserDTO{Email: "bambank@gmail.com", Password: "12345678"}
		_, responseBody := createUser(user)

		apiResponse := model.SuccessResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusCreated, apiResponse.Status)
	})

	t.Run("create user with invalid email", func(t *testing.T) {
		user := &model.UserDTO{Email: "apa", Password: "12345678"}
		_, responseBody := createUser(user)

		apiResponse := model.ErrorResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusBadRequest, apiResponse.Status)

		if len(apiResponse.Errors) != 1 {
			t.Errorf("Expected 1 error from invalid email")
		}
	})

	t.Run("create user with invalid password", func(t *testing.T) {
		user := &model.UserDTO{Email: "bambank@gmail.com", Password: "123"}
		_, responseBody := createUser(user)

		apiResponse := model.ErrorResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusBadRequest, apiResponse.Status)

		if len(apiResponse.Errors) != 1 {
			t.Errorf("Expected 1 error from invalid password")
		}
	})

	t.Run("create user with invalid request body", func(t *testing.T) {
		user := &model.UserDTO{Email: "12", Password: "123"}
		_, responseBody := createUser(user)

		apiResponse := model.ErrorResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusBadRequest, apiResponse.Status)

		if len(apiResponse.Errors) != 2 {
			t.Errorf("Expected 2 error from both email and password")
		}
	})

	t.Run("Create user with duplicate email", func(t *testing.T) {
		user1 := &model.UserDTO{Email: "paijo@gmail.com", Password: "12345678"}
		createUser(user1)

		user2 := &model.UserDTO{Email: "paijo@gmail.com", Password: "12345678"}
		_, responseBody := createUser(user2)

		apiResponse := model.ErrorResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusConflict, apiResponse.Status)
	})

	t.Run("set cookie after succesfully signing up", func(t *testing.T) {
		user := &model.UserDTO{Email: "paimin@gmail.com", Password: "12345678"}
		response, responseBody := createUser(user)

		apiResponse := model.ErrorResponse{}
		json.Unmarshal(responseBody, &apiResponse)

		assertResponseCode(t, http.StatusCreated, apiResponse.Status)

		assertCookiExist(t, response)
	})
}

func TesAuthtHandlerSignIn(t *testing.T) {
	t.Run("signin with already created user", func(t *testing.T) {
		user := &model.UserDTO{Email: "apx@gmail.com", Password: "12345678"}
		_, responseBody := createUser(user)

		apiResponse := model.BaseResponse{}
		json.Unmarshal(responseBody, &apiResponse)
		assertResponseCode(t, http.StatusCreated, apiResponse.Status)

		response, responseBody := signIn(user)

		json.Unmarshal(responseBody, &apiResponse)
		assertResponseCode(t, http.StatusOK, apiResponse.Status)

		assertCookiExist(t, response)
	})

	t.Run("sigin with wrong password", func(t *testing.T) {
		user := &model.UserDTO{Email: "pr@gmail.com", Password: "12345678"}
		_, responseBody := createUser(user)

		apiResponse := model.BaseResponse{}
		json.Unmarshal(responseBody, &apiResponse)
		assertResponseCode(t, http.StatusCreated, apiResponse.Status)

		signInDTO := &model.UserDTO{user.Email, "123123123123"}
		response, responseBody := signIn(signInDTO)
		assertResponseCode(t, http.StatusBadRequest, response.Code)
	})

	t.Run("signin with noneexistent email", func(t *testing.T) {
		signInDTO := &model.UserDTO{Email: "ayaya@gmail.com", Password: "werwerwerwer"}
		response, _ := signIn(signInDTO)
		assertResponseCode(t, http.StatusBadRequest, response.Code)
	})
}

func TestAuthHandlerCurrentUser(t *testing.T) {
	t.Run("get current user after signing up", func(t *testing.T) {
		user := &model.UserDTO{Email: "lollolo@gmail.com", Password: "12345678"}
		response, _ := createUser(user)
		assertResponseCode(t, http.StatusCreated, response.Code)

		assertCookiExist(t, response)

		request := httptest.NewRequest(http.MethodGet, "/api/auth", nil)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(response.Result().Cookies()[0])
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)

		responseBody, _ := ioutil.ReadAll(response.Body)
		currentUser := struct {
			Data entity.User `json:"data"`
		}{entity.User{}}
		json.Unmarshal(responseBody, &currentUser)

		got := currentUser.Data.Email
		want := user.Email
		if got != want {
			t.Errorf("Expecting current user email to be %s but got %s instead", want, got)
		}
	})

	t.Run("get user without cookie", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/auth", nil)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusBadRequest, response.Code)
	})
}

func TestAuthHandlerSignOut(t *testing.T) {
	t.Run("create user then remove cookie", func(t *testing.T) {
		user := &model.UserDTO{"pipio@gmail.com", "12345678"}
		response, _ := createUser(user)
		assertResponseCode(t, http.StatusCreated, response.Code)
		assertCookiExist(t, response)

		request := httptest.NewRequest(http.MethodPost, "/api/auth/signout", nil)
		request.AddCookie(response.Result().Cookies()[0])
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assertResponseCode(t, http.StatusOK, response.Code)
	})
}

func createUser(userDTO *model.UserDTO) (*httptest.ResponseRecorder, []byte) {
	userJSON, _ := json.Marshal(userDTO)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBuffer(userJSON))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	responseBody, _ := ioutil.ReadAll(response.Body)

	return response, responseBody
}

func signIn(userDTO *model.UserDTO) (*httptest.ResponseRecorder, []byte) {
	userJSON, _ := json.Marshal(userDTO)

	request := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(userJSON))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	responseBody, _ := ioutil.ReadAll(response.Body)

	return response, responseBody
}

func assertResponseCode(t testing.TB, want, got int) {
	t.Helper()

	if got != want {
		t.Errorf("Expected status code %d, but got %d instead", want, got)
	}
}

func assertCookiExist(t testing.TB, response *httptest.ResponseRecorder) {
	t.Helper()

	if len(response.Result().Cookies()) == 0 {
		t.Error("Should get cookie but get none")
	}
}
