package admin

import (
	"net/http"

	"github.com/NexaCard/API/internal/http/handlers/shared"
	"github.com/NexaCard/API/internal/http/response"
	"github.com/NexaCard/API/internal/logger"
	"github.com/NexaCard/API/internal/service"

	"github.com/gin-gonic/gin"
)

// ====================  文件上传  ====================

const (
	defaultAdminUploadMaxSize    = 10 << 20
	adminUploadMultipartOverhead = 1 << 20
)

// UploadFile 文件上传
func (h *Handler) UploadFile(c *gin.Context) {
	if c.Request != nil && c.Request.Body != nil {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.uploadBodyLimit())
	}
	file, err := c.FormFile("file")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.file_missing", nil)
		return
	}
	scene := c.DefaultPostForm("scene", "common")

	// 保存文件并获取元数据
	result, err := h.UploadService.SaveFileWithMeta(file, scene)
	if err != nil {
		if service.IsUploadValidationError(err) {
			shared.RespondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.upload_failed", err)
		return
	}

	// 记录到素材库
	var mediaID uint
	media, err := h.MediaService.RecordMedia(result, scene)
	if err != nil {
		logger.Warnw("upload_record_media_failed", "error", err, "url", result.URL)
	} else if media != nil {
		mediaID = media.ID
	}

	response.Success(c, gin.H{
		"url":      result.URL,
		"filename": result.Filename,
		"size":     result.Size,
		"media_id": mediaID,
	})
}

func (h *Handler) uploadBodyLimit() int64 {
	maxSize := int64(defaultAdminUploadMaxSize)
	if h != nil && h.Container != nil && h.Config != nil && h.Config.Upload.MaxSize > 0 {
		maxSize = h.Config.Upload.MaxSize
	}
	return maxSize + adminUploadMultipartOverhead
}
