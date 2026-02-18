package file

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entfile "github.com/Alijeyrad/simorq_backend/internal/repo/patientfile"
	s3pkg "github.com/Alijeyrad/simorq_backend/pkg/s3"
)

var (
	ErrPatientFileNotFound = errors.New("patient file not found")
	ErrAccessDenied        = errors.New("access denied")
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type UploadResult struct {
	Key      string
	FileName string
	Size     int64
	MimeType string
}

type CreatePatientFileRequest struct {
	Key         string
	FileName    string
	Size        int64
	MimeType    string
	LinkedType  *string
	LinkedID    *uuid.UUID
	Description *string
}

// ---------------------------------------------------------------------------
// Service interface
// ---------------------------------------------------------------------------

type Service interface {
	Upload(ctx context.Context, clinicID uuid.UUID, f *multipart.FileHeader) (*UploadResult, error)
	CreatePatientFile(ctx context.Context, clinicID, patientID, uploaderID uuid.UUID, req CreatePatientFileRequest) (*repo.PatientFile, error)
	ListPatientFiles(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientFile, error)
	GetDownloadURL(ctx context.Context, fileKey string) (string, error)
	DeletePatientFile(ctx context.Context, clinicID, patientID, fileID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type fileService struct {
	db  *repo.Client
	s3  *s3pkg.Client
}

func New(db *repo.Client, s3Client *s3pkg.Client) Service {
	return &fileService{db: db, s3: s3Client}
}

func (s *fileService) Upload(ctx context.Context, clinicID uuid.UUID, fh *multipart.FileHeader) (*UploadResult, error) {
	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	key := fmt.Sprintf("uploads/%s/%s%s", clinicID, uuid.New(), ext)

	mime := fh.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	if err := s.s3.Upload(ctx, key, mime, src, fh.Size); err != nil {
		return nil, fmt.Errorf("s3 upload: %w", err)
	}

	return &UploadResult{
		Key:      key,
		FileName: fh.Filename,
		Size:     fh.Size,
		MimeType: mime,
	}, nil
}

func (s *fileService) CreatePatientFile(ctx context.Context, clinicID, patientID, uploaderID uuid.UUID, req CreatePatientFileRequest) (*repo.PatientFile, error) {
	c := s.db.PatientFile.Create().
		SetPatientID(patientID).
		SetClinicID(clinicID).
		SetUploadedBy(uploaderID).
		SetFileKey(req.Key).
		SetFileName(req.FileName)

	if req.Size > 0 {
		c = c.SetFileSize(req.Size)
	}
	if req.MimeType != "" {
		c = c.SetMimeType(req.MimeType)
	}
	if req.LinkedType != nil {
		c = c.SetNillableLinkedType(req.LinkedType)
	}
	if req.LinkedID != nil {
		c = c.SetNillableLinkedID(req.LinkedID)
	}
	if req.Description != nil {
		c = c.SetNillableDescription(req.Description)
	}

	return c.Save(ctx)
}

func (s *fileService) ListPatientFiles(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientFile, error) {
	return s.db.PatientFile.Query().
		Where(entfile.PatientID(patientID), entfile.ClinicID(clinicID)).
		Order(entfile.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
}

func (s *fileService) GetDownloadURL(ctx context.Context, fileKey string) (string, error) {
	url, err := s.s3.PresignDownload(ctx, fileKey)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return url, nil
}

func (s *fileService) DeletePatientFile(ctx context.Context, clinicID, patientID, fileID uuid.UUID) error {
	f, err := s.db.PatientFile.Query().
		Where(entfile.ID(fileID), entfile.PatientID(patientID), entfile.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrPatientFileNotFound
		}
		return fmt.Errorf("get patient file: %w", err)
	}

	// Best-effort S3 delete (don't block DB delete if S3 fails)
	_ = s.s3.Delete(ctx, f.FileKey)

	return s.db.PatientFile.DeleteOne(f).Exec(ctx)
}
