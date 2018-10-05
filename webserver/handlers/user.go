package handlers

import (
	"net/http"
	"strconv"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-park-mail-ru/2018_2_LSP_USER/user"
	"github.com/gorilla/context"
	"github.com/thedevsaddam/govalidator"
)

func PutHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "User id should be unsigned integer"}}
	}

	claims := context.Get(r, "claims").(jwt.MapClaims)
	if int(claims["id"].(float64)) != id {
		return StatusData{http.StatusForbidden, map[string]string{"error": "Not enought permissions"}}
	}

	payload := struct {
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Password    string `json:"password"`
		OldPassword string `json:"oldpassword"`
		Fields      string `json:"fields"`
	}{}

	rules := govalidator.MapData{
		"firstname":   []string{"between:4,25"},
		"lastname":    []string{"between:4,25"},
		"password":    []string{"alpha_space"},
		"oldpassword": []string{"alpha_space"},
		"fields":      []string{"fields:username,email,firstname,lastname,rating", "required"},
	}

	opts := govalidator.Options{
		Request: r,
		Data:    &payload,
		Rules:   rules,
	}
	v := govalidator.New(opts)
	if e := v.ValidateJSON(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{http.StatusBadRequest, err}
	}
	data := make(map[string]string)
	if len(payload.FirstName) > 0 {
		data["firstname"] = payload.FirstName
	}
	if len(payload.LastName) > 0 {
		data["lastname"] = payload.LastName
	}

	if len(payload.Password) > 0 {
		if len(payload.OldPassword) == 0 {
			return StatusData{http.StatusBadRequest, map[string]string{"error": "Please, specify old password"}}
		}
		isValid, err := user.ValidateUserPassword(env.DB, payload.OldPassword, id)
		if err != nil {
			return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
		}
		if !isValid {
			return StatusData{http.StatusBadRequest, map[string]string{"error": "Wrong old password"}}
		}
		data["password"], _ = user.HashPassword(payload.Password) // TODO error
	}

	if len(data) == 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "Empty request"}}
	}

	u := user.User{}
	u.ID = id
	if err = u.UpdateOne(env.DB, data); err != nil {
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}

	fieldsToReturn := strings.Split(payload.Fields, ",")
	answer := extractFields(u, fieldsToReturn)

	return StatusData{http.StatusOK, answer}
}

func GetHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	var u user.User
	var err error
	u.ID, err = strconv.Atoi(idStr)
	if err != nil || u.ID < 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "User id should be unsigned integer"}}
	}

	payload := struct {
		Fields string
	}{}

	rules := govalidator.MapData{
		"fields": []string{"required", "fields:username,email,firstname,lastname,rating"},
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   rules,
	}
	v := govalidator.New(opts)
	if e := v.Validate(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{http.StatusBadRequest, err}
	}
	payload.Fields = r.URL.Query()["fields"][0]

	u, err = user.GetOne(env.DB, u.ID)
	if err != nil {
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}

	fieldsToReturn := strings.Split(payload.Fields, ",")
	answer := extractFields(u, fieldsToReturn)

	return StatusData{http.StatusOK, answer}
}