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

	if !utils.ValidParams(*url, *token, *rfc, *aplicacion, *cltid, *perid) {
		fmt.Println("Parametros invalidos")
		return
	}

	response, err := services.GetFiscalData(*url, *token, *rfc, *aplicacion)
	if err != nil {
		fmt.Println("Error al obtener la informacion fisca;:", err)
		return
	}

	if len(response) > 0 {
		certificado := response[0].Certificado
		llavePrivada := response[0].LlavePrivada

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.cer", certificado)
		if err != nil {
			fmt.Println("Error al guardar el Certificado:", err)
			return
		}

		err = utils.DecodeBase64AndSaveIntoCustomFile("archivo.key", llavePrivada)
		if err != nil {
			fmt.Println("Error al guardar la LlavePrivada:", err)
			return
		}

		basePath := filepath.Join("cfdi", fmt.Sprintf("%d", *cltid), fmt.Sprintf("%d", *perid), "generales")
		pathParts := []string{"cfdi", fmt.Sprintf("%d", *cltid), fmt.Sprintf("%d", *perid), "generales"}

		fmt.Println("Conectando al SMB...")

		// Crear un archivo de autenticación temporal
		authFileContent := fmt.Sprintf("username = %s\npassword = %s\n", *smbUser, *smbPass)
		authFile, err := ioutil.TempFile("", "smb_auth_")
		if err != nil {
			fmt.Println("Error al crear el archivo de autenticación:", err)
			return
		}
		defer os.Remove(authFile.Name())

		if _, err := authFile.Write([]byte(authFileContent)); err != nil {
			fmt.Println("Error al escribir en el archivo de autenticación:", err)
			return
		}
		authFile.Close()

		// Crear las carpetas de manera secuencial
		currentPath := ""
		for _, part := range pathParts {
			currentPath = filepath.Join(currentPath, part)
			cmd := exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("mkdir %s", currentPath))
			fmt.Printf("Comando a ejecutar: %s\n", cmd.String())
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error al crear la carpeta %s en el servidor SMB: %s\n", currentPath, string(output))
				services.RemoveFiles()
				return
			}
		}

		fmt.Println("Subiendo archivo .cer al SMB...")
		cmd := exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.cer %s/archivo.cer", basePath))
		fmt.Println("Comando a ejecutar: ", cmd.String())
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error al subir el archivo .cer al servidor SMB: %s\n", string(output))
			services.RemoveFiles()
			return
		}

		fmt.Println("Subiendo archivo .key al SMB...")
		cmd = exec.Command("smbclient", *smbPath, "-A", authFile.Name(), "-c", fmt.Sprintf("put archivo.key %s/archivo.key", basePath))
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error al subir el archivo .key al servidor SMB: %s\n", string(output))
			services.RemoveFiles()
			return
		}

		fmt.Println("Archivos subidos exitosamente al servidor SMB")
	} else {
		fmt.Println("No data found")
	}
}
