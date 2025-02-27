package webhook

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"go.uber.org/zap"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

const (
	contentTypeHeader    = "Content-Type"
	contentTypePlaintext = "text/plain"
	acceptHeader         = "Accept"
	varyHeader           = "Vary"

	errorAcceptHeader = "client must provide an accept header"
	errorContentType  = "client must provide a content type"
)

type Webhook struct {
	provider provider.Provider
}

func New(provider provider.Provider) *Webhook {
	return &Webhook{provider: provider}
}

func (w *Webhook) contentTypeHeaderCheck(c *gin.Context) error {
	header := c.GetHeader(contentTypeHeader)
	if len(header) == 0 {
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": errorContentType})

		logger.Log.Error(errorContentType, zap.String("header", header))
		return errors.New(errorContentType)
	}

	return w.headerCheck(header, c)
}

func (w *Webhook) acceptHeaderCheck(c *gin.Context) error {
	header := c.GetHeader(acceptHeader)
	if len(header) == 0 {
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": errorAcceptHeader})

		logger.Log.Error(errorAcceptHeader, zap.String("header", header))
		return errors.New(errorAcceptHeader)
	}

	return w.headerCheck(header, c)
}

func (w *Webhook) headerCheck(header string, c *gin.Context) error {
	// as we support only one media type version, we can ignore the returned value
	if _, err := checkAndGetMediaTypeHeaderValue(header); err != nil {
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "client must provide a valid versioned media type"})

		logger.Log.Error("client must provide a valid versioned media type", zap.String("header", header), zap.Error(err))
		return err
	}

	return nil
}

func (w *Webhook) Records(c *gin.Context) {
	if err := w.acceptHeaderCheck(c); err != nil {
		logger.Log.Error("accept header check failed", zap.Error(err))
		return
	}

	records, err := w.provider.Records(c.Request.Context())
	if err != nil {
		logger.Log.Error("error getting records", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(contentTypeHeader, string(mediaTypeVersion1))
	c.Header(varyHeader, contentTypeHeader)
	if err = json.NewEncoder(c.Writer).Encode(records); err != nil {
		logger.Log.Error("error encoding records", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func (w *Webhook) ApplyChanges(c *gin.Context) {
	if err := w.contentTypeHeaderCheck(c); err != nil {
		logger.Log.Error("content type header check failed", zap.Error(err))
		return
	}

	var changes plan.Changes
	if err := json.NewDecoder(c.Request.Body).Decode(&changes); err != nil {
		logger.Log.Error("error decoding changes", zap.Error(err))
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error decoding changes"})
		return
	}

	logger.Log.Debug("requesting apply changes", zap.Int("create", len(changes.Create)),
		zap.Int("update_old", len(changes.UpdateOld)), zap.Int("update_new", len(changes.UpdateNew)),
		zap.Int("delete", len(changes.Delete)))

	if err := w.provider.ApplyChanges(c.Request.Context(), &changes); err != nil {
		logger.Log.Error("error when applying changes", zap.Error(err))
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO: check if it is required
	c.Status(http.StatusNoContent)
}

func (w *Webhook) AdjustEndpoints(c *gin.Context) {
	if err := w.contentTypeHeaderCheck(c); err != nil {
		logger.Log.Error("content type header check failed", zap.Error(err))
		return
	}

	if err := w.acceptHeaderCheck(c); err != nil {
		logger.Log.Error("accept header check failed", zap.Error(err))
		return
	}

	var pve []*endpoint.Endpoint
	if err := json.NewDecoder(c.Request.Body).Decode(&pve); err != nil {
		c.Header(contentTypeHeader, contentTypePlaintext)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error decoding request body"})
		return
	}

	logger.Log.Debug("webhook adjust endpoints", zap.Int("endpoints", len(pve)))
	c.Header(contentTypeHeader, contentTypePlaintext)
	pve, err := w.provider.AdjustEndpoints(pve)
	if err != nil {
		c.Header(varyHeader, contentTypeHeader)
		logger.Log.Error("error adjusting endpoints", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(contentTypeHeader, string(mediaTypeVersion1))
	c.Header(varyHeader, contentTypeHeader)

	out, err := json.Marshal(&pve)
	if err != nil {
		logger.Log.Error("error encoding adjusted endpoints", zap.Error(err))
	}

	if _, err := c.Writer.Write(out); err != nil {
		logger.Log.Error("error writing response", zap.Error(err))
	}

	logger.Log.Debug("adjusted endpoints", zap.Int("endpoints", len(pve)))
}

func (w *Webhook) Negotiate(c *gin.Context) {
	if err := w.acceptHeaderCheck(c); err != nil {
		logger.Log.Error("accept header check failed", zap.Error(err))
		return
	}

	b, err := json.Marshal(w.provider.GetDomainFilter())
	if err != nil {
		logger.Log.Error("error marshaling domain filter", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(contentTypeHeader, string(mediaTypeVersion1))

	_, err = c.Writer.Write(b)
	if err != nil {
		logger.Log.Error("error writing response", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
