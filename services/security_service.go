package services

import (
	"fmt"
	"os"
)

func RemoveFiles() {
	err := os.Remove("archivo.cer")
	if err != nil {
		fmt.Println("Error al eliminar el archivo archivo.cer:", err)
	}

	err = os.Remove("archivo.key")
	if err != nil {
		fmt.Println("Error al eliminar el archivo archivo.key:", err)
	}
}
