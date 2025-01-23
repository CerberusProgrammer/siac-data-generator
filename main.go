package main

import (
	"flag"
	"fmt"
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
		remotePath := fmt.Sprintf("%s/%s", *smbPath, basePath)

		cmd := exec.Command("cmd", "/C", fmt.Sprintf("mkdir %s", remotePath))
		cmd.Env = append(os.Environ(), fmt.Sprintf("USER=%s", *smbUser), fmt.Sprintf("PASS=%s", *smbPass))
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al crear las carpetas en el servidor SMB:", err)
			services.RemoveFiles()
			return
		}

		cmd = exec.Command("cmd", "/C", fmt.Sprintf("copy archivo.cer %s", remotePath))
		cmd.Env = append(os.Environ(), fmt.Sprintf("USER=%s", *smbUser), fmt.Sprintf("PASS=%s", *smbPass))
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al subir el archivo .cer al servidor SMB:", err)
			services.RemoveFiles()
			return
		}

		cmd = exec.Command("cmd", "/C", fmt.Sprintf("copy archivo.key %s", remotePath))
		cmd.Env = append(os.Environ(), fmt.Sprintf("USER=%s", *smbUser), fmt.Sprintf("PASS=%s", *smbPass))
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al subir el archivo .key al servidor SMB:", err)
			services.RemoveFiles()
			return
		}

		fmt.Println("Archivos subidos exitosamente al servidor SMB")
		services.RemoveFiles()
	} else {
		fmt.Println("No data found")
	}
}
