package models

type DatosFiscales struct {
	RFC           string `json:"RFC"`
	NombreFiscal  string `json:"NombreFiscal"`
	Certificado   string `json:"Certificado"`
	LlavePrivada  string `json:"LlavePrivada"`
	Contraseña    string `json:"Contraseña"`
	EmpresaActiva bool   `json:"EmpresaActiva"`
	Activo        bool   `json:"Activo"`
}
