package services

import (
	"encoding/json"
	"io"
	"net/http"
	"siac/models"
)

func GetFiscalData(url, token, rfc, aplicacion string) ([]models.DatosFiscales, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Add("RFC", rfc)
	req.Header.Add("Aplicacion", aplicacion)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response []models.DatosFiscales
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
