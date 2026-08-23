package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-090/internal/application/services"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/middleware"
)

type Server struct {
	router        *gin.Engine
	events        *services.EventService
	matching      *services.MatchingService
	invitations   *services.InvitationService
	registrations *services.RegistrationService
	validate      *validator.Validate
	startedAt     time.Time
}

func NewServer(events *services.EventService, matching *services.MatchingService, invitations *services.InvitationService, registrations *services.RegistrationService) *Server {
	return &Server{events: events, matching: matching, invitations: invitations, registrations: registrations, validate: validator.New(), startedAt: time.Now().UTC()}
}
func (s *Server) Router() *gin.Engine {
	if s.router != nil {
		return s.router
	}
	r := gin.New()
	r.Use(middleware.RequestContext(), middleware.Recovery(), middleware.SecurityHeaders(), middleware.CORS())
	r.GET("/healthz", s.health)
	r.GET("/readyz", s.ready)
	api := r.Group("/api/v1")
	api.GET("/events", s.listEvents)
	api.POST("/events/:id/submit", s.submitEvent)
	api.POST("/events/:id/approve", s.approveEvent)
	api.POST("/events/:id/cancel", s.cancelEvent)
	api.GET("/needs/:id/matches", s.matches)
	api.POST("/invitations/:id/accept", s.acceptInvitation)
	api.POST("/events/:id/registrations", s.register)
	s.router = r
	return r
}
func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "salsa-platform", "started_at": s.startedAt})
}
func (s *Server) ready(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) }
func (s *Server) listEvents(c *gin.Context) {
	page := common.PageRequest{Page: parseInt(c.Query("page")), Limit: parseInt(c.Query("limit")), Sort: c.Query("sort"), Filters: map[string]string{}}
	result, err := s.events.List(c.Request.Context(), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (s *Server) submitEvent(c *gin.Context) {
	version, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil {
		writeError(c, common.FieldError("version", "must be an integer"))
		return
	}
	if err := s.events.Submit(c.Request.Context(), c.Param("id"), actor(c), version); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (s *Server) approveEvent(c *gin.Context) {
	version, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil {
		writeError(c, common.FieldError("version", "must be an integer"))
		return
	}
	if err := s.events.Approve(c.Request.Context(), c.Param("id"), actor(c), version); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type cancelRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (s *Server) cancelEvent(c *gin.Context) {
	var req cancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, common.FieldError("reason", "is required"))
		return
	}
	version, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil {
		writeError(c, common.FieldError("version", "must be an integer"))
		return
	}
	if err := s.events.Cancel(c.Request.Context(), c.Param("id"), actor(c), req.Reason, version); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (s *Server) matches(c *gin.Context) {
	results, err := s.matching.Find(c.Request.Context(), c.Param("id"), map[string]bool{})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": results, "statistic": "style 35, distance 20, tags 10 each, level 15"})
}
func (s *Server) acceptInvitation(c *gin.Context) {
	version, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil {
		writeError(c, common.FieldError("version", "must be an integer"))
		return
	}
	if err := s.invitations.Accept(c.Request.Context(), c.Param("id"), actor(c), version); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (s *Server) register(c *gin.Context) {
	result, err := s.registrations.Register(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}
func parseInt(value string) int {
	if value == "" {
		return 0
	}
	n, _ := strconv.Atoi(value)
	return n
}

var _ = event.StatusPublished
