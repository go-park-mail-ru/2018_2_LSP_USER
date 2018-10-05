package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2018_2_LSP_USER/user"
	"github.com/gorilla/context"
	"github.com/thedevsaddam/govalidator"
)

// PostHandlerAll creates new user
func PostHandlerAll(env *Env, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(context.Get(r, "claims"))
	if context.Get(r, "claims") != nil { // TODO FIX
		return StatusData{http.StatusConflict, map[string]string{"error": "User is alredy logged in"}}
	}
	var u user.User
	rules := govalidator.MapData{
		"username":  []string{"required", "between:4,25"},
		"email":     []string{"required", "between:4,25", "email"},
		"password":  []string{"required", "alpha_space"},
		"firstname": []string{"alpha_space", "between:4,25"},
		"lastname":  []string{"alpha_space", "between:4,25"},
	}

	opts := govalidator.Options{
		Request: r,
		Data:    &u,
		Rules:   rules,
	}
	v := govalidator.New(opts)
	if e := v.ValidateJSON(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{http.StatusBadRequest, err}
	}

	if err := u.Register(env.DB); err != nil {
		return StatusData{http.StatusConflict, map[string]string{"error": err.Error()}}
	}

	setAuthCookies(w, u.Token)
	return StatusData{http.StatusOK, map[string]string{"token": u.Token}}
}

// GetHandlerAll returns all users
func GetHandlerAll(env *Env, w http.ResponseWriter, r *http.Request) error {
	payload := struct {
		Page    int
		Fields  string
		OrderBy string
	}{}
	rules := govalidator.MapData{
		"page":    []string{"required", "numeric"},
		"fields":  []string{"required", "fields:username,email,firstname,lastname,rating,id,avatar"},
		"orderby": []string{"required", "in:id,username,email,firstname,lastname,rating"},
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
	payload.Page, _ = strconv.Atoi(r.URL.Query()["page"][0])
	payload.OrderBy = r.URL.Query()["orderby"][0]

	users, err := user.GetAll(env.DB, payload.Page, payload.OrderBy)
	if err != nil {
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}

	answer := []map[string]interface{}{}
	fieldsToReturn := strings.Split(payload.Fields, ",")
	for _, u := range users {
		answer = append(answer, extractFields(u, fieldsToReturn))
	}
	return StatusData{http.StatusOK, answer}
}
