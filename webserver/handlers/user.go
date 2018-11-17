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

func PostHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "User id should be unsigned integer"}}
	}

	claims := context.Get(r, "claims").(jwt.MapClaims)
	if int(claims["id"].(float64)) != id {
		env.Logger.Infow("Not enough permssions",
			"user", int(claims["id"].(float64)),
			"requested_user", id,
		)
		return StatusData{http.StatusForbidden, map[string]string{"error": "Not enought permissions"}}
	}

	rules := govalidator.MapData{
		"file:file": []string{"required", "ext:jpg,png", "size:300000", "mime:image/jpg,image/png"},
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

	file, handle, err := r.FormFile("file")
	if err != nil {
		env.Logger.Infow("File read error",
			"error", err.Error(),
		)
		return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
	}
	defer file.Close()

	var u user.User
	u.ID = id

	err = saveFile(file, handle, u.ID)
	if err != nil {
		env.Logger.Infow("File save error",
			"error", err.Error(),
		)
		return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
	}
	response := map[string]string{"URL": "/avatars/" + strconv.Itoa(u.ID) + "_" + handle.Filename}
	err = u.UpdateOne(env.DB, map[string]string{"avatar": response["URL"]})
	if err != nil {
		env.Logger.Infow("User update error",
			"error", err.Error(),
		)
		return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
	}

	return StatusData{http.StatusOK, response}
}

func PutHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "User id should be unsigned integer"}}
	}

	claims := context.Get(r, "claims").(jwt.MapClaims)
	if int(claims["id"].(float64)) != id {
		env.Logger.Infow("Not enough permssions",
			"user", int(claims["id"].(float64)),
			"requested_user", id,
		)
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
		"fields":      []string{"fields:username,email,firstname,lastname,rating,id,avatar", "required"},
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
			env.Logger.Infow("User validate password error",
				"error", err.Error(),
			)
			return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
		}
		if !isValid {
			return StatusData{http.StatusBadRequest, map[string]string{"error": "Wrong old password"}}
		}
		data["password"], err = user.HashPassword(payload.Password)
		if err != nil {
			env.Logger.Infow("User hash password error",
				"error", err.Error(),
			)
			return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
		}
	}

	if len(data) == 0 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "Empty request"}}
	}

	u := user.User{}
	u.ID = id
	if err = u.UpdateOne(env.DB, data); err != nil {
		env.Logger.Infow("Can't update user",
			"error", err.Error(),
			"data", data,
		)
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
		"fields": []string{"required", "fields:username,email,firstname,lastname,rating,id,avatar,totalscore,totalgames"},
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
		env.Logger.Infow("Can't get one user",
			"id", u.ID,
			"error", err.Error(),
		)
		return StatusData{http.StatusBadRequest, map[string]string{"error": err.Error()}}
	}

	fieldsToReturn := strings.Split(payload.Fields, ",")
	answer := extractFields(u, fieldsToReturn)

	return StatusData{http.StatusOK, answer}
}
