package utils

import (
	"encoding/base64"
	"fmt"
	"os"
)

func DecodeBase64AndSaveIntoCustomFile(path string, data string) error {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, decodedData, 0644)
	if err != nil {
		fmt.Println("Error al crear el archivo:", err)
		return err
	} else {
		fmt.Println("Se ha creado el archivo:", path)
	}

	return nil
}
