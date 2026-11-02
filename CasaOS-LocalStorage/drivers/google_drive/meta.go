package google_drive

import (
	"github.com/IceWhaleTech/CasaOS-LocalStorage/internal/driver"
	"github.com/IceWhaleTech/CasaOS-LocalStorage/internal/op"
)

const ICONURL = "./img/driver/GoogleDrive.svg"
const CLIENTID = ""
const CLIENTSECRET = ""

type Addition struct {
	driver.RootID
	RefreshToken   string `json:"refresh_token" required:"true" omit:"true"`
	OrderBy        string `json:"order_by" type:"string" help:"such as: folder,name,modifiedTime" omit:"true"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" omit:"true"`
	ClientID       string `json:"client_id" required:"true" default:"" omit:"true"`
	ClientSecret   string `json:"client_secret" required:"true" default:"" omit:"true"`
	ChunkSize      int64  `json:"chunk_size" type:"number" help:"chunk size while uploading (unit: MB)" omit:"true"`
	AuthUrl        string `json:"auth_url" type:"string" default:"https://accounts.google.com/o/oauth2/auth/oauthchooseaccount?response_type=code&client_id=&redirect_uri=https%3A%2F%2Fcloudoauth.files.casaos.app&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive&access_type=offline&approval_prompt=force&state=${HOST}%2Fv1%2Frecover%2FGoogleDrive&service=lso&o2v=1&flowName=GeneralOAuthFlow"`
	Icon           string `json:"icon" type:"string" default:"./img/driver/GoogleDrive.svg"`
	Code           string `json:"code" type:"string" help:"code from auth_url" omit:"true"`
}

var config = driver.Config{
	Name:        "GoogleDrive",
	OnlyProxy:   true,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &GoogleDrive{}
	})
}
