package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2018_2_LSP_USER/user"
	"github.com/thedevsaddam/govalidator"
)

func init() {
	govalidator.AddCustomRule("fields", func(field string, rule string, message string, value interface{}) error {
		if value == nil {
			return nil
		}
		fields := strings.Split(value.(string), ",")
		if len(fields) == 0 {
			return errors.New("Field keyword should be field list divided by comma. Available fields: " + strings.TrimPrefix(rule, "fields:"))
		}
		fieldListStr := strings.TrimPrefix(rule, "fields:")
		fieldListSlice := strings.Split(fieldListStr, ",")
		for _, f := range fields {
			if !contains(fieldListSlice, f) {
				return errors.New("Field keyword should be field list divided by comma. Available fields: " + strings.TrimPrefix(rule, "fields:"))
			}
		}
		return nil
	})
}

func PostHandlerAll(env *Env, w http.ResponseWriter, r *http.Request) error {
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

func GetHandlerAll(env *Env, w http.ResponseWriter, r *http.Request) error {
	payload := struct {
		Page    int
		Fields  string
		OrderBy string
	}{}

	rules := govalidator.MapData{
		"page":    []string{"required", "numeric"},
		"fields":  []string{"required", "fields:username,email,firstname,lastname,rating"},
		"orderby": []string{"in:username,email,firstname,lastname,rating"},
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
	if keys, ok := r.URL.Query()["orderby"]; ok {
		payload.OrderBy = keys[0]
	}

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

func PutHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	var u user.User
	var err error
	u.ID, err = strconv.Atoi(idStr)
	if err != nil || u.ID < 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "User id should be unsigned integer"}}
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
	fmt.Println(payload)
	data := make(map[string]string)
	if len(payload.FirstName) > 0 {
		fmt.Println("fffff")
		data["firstname"] = u.FirstName
	}
	if len(payload.LastName) > 0 {
		fmt.Println("lllll")
		data["lastname"] = u.LastName
	}

	if len(payload.Password) > 0 {
		if len(payload.OldPassword) == 0 {
			return StatusData{http.StatusBadRequest, map[string]string{"error": "Please, specify old password"}}
		}
		isValid, err := user.ValidateUserPassword(env.DB, payload.OldPassword, u.ID)
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

	request := "UPDATE users SET "

	for k, v := range data {
		request += k + "='" + v + "',"
	}
	request = request[:len(request)-1]
	request += " WHERE id = $1 RETURNING username, email, firstname, lastname, rating"
	fmt.Println(request)
	rows, err := env.DB.Query(request, u.ID)
	if err != nil {
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}
	rows.Next()
	var firstname sql.NullString
	var lastname sql.NullString
	err = rows.Scan(&u.Username, &u.Email, &firstname, &lastname, &u.Rating)
	if err != nil {
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}
	if firstname.Valid {
		temp, _ := firstname.Value()
		u.FirstName = temp.(string)
	} else {
		u.FirstName = ""
	}
	if lastname.Valid {
		temp, _ := lastname.Value()
		u.LastName = temp.(string)
	} else {
		u.LastName = ""
	}
	if err != nil {
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

func extractFields(u user.User, fieldsToReturn []string) map[string]interface{} {
	answer := map[string]interface{}{}
	for _, f := range fieldsToReturn {
		switch f {
		case "firstname":
			answer["firstname"] = u.FirstName
		case "lastname":
			answer["lastname"] = u.LastName
		case "email":
			answer["email"] = u.Email
		case "username":
			answer["username"] = u.Username
		case "rating":
			answer["rating"] = u.Rating
		}
	}
	return answer
}
