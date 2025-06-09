package oauth

import "github.com/go-playground/validator/v10"

// Global validator instance shared across all OAuth files
var validate = validator.New()
