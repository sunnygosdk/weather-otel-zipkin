package config

import "errors"

var (
	ErrZipCodeInvalido      = errors.New("zipcode invalido")
	ErrZipCodeNaoEncontrado = errors.New("zipcode nao encontrado")
)
