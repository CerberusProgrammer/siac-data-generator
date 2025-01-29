package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"siac/services"
	"siac/utils"
)

func main() {
	url := flag.String("url", "", "URL")
	rfc := flag.String("rfc", "", "RFC")
	token := flag.String("token", "", "Token")
	aplicacion := flag.String("aplicacion", "", "Aplicacion")
	cltid := flag.Int("cltid", 0, "CLTID")
	perid := flag.Int("perid", 0, "PERID")
	smbUser := flag.String("smbUser", "", "SMB User")
	smbPass := flag.String("smbPass", "", "SMB Password")
	smbPath := flag.String("smbPath", "", "SMB Path")

	flag.Parse()

	if !utils.ValidParams(*url, *token, *rfc, *aplicacion, fmt.Sprintf("%d", *cltid), fmt.Sprintf("%d", *perid)) {
		fmt.Println("Invalid parameters")
		return
	}

	response, err := services.GetFiscalData(*url, *token, *rfc, *aplicacion)
	if err != nil {
		fmt.Println("Error retrieving fiscal data:", err)
		return
	}

	if len(response) > 0 {
		certificado := response[0].Certificado
		llavePrivada := response[0].LlavePrivada

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.cer", certificado)
		if err != nil {
			fmt.Println("Error saving certificate:", err)
			return
		}

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.key", llavePrivada)
		if err != nil {
			fmt.Println("Error saving private key:", err)
			return
		}

		basePath := filepath.Join("cfdi", fmt.Sprintf("%d", *cltid), fmt.Sprintf("%d", *perid), "generales")
		pathParts := []string{"cfdi", fmt.Sprintf("%d", *cltid), fmt.Sprintf("%d", *perid), "generales"}

		fmt.Println("Connecting to SMB...")

		authFileContent := fmt.Sprintf("username = %s\npassword = %s\n", *smbUser, *smbPass)
		authFile, err := ioutil.TempFile("", "smb_auth_")
		if err != nil {
			fmt.Println("Error creating auth file:", err)
			return
		}
		defer os.Remove(authFile.Name())

		if _, err := authFile.Write([]byte(authFileContent)); err != nil {
			fmt.Println("Error writing to auth file:", err)
			return
		}
		authFile.Close()

		currentPath := ""
		for _, part := range pathParts {
			currentPath = filepath.Join(currentPath, part)
			cmd := exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("mkdir %s", currentPath))
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error creating directory %s on SMB server: %s\n", currentPath, string(output))
				services.RemoveFiles()
				return
			}
		}

		fmt.Println("Uploading .cer file to SMB...")
		cmd := exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.cer %s/archivo.cer", basePath))
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error uploading .cer file to SMB server: %s\n", string(output))
			services.RemoveFiles()
			return
		}

		fmt.Println("Uploading .key file to SMB...")
		cmd = exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.key %s/archivo.key", basePath))
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error uploading .key file to SMB server: %s\n", string(output))
			services.RemoveFiles()
			return
		}

		fmt.Println("Files uploaded successfully to SMB server")
	} else {
		fmt.Println("No data found")
	}
}
