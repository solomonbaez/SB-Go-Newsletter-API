package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	// "html"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	"golang.org/x/crypto/argon2"
)

const invalidRunes = "{}/\\<>() "

type HashParams struct {
	saltLen    uint32
	iterations uint32
	memory     uint32
	threads    uint8
	keyLen     uint32
}

var params = HashParams{
	saltLen:    16,
	iterations: 1,
	memory:     64 * 1024,
	threads:    4,
	keyLen:     32,
}

type Credentials struct {
	Username string
	Password string
}

var baseHash string

func init() {
	randomPassword := uuid.NewString()
	baseHash, _ = GeneratePHC(randomPassword)
}

func (rh *RouteHandler) PostNewsletter(c *gin.Context, client clients.EmailClient) {
	var newsletter models.Newsletter
	var body models.Body
	var response string
	var e error

	requestID := c.GetString("requestID")
	key, _ := c.GetPostForm("idempotency_key")
	newsletter.Key = key

	body.Title, _ = c.GetPostForm("title")
	body.Text, _ = c.GetPostForm("text")
	body.Html, _ = c.GetPostForm("html")

	if e = ParseNewsletter(&body); e != nil {
		response = "Failed to parse newsletter"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}
	newsletter.Content = &body

	subscribers := rh.GetConfirmedSubscribers(c)
	for _, s := range subscribers {
		// re-parse email to ensure data integrity
		newsletter.Recipient, e = models.ParseEmail(s.Email.String())
		if e != nil {
			response = fmt.Sprintf("Invalid email: %v", s.Email.String())
			HandleError(c, requestID, e, response, http.StatusConflict)
			continue
		}
		if e = ParseNewsletter(&newsletter); e != nil {
			response = "Invalid newsletter"
			HandleError(c, requestID, e, response, http.StatusBadRequest)
			continue
		}
		if e = client.SendEmail(&newsletter); e != nil {
			response = "Failed to send newsletter"
			HandleError(c, requestID, e, response, http.StatusInternalServerError)
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{"requestID": requestID, "message": "Emails successfully delivered"})
}

// TODO Sanitize credentials again?
func (rh *RouteHandler) ValidateCredentials(c *gin.Context, credentials *Credentials) (*string, error) {
	var id string
	var password_hash string

	requestID := c.GetString("requestID")

	var user_e error
	query := "SELECT id, password_hash FROM users WHERE username=$1"
	e := rh.DB.QueryRow(c, query, credentials.Username).Scan(&id, &password_hash)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Invalid username")

		// prevent timing attacks!
		password_hash = baseHash
		user_e = e
	}

	if e := ValidatePHC(credentials.Password, password_hash); e != nil {
		if user_e != nil {
			e = user_e
		}
		return nil, e
	}

	log.Info().
		Str("requestID", requestID).
		Str("userID", id).
		Msg("Successfully validated user credentials")

	return &id, nil
}

func ParseNewsletter(c interface{}) error {
	v := reflect.ValueOf(c).Elem()
	nFields := v.NumField()

	for i := 0; i < nFields; i++ {
		field := v.Field(i)
		valid := field.IsValid() && !field.IsZero()
		if !valid {
			name := v.Type().Field(i).Name
			return fmt.Errorf("field: %s cannot be empty", name)
		}
	}

	return nil
}

// TODO validate PHC
func ValidatePHC(password string, phc string) error {
	p, s, h, e := DecodePHC(phc)
	if e != nil {
		return e
	}

	k := argon2.IDKey([]byte(password), s, p.iterations, p.memory, p.threads, p.keyLen)

	// ctc to protect against timing attacks
	if subtle.ConstantTimeCompare(h, k) != 1 {
		e = errors.New("PHC are not equivalent")
		return e
	}

	return nil
}

func GeneratePHC(password string) (string, error) {
	salt, e := GenerateSalt(params.saltLen)
	if e != nil {
		return "", e
	}

	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.threads, params.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.memory, params.iterations, params.threads, b64Salt, b64Hash,
	)

	return encodedHash, nil
}

func DecodePHC(phc string) (p *HashParams, s, h []byte, e error) {
	values := strings.Split(phc, "$")
	if len(values) != 6 {
		e = errors.New("invalid PHC")
		return nil, nil, nil, e
	}

	var version int
	_, e = fmt.Sscanf(values[2], "v=%d", &version)
	if e != nil {
		return nil, nil, nil, e
	}
	if version != argon2.Version {
		e = errors.New("invalid version")
		return nil, nil, nil, e
	}

	p = &HashParams{}
	_, e = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.threads)
	if e != nil {
		e = errors.New("invalid parameters")
		return nil, nil, nil, e
	}

	s, e = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if e != nil {
		return nil, nil, nil, e
	}
	p.saltLen = uint32(len(s))

	h, e = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if e != nil {
		return nil, nil, nil, e
	}
	p.keyLen = uint32(len(h))

	return p, s, h, nil
}

func GenerateSalt(s uint32) ([]byte, error) {
	b := make([]byte, s)
	_, e := rand.Read(b)
	if e != nil {
		return nil, e
	}

	return b, nil
}

func ParseField(n string) (string, error) {
	// injection check
	for _, r := range n {
		c := string(r)
		if strings.Contains(invalidRunes, c) {
			return "", fmt.Errorf("invalid character in name: %v", c)
		}
	}

	// empty field check
	nTrim := strings.Trim(n, " ")
	if nTrim == "" {
		return "", errors.New("name cannot be empty or whitespace")
	}

	return n, nil
}
