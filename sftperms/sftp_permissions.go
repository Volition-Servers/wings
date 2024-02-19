package sftperms

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"github.com/apex/log"
	"github.com/pkg/sftp"
	"github.com/pterodactyl/wings/config"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type HideFilesStruct struct {
	User  string
	Admin string
	Egg   string
}

type FilesPermissions struct {
	User      []string
	Admin     []string
	Egg       []string
	HideFiles HideFilesStruct
}

type ApiResponse struct {
	Access        bool     `json:"access"`
	AcceptedFiles []string `json:"acceptedFiles"`
	UserAccess    []string `json:"userAccess"`
	AdminAccess   []string `json:"adminAccess"`
	EggAccess     []string `json:"eggAccess"`
	HiddenFiles   []string `json:"hiddenFiles"`
}

func HasAccess(file interface{}, permissions FilesPermissions) ([]string, []string, error) {
	var User = b64.StdEncoding.EncodeToString([]byte(strings.Join(permissions.User, ",")))
	var Admin = b64.StdEncoding.EncodeToString([]byte(strings.Join(permissions.Admin, ",")))
	var Egg = b64.StdEncoding.EncodeToString([]byte(strings.Join(permissions.Egg, ",")))

	var requestBody = map[string]string{"admin": Admin, "deny": User, "egg": Egg}

	switch value := file.(type) {
	case string:
		requestBody["file"] = value
	case []string:
		requestBody["files"] = b64.StdEncoding.EncodeToString([]byte(strings.Join(value, ",")))
	}

	requestBody["hideFiles"] = permissions.HideFiles.User + "|" + permissions.HideFiles.Admin + "|" + permissions.HideFiles.Egg

	jsonBody, _ := json.Marshal(requestBody)

	var client = &http.Client{}
	req, _ := http.NewRequest("POST", "https://addons.minerpl.xyz/sftperms/validate?includes=hiddenFiles", bytes.NewBuffer(jsonBody))
	req.Header.Set("panel", config.Get().PanelLocation)
	req.Header.Set("placeholders", "BuiltByBit Volition (487001) v119315 (1708356989)")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := client.Do(req)
	if err != nil {
		log.Error("Error while validating permissions: " + err.Error())
		if requestBody["file"] != "" {
			return nil, nil, nil
		} else {
			return file.([]string), nil, nil
		}
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, nil
	}

	var parsed ApiResponse
	json.Unmarshal(body, &parsed)

	if parsed.AcceptedFiles != nil {
		return parsed.AcceptedFiles, parsed.HiddenFiles, nil
	} else {
		access := parsed.Access
		if !access {
			return nil, nil, sftp.ErrSshFxPermissionDenied
		} else {
			return nil, nil, nil
		}
	}
}

func Includes(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func HideFiles(request *sftp.Request, files []os.FileInfo, permissions FilesPermissions) []os.FileInfo {
	var fileNames []string

	dirname := strings.Replace(path.Join(request.Filepath)+"/", "//", "/", -1)

	for _, file := range files {
		fileName := path.Join(dirname, file.Name())
		if file.IsDir() {
			fileName += "/"
		}
		fileNames = append(fileNames, fileName)
	}

	acceptedFiles, hiddenFiles, err := HasAccess(fileNames, permissions)
	if err != nil {
		return nil
	}

	for i, file := range acceptedFiles {
		if !Includes(hiddenFiles, file) {
			file = path.Clean(file)

			if strings.HasPrefix(file, dirname) {
				file = file[len(dirname):]
			}

			acceptedFiles[i] = file
		}
	}

	var filteredFiles []os.FileInfo
	for _, file := range files {
		for _, acceptedFile := range acceptedFiles {
			if file.Name() == acceptedFile {
				filteredFiles = append(filteredFiles, file)
				break
			}
		}
	}

	return filteredFiles
}
