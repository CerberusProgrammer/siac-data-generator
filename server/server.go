package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"siac/services"
	"siac/utils"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	url := r.Header.Get("url")
	rfc := r.Header.Get("rfc")
	token := r.Header.Get("token")
	aplicacion := r.Header.Get("aplicacion")
	cltid := r.Header.Get("cltid")
	perid := r.Header.Get("perid")
	smbUser := r.Header.Get("smbUser")
	smbPass := r.Header.Get("smbPass")
	smbPath := r.Header.Get("smbPath")

	if !utils.ValidParams(url, token, rfc, aplicacion, cltid, perid) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	response, err := services.GetFiscalData(url, token, rfc, aplicacion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving fiscal data: %v", err), http.StatusInternalServerError)
		return
	}

	if len(response) > 0 {
		certificado := response[0].Certificado
		llavePrivada := response[0].LlavePrivada

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.cer", certificado)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving certificate: %v", err), http.StatusInternalServerError)
			return
		}

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.key", llavePrivada)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving private key: %v", err), http.StatusInternalServerError)
			return
		}

		basePath := filepath.Join("cfdi", cltid, perid, "generales")
		pathParts := []string{"cfdi", cltid, perid, "generales"}

		authFileContent := fmt.Sprintf("username = %s\npassword = %s\n", smbUser, smbPass)
		authFile, err := os.CreateTemp("", "smb_auth_")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating auth file: %v", err), http.StatusInternalServerError)
			return
		}
		defer os.Remove(authFile.Name())

		if _, err := authFile.Write([]byte(authFileContent)); err != nil {
			http.Error(w, fmt.Sprintf("Error writing to auth file: %v", err), http.StatusInternalServerError)
			return
		}
		authFile.Close()

		currentPath := ""
		for _, part := range pathParts {
			currentPath = filepath.Join(currentPath, part)
			cmd := exec.Command("smbclient", smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("mkdir %s", currentPath))
			output, err := cmd.CombinedOutput()
			if err != nil {
				services.RemoveFiles()
				http.Error(w, fmt.Sprintf("Error creating directory %s on SMB server: %s", currentPath, string(output)), http.StatusInternalServerError)
				return
			}
		}

		cmd := exec.Command("smbclient", smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.cer %s/archivo.cer", basePath))
		output, err := cmd.CombinedOutput()
		if err != nil {
			services.RemoveFiles()
			http.Error(w, fmt.Sprintf("Error uploading .cer file to SMB server: %s", string(output)), http.StatusInternalServerError)
			return
		}

		cmd = exec.Command("smbclient", smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.key %s/archivo.key", basePath))
		output, err = cmd.CombinedOutput()
		if err != nil {
			services.RemoveFiles()
			http.Error(w, fmt.Sprintf("Error uploading .key file to SMB server: %s", string(output)), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Files uploaded successfully to SMB server")
	} else {
		http.Error(w, "No data found", http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8010", nil)
	fmt.Println("Server running on port 8010")
}
