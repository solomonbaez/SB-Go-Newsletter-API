package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func (rh *RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []*models.Subscriber
	requestID := c.GetString("requestID")

	var response string
	var e error

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscribers...")

	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response = "Failed to fetch subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[*models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response = "Failed to parse subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	if len(subscribers) > 0 {
		for _, s := range subscribers {
			// re-parse email to ensure data integrity
			s.Email, e = models.ParseEmail(s.Email.String())
			if e != nil {
				response = "Failed to parse subscriber"
				HandleError(c, requestID, e, response, http.StatusInternalServerError)
				continue
			}
		}

		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": subscribers})
	} else {
		response = "No subscribers"
		log.Info().
			Str("requestID", requestID).
			Msg(response)

		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": response})
	}
}

func (rh *RouteHandler) GetSubscriberByID(c *gin.Context) {
	requestID := c.GetString("requestID")

	var response string
	var e error

	log.Info().
		Str("requestID", requestID).
		Msg("Validating ID...")

	// Validate UUID
	u := c.Param("id")
	id, e := uuid.Parse(u)
	if e != nil {
		response = "Invalid ID format"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscriber...")

	var subscriber models.Subscriber
	e = rh.DB.QueryRow(c, "SELECT id, email, name, status FROM subscriptions WHERE id=$1", id).Scan(&subscriber.ID, &subscriber.Email, &subscriber.Name, &subscriber.Status)
	if e != nil {
		if e == pgx.ErrNoRows {
			response = "Subscriber not found"
		} else {
			response = "Database query error"
		}
		HandleError(c, requestID, e, response, http.StatusNotFound)
		return
	}

	subscriber.Email, e = models.ParseEmail(subscriber.Email.String())
	if e != nil {
		response = "Invalid email"
		HandleError(c, requestID, e, response, http.StatusConflict)
		return
	}

	c.JSON(http.StatusFound, gin.H{"requestID": requestID, "subscriber": subscriber})
}

func (rh RouteHandler) GetConfirmedSubscribers(c *gin.Context) []*models.Subscriber {
	var subscribers []*models.Subscriber
	requestID := c.GetString("requestID")

	var response string
	var e error

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching confirmed subscribers...")

	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions WHERE status=$1", "confirmed")
	if e != nil {
		response = "Failed to fetch confirmed subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[*models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response = "Failed to parse confirmed subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return nil
	}

	if len(subscribers) > 0 {
		for _, s := range subscribers {
			s.Email, e = models.ParseEmail(s.Email.String())
			if e != nil {
				response = "Invalid email"
				HandleError(c, requestID, e, response, http.StatusConflict)
				return nil
			}
		}
		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": subscribers})
		return subscribers
	} else {
		response = "No confirmed subscribers"
		log.Info().
			Str("requestID", requestID).
			Msg(response)

		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": response})
		return nil
	}
}
