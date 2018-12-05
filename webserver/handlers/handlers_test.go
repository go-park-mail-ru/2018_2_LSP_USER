package handlers

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

var baseURL = "http://jackal.online"

func getCreateNewUserHandlerBody() map[string]interface{} {
	choice := rand.Intn(6)
	res := make(map[string]interface{})
	switch choice {
	case 0: // empty
	case 1: // one field only
		switch rand.Intn(2) {
		case 0:
			res["username"] = randStringRunes(rand.Intn(10))
		case 1:
			res["password"] = randStringRunes(rand.Intn(10))
		case 2:
			res["email"] = randStringRunes(rand.Intn(10))
		}
	case 2, 3: // all used fields
		res["username"] = randStringRunes(rand.Intn(10))
		res["password"] = randStringRunes(rand.Intn(10))
		res["email"] = randStringRunes(rand.Intn(10))
	case 4, 5:
		res["username"] = randStringRunes(rand.Intn(10))
		res["password"] = randStringRunes(rand.Intn(10))
		res["firstname"] = randStringRunes(rand.Intn(10))
		res["lastname"] = randStringRunes(rand.Intn(10))
	}
	return res
}

func randomChooseFromSlice(src []string) (res []string) {
	for _, s := range src {
		if rand.Intn(1) == 0 {
			res = append(res, s)
		}
	}
	return res
}

func getGetManyUsersHandlerBody() map[string]interface{} {
	choice := rand.Intn(4)
	res := make(map[string]interface{})
	fields := []string{"username", "email", "firstname", "lastname", "id", "avatar", "totalscore", "totalgames"}
	switch choice {
	case 0: // empty
	case 1: // one field only
		switch rand.Intn(2) {
		case 0:
			res["page"] = strconv.Itoa(rand.Intn(10))
		case 1:
			res["fields"] = strings.Join(randomChooseFromSlice(fields), ",")
		case 2:
			res["orderby"] = fields[rand.Intn(len(fields))]
		}
	case 2, 3: // all used fields
		res["page"] = strconv.Itoa(rand.Intn(10))
		res["fields"] = strings.Join(randomChooseFromSlice(fields), ",")
		res["orderby"] = fields[rand.Intn(len(fields))]
	}
	return res
}

func BenchmarkCreateNewUserHandler(b *testing.B) {
	client := &http.Client{}
	for i := 0; i < b.N; i++ {
		jsonBuff, _ := json.Marshal(getCreateNewUserHandlerBody())
		req, _ := http.NewRequest("POST", baseURL+"/users", bytes.NewBuffer(jsonBuff))
		req.Header.Set("Content-Type", "application/json")
		client.Do(req)
	}
}

func BenchmarkGetManyUsersHandler(b *testing.B) {
	client := &http.Client{}
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", baseURL+"/api/users", nil)
		parameters := getGetManyUsersHandlerBody()
		query := req.URL.Query()
		for p := range parameters {
			query.Add(p, parameters[p].(string))
		}
		req.URL.RawQuery = query.Encode()
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		b.Log(res.StatusCode)
		if err != nil {
			b.Error("Error during request: ", err)
		}
	}
}
