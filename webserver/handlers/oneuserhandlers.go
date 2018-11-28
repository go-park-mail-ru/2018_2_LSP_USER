package handlers

import (
	cnt "context"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	"github.com/gorilla/context"
	"github.com/thedevsaddam/govalidator"
)

// UpdateAvatarHandler updates user avatar
func UpdateAvatarHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	ID, err := parseIDFromURL(r)
	if err != nil {
		return err
	}

	claims := context.Get(r, "claims").(map[string]interface{})
	if err := checkPermissions(env, ID, claims); err != nil {
		return err
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   avatarUpdateRules,
	}
	v := govalidator.New(opts)
	if e := v.Validate(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{
			Code: http.StatusBadRequest,
			Data: err,
		}
	}

	file, handle, err := r.FormFile("file")
	if err != nil {
		env.Logger.Infow("File read error",
			"error", err.Error(),
		)
		return StatusData{
			Code: http.StatusInternalServerError,
			Data: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}
	defer file.Close()

	newFilename, err := saveFile(file, handle, ID)
	if err != nil {
		env.Logger.Infow("File save error",
			"error", err.Error(),
		)
		return StatusData{
			Code: http.StatusInternalServerError,
			Data: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	toBeUpdated := user_proto.User{
		ID:     int64(ID),
		Avatar: newFilename,
	}

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	_, err = userManager.Update(ctx, &toBeUpdated)
	if err := handleUpdateUserGrpcError(env, err); err != nil {
		return err
	}

	return nil
}

// UpdateUserHandler updates user
func UpdateUserHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	ID, err := parseIDFromURL(r)
	if err != nil {
		return err
	}

	claims := context.Get(r, "claims").(map[string]interface{})
	if err := checkPermissions(env, ID, claims); err != nil {
		return err
	}

	payload := updateUserPayload{}
	opts := govalidator.Options{
		Request: r,
		Data:    &payload,
		Rules:   updateValidationRules,
	}
	v := govalidator.New(opts)
	if e := v.ValidateJSON(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{
			Code: http.StatusBadRequest,
			Data: err,
		}
	}

	toBeUpdated := user_proto.User{ID: int64(ID)}
	if len(payload.FirstName) > 0 {
		toBeUpdated.FirstName = payload.FirstName
	}
	if len(payload.LastName) > 0 {
		toBeUpdated.LastName = payload.LastName
	}

	if err := checkPasswordUpdate(env, payload, claims); err != nil {
		return err
	}
	toBeUpdated.Password = payload.Password

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	_, err = userManager.Update(ctx, &toBeUpdated)
	if err := handleUpdateUserGrpcError(env, err); err != nil {
		return err
	}

	return nil
}

// GetOneUserHandler returns user
func GetOneUserHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	ID, err := parseIDFromURL(r)
	if err != nil {
		return err
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   getOneRules,
	}
	v := govalidator.New(opts)
	if e := v.Validate(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{
			Code: http.StatusBadRequest,
			Data: err,
		}
	}
	payload := getOneUserPayload{r.URL.Query()["fields"][0]}

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	u, err := userManager.GetOne(ctx,
		&user_proto.UserID{
			ID: int64(ID),
		})

	if err := handleGetOneUserGrpcError(env, err); err != nil {
		return err
	}

	fieldsToReturn := strings.Split(payload.Fields, ",")
	answer := extractFields(u, fieldsToReturn)

	return StatusData{
		Code: http.StatusOK,
		Data: answer,
	}
}
