package handlers

import "github.com/thedevsaddam/govalidator"

var updateValidationRules = govalidator.MapData{
	"firstname":   []string{"between:4,25"},
	"lastname":    []string{"between:4,25"},
	"password":    []string{"alpha_space"},
	"oldpassword": []string{"alpha_space"},
}

var getOneRules = govalidator.MapData{
	"fields": []string{"required", "fields:username,email,firstname,lastname,rating,id,avatar,totalscore,totalgames"},
}

var createNewUserRules = govalidator.MapData{
	"username":  []string{"required", "between:4,25"},
	"email":     []string{"required", "between:4,25", "email"},
	"password":  []string{"required", "alpha_space"},
	"firstname": []string{"alpha_space", "between:4,25"},
	"lastname":  []string{"alpha_space", "between:4,25"},
}

var getManyUsersRules = govalidator.MapData{
	"page":    []string{"required", "numeric"},
	"fields":  []string{"required", "fields:username,email,firstname,lastname,id,avatar,totalscore,totalgames"},
	"orderby": []string{"required", "in:id,username,email,firstname,lastname,totalscore,totalgames"},
}

var avatarUpdateRules = govalidator.MapData{
	"file:file": []string{"required", "ext:jpg,png", "size:300000", "mime:image/jpg,image/png"},
}
