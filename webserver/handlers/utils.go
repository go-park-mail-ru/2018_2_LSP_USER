package handlers

import (
	cnt "context"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	"golang.org/x/crypto/bcrypt"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func extractFields(u *user_proto.User, fieldsToReturn []string) map[string]interface{} {
	answer := map[string]interface{}{}
	for _, f := range fieldsToReturn {
		switch f {
		case "id":
			answer["id"] = u.ID
		case "firstname":
			answer["firstname"] = u.FirstName
		case "lastname":
			answer["lastname"] = u.LastName
		case "email":
			answer["email"] = u.Email
		case "username":
			answer["username"] = u.Username
		case "avatar":
			answer["avatar"] = u.Avatar
		case "totalscore":
			answer["totalscore"] = u.TotalScore
		case "totalgames":
			answer["totalgames"] = u.TotalGames
		}
	}
	return answer
}

func setAuthCookies(w http.ResponseWriter, tokenString string) {
	firstDot := strings.Index(tokenString, ".") + 1
	secondDot := strings.Index(tokenString[firstDot:], ".") + firstDot
	cookieHeaderPayload := http.Cookie{
		Name:    "header.payload",
		Value:   tokenString[:secondDot],
		Expires: time.Now().Add(30 * time.Minute),
		Secure:  true,
		Domain:  ".jackal.online",
	}
	cookieSignature := http.Cookie{
		Name:     "signature",
		Value:    tokenString[secondDot+1:],
		Expires:  time.Now().Add(720 * time.Hour),
		Secure:   true,
		HttpOnly: true,
		Domain:   ".jackal.online",
	}
	http.SetCookie(w, &cookieHeaderPayload)
	http.SetCookie(w, &cookieSignature)
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func saveFile(file multipart.File, handle *multipart.FileHeader, id int) (string, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	newFilename := os.Getenv("AVATARS_PATH") + strconv.Itoa(id) + "_" + handle.Filename
	if err := ioutil.WriteFile(newFilename, data, 0666); err != nil {
		return "", err
	}

	return newFilename, nil
}

func checkPermissions(env *Env, ID int, claims map[string]interface{}) error {
	if int(claims["id"].(float64)) != ID {
		env.Logger.Infow("Not enough permssions for user change",
			"user", int(claims["id"].(float64)),
		)
		return StatusData{
			Code: http.StatusForbidden,
			Data: map[string]string{
				"error": "Not enought permissions",
			},
		}
	}
	return nil
}

func parseIDFromURL(r *http.Request) (int, error) {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	ID, err := strconv.Atoi(idStr)
	if err != nil || ID < 0 {
		return 0, StatusData{
			Code: http.StatusBadRequest,
			Data: map[string]string{
				"error": "User id should be unsigned integer",
			},
		}
	}
	return ID, nil
}

func checkPasswordUpdate(env *Env, payload updateUserPayload, claims map[string]interface{}) error {
	if len(payload.Password) > 0 {
		if len(payload.OldPassword) == 0 {
			env.Logger.Infow("Requested password change without old password",
				"user", int(claims["id"].(int)),
			)
			return StatusData{
				Code: http.StatusBadRequest,
				Data: map[string]string{
					"error": "Please, specify old password",
				},
			}
		}
		ctx := cnt.Background()
		userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
		u, err := userManager.GetOne(ctx,
			&user_proto.UserID{
				ID: int64(claims["id"].(int)),
			})

		if err := handleGetOneUserGrpcError(env, err); err != nil {
			return err
		}

		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(payload.OldPassword))
		if err != nil {
			return StatusData{
				Code: http.StatusBadRequest,
				Data: map[string]string{
					"error": "Wrong old user password",
				},
			}
		}
	}
	return nil
}
