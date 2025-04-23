package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
)

type URLHandler struct {
	repo    repository.URLRepository
	metrics *middleware.Metrics
}

func NewURLHandler(repo repository.URLRepository, metrics *middleware.Metrics) *URLHandler {
	return &URLHandler{
		repo:    repo,
		metrics: metrics,
	}
}

type CreateShortURLRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type CreateShortURLResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *URLHandler) CreateShortURL(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateShortURLRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid URL format")
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	shortCode, err := h.repo.Create(ctx, parsedURL.String())
	if err != nil {
		return fmt.Errorf("failed to create short URL: %w", err)
	}

	h.metrics.URLsCreated.Inc()

	return c.JSON(http.StatusCreated, CreateShortURLResponse{
		ShortURL: buildShortURL(c, shortCode),
	})
}

func (h *URLHandler) RedirectByCode(c echo.Context) error {
	ctx := c.Request().Context()
	shortCode := c.Param("code")

	originalURL, err := h.repo.GetOriginalURL(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "short URL not found")
		}
		return fmt.Errorf("failed to get original URL: %w", err)
	}

	return c.Redirect(http.StatusMovedPermanently, originalURL)
}

func buildShortURL(c echo.Context, code string) string {
	return c.Scheme() + "://" + c.Request().Host + "/" + code
}
