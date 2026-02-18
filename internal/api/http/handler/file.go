package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	svcfile "github.com/Alijeyrad/simorq_backend/internal/service/file"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type FileHandler struct {
	svc svcfile.Service
}

func NewFileHandler(svc svcfile.Service) *FileHandler {
	return &FileHandler{svc: svc}
}

// POST /files/upload
// Multipart upload; returns {key, file_name, size, mime_type}.
func (h *FileHandler) Upload(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return badRequest(c, "file field is required")
	}

	result, err := h.svc.Upload(c.Context(), clinicID, fh)
	if err != nil {
		return internalError(c)
	}

	return created(c, fiber.Map{
		"key":       result.Key,
		"file_name": result.FileName,
		"size":      result.Size,
		"mime_type": result.MimeType,
	})
}

// GET /files/:key
// Returns a presigned download URL (redirect).
func (h *FileHandler) GetByKey(c fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return badRequest(c, "key is required")
	}

	url, err := h.svc.GetDownloadURL(c.Context(), key)
	if err != nil {
		return internalError(c)
	}

	return c.Redirect().To(url)
}

// GET /patients/:id/files
func (h *FileHandler) ListPatientFiles(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	files, err := h.svc.ListPatientFiles(c.Context(), clinicID, patientID)
	if err != nil {
		return internalError(c)
	}

	return ok(c, files)
}

// POST /patients/:id/files
// Multipart upload + create PatientFile DB record.
func (h *FileHandler) UploadPatientFile(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	claims, authOK := pasetotoken.ClaimsFromFiber(c)
	if !authOK {
		return unauthorized(c)
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return badRequest(c, "file field is required")
	}

	// Upload to S3 first
	uploaded, err := h.svc.Upload(c.Context(), clinicID, fh)
	if err != nil {
		return internalError(c)
	}

	// Optional metadata fields
	var linkedType *string
	var linkedID *uuid.UUID
	var description *string

	if lt := c.FormValue("linked_type"); lt != "" {
		linkedType = &lt
	}
	if li := c.FormValue("linked_id"); li != "" {
		id, err := uuid.Parse(li)
		if err != nil {
			return badRequest(c, "invalid linked_id")
		}
		linkedID = &id
	}
	if d := c.FormValue("description"); d != "" {
		description = &d
	}

	pf, err := h.svc.CreatePatientFile(c.Context(), clinicID, patientID, claims.UserID, svcfile.CreatePatientFileRequest{
		Key:         uploaded.Key,
		FileName:    uploaded.FileName,
		Size:        uploaded.Size,
		MimeType:    uploaded.MimeType,
		LinkedType:  linkedType,
		LinkedID:    linkedID,
		Description: description,
	})
	if err != nil {
		return internalError(c)
	}

	return created(c, pf)
}

// GET /patients/:id/files/:fid/download
// Returns a presigned download URL (redirect).
func (h *FileHandler) DownloadPatientFile(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	fileID, err := uuid.Parse(c.Params("fid"))
	if err != nil {
		return badRequest(c, "invalid file id")
	}

	files, err := h.svc.ListPatientFiles(c.Context(), clinicID, patientID)
	if err != nil {
		return internalError(c)
	}

	var fileKey string
	for _, f := range files {
		if f.ID == fileID {
			fileKey = f.FileKey
			break
		}
	}
	if fileKey == "" {
		return notFound(c, "file not found")
	}

	url, err := h.svc.GetDownloadURL(c.Context(), fileKey)
	if err != nil {
		return internalError(c)
	}

	return c.Redirect().To(url)
}

// DELETE /patients/:id/files/:fid
func (h *FileHandler) DeletePatientFile(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	fileID, err := uuid.Parse(c.Params("fid"))
	if err != nil {
		return badRequest(c, "invalid file id")
	}

	if err := h.svc.DeletePatientFile(c.Context(), clinicID, patientID, fileID); err != nil {
		if errors.Is(err, svcfile.ErrPatientFileNotFound) {
			return notFound(c, err.Error())
		}
		return internalError(c)
	}

	return noContent(c)
}
