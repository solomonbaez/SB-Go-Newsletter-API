package api_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func TestValidatePHC(t *testing.T) {
	password := uuid.NewString()

	phc, e := models.GeneratePHC(password)
	if e != nil {
		t.Errorf("Failed to generate PHC: %s", e)
		return
	}

	if e := models.ValidatePHC(password, phc); e != nil {
		t.Errorf("Failed to validate PHC: %s", e)
		return
	}

}
