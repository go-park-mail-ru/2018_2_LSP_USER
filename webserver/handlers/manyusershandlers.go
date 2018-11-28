package handlers

import (
	cnt "context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2018_2_LSP_AUTH_GRPC/auth_proto"
	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	"github.com/thedevsaddam/govalidator"
)

// CreateNewUserHandler creates new user
func CreateNewUserHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	var u user_proto.User
	opts := govalidator.Options{
		Request: r,
		Data:    &u,
		Rules:   createNewUserRules,
	}
	v := govalidator.New(opts)
	if e := v.ValidateJSON(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{http.StatusBadRequest, err}
	}

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	userID, err := userManager.Create(ctx, &u)
	if err := handleCreateUserGrpcError(env, err); err != nil {
		return err
	}

	authManager := auth_proto.NewAuthCheckerClient(env.GRCPAuth)
	token, err := authManager.Generate(ctx,
		&auth_proto.TokenPayload{
			ID: userID.ID,
		})

	if err != nil {
		return StatusData{
			Code: http.StatusBadRequest,
			Data: map[string]string{
				"error": err.Error(),
			},
		}
	}

	setAuthCookies(w, token.Token)
	return nil
}

// GetManyUsersHandler returns all users
func GetManyUsersHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	opts := govalidator.Options{
		Request: r,
		Rules:   getManyUsersRules,
	}
	v := govalidator.New(opts)
	if e := v.Validate(); len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return StatusData{http.StatusBadRequest, err}
	}

	payload := getManyUsersPayload{}
	payload.Fields = r.URL.Query()["fields"][0]
	payload.Page, _ = strconv.Atoi(r.URL.Query()["page"][0])
	payload.OrderBy = r.URL.Query()["orderby"][0]

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	stream, err := userManager.GetMany(ctx,
		&user_proto.ManyUsersOptions{
			Page:    int64(payload.Page),
			OrderBy: payload.OrderBy,
		})
	if err := handleGeneralGrpcError(env, err); err != nil {
		return err
	}
	defer stream.CloseSend()

	answer := []map[string]interface{}{}
	fieldsToReturn := strings.Split(payload.Fields, ",")

	for {
		u, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			env.Logger.Fatalw("Stream read error",
				"err", err.Error(),
			)
			return StatusData{
				Code: http.StatusBadRequest,
				Data: map[string]string{
					"error": err.Error(),
				},
			}
		}
		answer = append(answer, extractFields(u, fieldsToReturn))
	}
	return StatusData{http.StatusOK, answer}
}
