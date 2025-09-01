package handlers

import (
	"delayednotifier/internal/entities/notification"
	"delayednotifier/internal/entities/request"
	"delayednotifier/internal/entities/response"
	"delayednotifier/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type notifyer interface {
	CreateNotification(n notification.Notification) (int64, error)
	Notification(id int64) (notification.Notification, error)
	DeleteNotification(id int64) error
	UpdateNotificationStatus(status string, id int64) error
}

// Main godoc
// @Summary вывести главную страницу с формай
// @Tags frontend
// @Produce html
// @Success 200
// @Router / [get]
func Main(s notifyer) gin.HandlerFunc {
	return func(c *ginext.Context) {
		c.HTML(http.StatusOK, "main.html", nil)
	}
}

// CreateNotify godoc
// @Summary Создать уведомление
// @Description Создание нового уведомления в очереди, date: RFC3339 UTC
// @Tags notifications
// @Accept json
// @Produce json
// @Param notify body request.CreateNotify true "Данные уведомления"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 504 {object} response.Response
// @Router /notify [post]
func CreateNotify(s notifyer) gin.HandlerFunc {
	return func(c *ginext.Context) {
		const op = "internal.handlers.CreateNotify"

		var r request.CreateNotification
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSONP(http.StatusBadRequest, response.Error(
				"wrong data types or fields in json ",
			))
			return
		}

		n, msg := r.Validate()
		if msg != "" {
			c.JSONP(http.StatusBadRequest, response.Error(
				msg,
			))
			return
		}

		id, err := s.CreateNotification(n)
		if errors.Is(err, service.ErrNotValidData) {
			c.JSONP(http.StatusServiceUnavailable, response.Error(
				err.Error(),
			))
			return
		} else if err != nil {
			zlog.Logger.Error().AnErr("err", err).Msg(op)
			c.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		c.JSONP(http.StatusOK, response.OK(
			id,
		))
	}
}

// GetNotify godoc
// @Summary Получить уведомление по ID
// @Description Получение информации о конкретном уведомлении
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "ID уведомления"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /notify/{id} [get]
func GetNotify(s notifyer) gin.HandlerFunc {
	return func(c *ginext.Context) {
		const op = "internal.handlers.GetNotify"

		idTmp := c.Param("id")
		id, err := strconv.ParseInt(idTmp, 10, 64)
		if err != nil {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be numeric value",
			))
			return
		}
		if id <= 0 {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be positive",
			))
			return
		}

		n, err := s.Notification(id)
		if errors.Is(err, service.ErrNotValidData) {
			c.JSONP(http.StatusServiceUnavailable, response.Error(
				err.Error(),
			))
			return
		} else if errors.Is(err, service.ErrNotFound) {
			c.JSONP(http.StatusNotFound, response.Error(
				err.Error(),
			))
			return
		} else if err != nil {
			zlog.Logger.Error().AnErr("err", err).Msg(op)
			c.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		c.JSONP(http.StatusOK, response.OK(
			n,
		))
	}
}

// DeleteNotify godoc
// @Summary удалить уведомление по ID
// @Description удаление конкретного уведомления из очереди
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "ID уведомления"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /notify/{id} [delete]
func DeleteNotify(s notifyer) gin.HandlerFunc {
	return func(c *ginext.Context) {
		const op = "internal.handlers.GetNotify"

		idTmp := c.Param("id")
		id, err := strconv.ParseInt(idTmp, 10, 64)
		if err != nil {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be numeric value",
			))
			return
		}
		if id <= 0 {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be positive",
			))
			return
		}

		err = s.DeleteNotification(id)
		if errors.Is(err, service.ErrNotValidData) {
			c.JSONP(http.StatusServiceUnavailable, response.Error(
				err.Error(),
			))
			return
		} else if errors.Is(err, service.ErrNotAffected) {
			c.JSONP(http.StatusNotFound, response.Error(
				err.Error(),
			))
			return
		} else if err != nil {
			zlog.Logger.Error().AnErr("err", err).Msg(op)
			c.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		c.JSONP(http.StatusOK, response.OK(
			"successfull deleted",
		))
	}
}

// Update godoc
// @Summary обновить статус по ID
// @Description обновить конкретное уведомление из очереди
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "ID уведомления"
// @Param status body string true "complete || pending"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /notify/{id} [patch]
func UpdateNotify(s notifyer) gin.HandlerFunc {
	return func(c *ginext.Context) {
		const op = "internal.handlers.UpdateNotify"

		idTmp := c.Param("id")
		id, err := strconv.ParseInt(idTmp, 10, 64)
		if err != nil {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be numeric value",
			))
			return
		}
		if id <= 0 {
			c.JSONP(http.StatusBadRequest, response.Error(
				"id should be positive",
			))
			return
		}
		var r request.UpdateNotification
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSONP(http.StatusBadRequest, response.Error(
				"wrong data types or fields in json ",
			))
			return
		}
		status, msg := r.Validate()
		if msg != "" {
			c.JSONP(http.StatusBadRequest, response.Error(
				msg,
			))
			return
		}

		err = s.UpdateNotificationStatus(status, id)
		if errors.Is(err, service.ErrNotValidData) {
			c.JSONP(http.StatusServiceUnavailable, response.Error(
				err.Error(),
			))
			return
		} else if errors.Is(err, service.ErrNotAffected) {
			c.JSONP(http.StatusNotFound, response.Error(
				err.Error(),
			))
			return
		} else if err != nil {
			zlog.Logger.Error().AnErr("err", err).Msg(op)
			c.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		c.JSONP(http.StatusOK, response.OK(
			"successfull updated",
		))
	}
}
