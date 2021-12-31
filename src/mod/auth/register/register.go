package register

/*
	Register Module
	author: tobychui

	Register interface handler
*/

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/valyala/fasttemplate"
	auth "imuslab.com/arozos/mod/auth"
	db "imuslab.com/arozos/mod/database"
	permission "imuslab.com/arozos/mod/permission"
)

type RegisterOptions struct {
	Hostname   string
	VendorIcon string
}

type RegisterHandler struct {
	database          *db.Database
	authAgent         *auth.AuthAgent
	permissionHandler *permission.PermissionHandler
	options           RegisterOptions
	DefaultUserGroup  string
	AllowRegistry     bool
}

func NewRegisterHandler(database *db.Database, authAgent *auth.AuthAgent, ph *permission.PermissionHandler, options RegisterOptions) *RegisterHandler {
	//Create the database for registration
	database.NewTable("register")

	//Check if the default group has been set. If not a new usergroup
	defaultUserGroup := ""
	if database.KeyExists("register", "defaultGroup") {
		//Use the configured default group
		database.Read("register", "defaultGroup", &defaultUserGroup)

		//Check the group exists
		if !ph.GroupExists(defaultUserGroup) {
			//Group not exists. Create default group.
			if !ph.GroupExists("default") {
				createDefaultGroup(ph)
			}
			defaultUserGroup = "default"
		}
	} else {
		//Default group not set or not exists. Create a new default group
		if !ph.GroupExists("default") {
			createDefaultGroup(ph)
		}
		defaultUserGroup = "default"

	}

	return &RegisterHandler{
		database:          database,
		options:           options,
		permissionHandler: ph,
		authAgent:         authAgent,
		DefaultUserGroup:  defaultUserGroup,
		AllowRegistry:     true,
	}
}

//Create the default usergroup used by new users
func createDefaultGroup(ph *permission.PermissionHandler) {
	//Default storage space: 15GB
	ph.NewPermissionGroup("default", false, 15<<30, []string{}, "Desktop")
}

func (h *RegisterHandler) HandleRegisterCheck(w http.ResponseWriter, r *http.Request) {
	if h.AllowRegistry {
		sendJSONResponse(w, "true")
	} else {
		sendJSONResponse(w, "false")
	}
}

//Handle and serve the register itnerface
func (h *RegisterHandler) HandleRegisterInterface(w http.ResponseWriter, r *http.Request) {
	//Serve the register interface
	if h.AllowRegistry {
		template, err := ioutil.ReadFile("system/auth/register.system")
		if err != nil {
			log.Println("Template not found: system/auth/register.system")
			http.NotFound(w, r)
			return
		}

		//Load the vendor icon as base64
		imagecontent, _ := readImageFileAsBase64(h.options.VendorIcon)

		//Apply templates
		t := fasttemplate.New(string(template), "{{", "}}")
		s := t.ExecuteString(map[string]interface{}{
			"host_name":   h.options.Hostname,
			"vendor_logo": imagecontent,
		})

		w.Write([]byte(s))
	} else {
		//Registry is closed
		http.NotFound(w, r)
	}
}

func readImageFileAsBase64(src string) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(content)
	return encoded, nil
}

//Get the default usergroup for this register handler
func (h *RegisterHandler) GetDefaultUserGroup() string {
	return h.DefaultUserGroup
}

//Set the default usergroup for this register handler
func (h *RegisterHandler) SetDefaultUserGroup(groupname string) error {
	if !h.permissionHandler.GroupExists(groupname) {
		return errors.New("Group not exists")
	}

	//Update the default registry in struct
	h.DefaultUserGroup = groupname

	//Write change to database
	h.database.Write("register", "defaultGroup", groupname)

	return nil
}

//Toggle registry on the fly
func (h *RegisterHandler) SetAllowRegistry(allow bool) {
	h.AllowRegistry = allow
}

//Clearn Register information by removing all users info whose account is no longer registered
func (h *RegisterHandler) CleanRegisters() {
	entries, _ := h.database.ListTable("register")
	for _, keypairs := range entries {
		if strings.Contains(string(keypairs[0]), "user/email/") {
			c := strings.Split(string(keypairs[0]), "/")
			//Get username and emails
			username := c[len(c)-1]
			if !h.authAgent.UserExists(username) {
				//Delete this record
				h.database.Delete("register", string(keypairs[0]))
			}

		}
	}
}

func (h *RegisterHandler) ListAllUserEmails() [][]interface{} {
	results := [][]interface{}{}
	entries, _ := h.database.ListTable("register")
	for _, keypairs := range entries {
		if strings.Contains(string(keypairs[0]), "user/email/") {
			c := strings.Split(string(keypairs[0]), "/")
			//Get username and emails
			username := c[len(c)-1]
			email := ""
			json.Unmarshal(keypairs[1], &email)

			//Check if the user still registered in the system
			userStillRegistered := h.authAgent.UserExists(username)

			results = append(results, []interface{}{username, email, userStillRegistered})
		}
	}

	return results
}

//Handle the request for creating a new user
func (h *RegisterHandler) HandleRegisterRequest(w http.ResponseWriter, r *http.Request) {
	if h.AllowRegistry == false {
		sendErrorResponse(w, "Public account registry is currently closed")
		return
	}
	//Get input paramter
	email, err := mv(r, "email", true)
	if err != nil {
		sendErrorResponse(w, "Invalid Email")
		return
	}

	username, err := mv(r, "username", true)
	if username == "" || strings.TrimSpace(username) == "" || err != nil {
		sendErrorResponse(w, "Invalid Username")
		return
	}

	password, err := mv(r, "password", true)
	if password == "" || err != nil {
		sendErrorResponse(w, "Invalid Password")
		return
	}

	//Check if password too short
	if len(password) < 8 {
		sendErrorResponse(w, "Password too short. Must be at least 8 characters.")
		return
	}

	//Check if the username is too short
	if len(username) < 2 {
		sendErrorResponse(w, "Username too short. Must be at least 2 characters.")
		return
	}

	//Check if the user already exists
	if h.authAgent.UserExists(username) {
		sendErrorResponse(w, "This username has already been used")
		return
	}

	//Get the default user group for public registration
	defaultGroup := h.DefaultUserGroup
	if h.permissionHandler.GroupExists(defaultGroup) == false {
		//Public registry user group not exists. Raise 500 Error
		log.Println("[CRITICAL] PUBLIC REGISTRY USER GROUP NOT FOUND! PLEASE RESTART YOUR SYSTEM!")
		sendErrorResponse(w, "Internal Server Error")
		return
	}

	//OK. Record this user to the system
	err = h.authAgent.CreateUserAccount(username, password, []string{defaultGroup})
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Write email to database as well
	h.database.Write("register", "user/email/"+username, email)

	sendOK(w)
	log.Println("New User Registered: ", email, username, strings.Repeat("*", len(password)))

}
