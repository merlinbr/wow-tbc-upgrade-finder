package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/wowsims/tbc/internal/forever/domain"
	"github.com/wowsims/tbc/internal/forever/service"
	"github.com/wowsims/tbc/internal/forever/store"
	"github.com/wowsims/tbc/internal/forever/validate"
)

const maxBodyBytes = 1 << 20

func (a *api) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	characters, err := a.service.ListCharacters(r.Context())
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Characters []domain.Character `json:"characters"`
	}{Characters: characters})
}

func (a *api) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	var input service.CreateCharacterInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	character, err := a.service.CreateCharacter(r.Context(), input)
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, character)
}

func (a *api) handleGetCharacter(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	character, err := a.service.GetCharacter(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, character)
}

func (a *api) handleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	var input service.UpdateCharacterInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	character, warnings, err := a.service.UpdateCharacter(r.Context(), r.PathValue("id"), input)
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	if warnings == nil {
		warnings = []validate.Violation{}
	}
	writeJSON(w, http.StatusOK, struct {
		Character domain.Character     `json:"character"`
		Warnings  []validate.Violation `json:"warnings"`
	}{Character: character, Warnings: warnings})
}

func (a *api) handleListBuilds(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	builds, err := a.service.ListBuilds(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Builds []domain.Build `json:"builds"`
	}{Builds: builds})
}

func (a *api) handleCreateBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	var input service.CreateBuildInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	build, err := a.service.CreateBuild(r.Context(), r.PathValue("id"), input)
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, build)
}

func (a *api) handleGetBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	build, err := a.service.GetBuild(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, build)
}

func (a *api) handleUpdateBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	var input service.UpdateBuildInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	build, violations, err := a.service.UpdateBuild(r.Context(), r.PathValue("id"), input)
	if err != nil {
		if len(violations) > 0 {
			writeValidationError(w, violations)
			return
		}
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, build)
}

func (a *api) handleDeleteBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	if err := a.service.DeleteBuild(r.Context(), r.PathValue("id")); err != nil {
		a.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) handleCopyBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	build, err := a.service.CopyBuild(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, build)
}

func (a *api) handleValidateBuild(w http.ResponseWriter, r *http.Request) {
	if !a.ensureService(w) {
		return
	}
	character, err := a.service.GetCharacter(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	var input struct {
		Equipment domain.Equipment `json:"equipment"`
		Talents   domain.Talents   `json:"talents"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	violations, err := a.service.ValidateBuild(character, input.Equipment, input.Talents)
	if err != nil {
		a.writeServiceError(w, err)
		return
	}
	if violations == nil {
		violations = []validate.Violation{}
	}
	writeJSON(w, http.StatusOK, struct {
		Violations []validate.Violation `json:"violations"`
	}{Violations: violations})
}

func (a *api) ensureService(w http.ResponseWriter) bool {
	if a.service == nil {
		writeError(w, http.StatusServiceUnavailable, "service_unavailable", "persistence is not configured")
		return false
	}
	return true
}

func (a *api) writeServiceError(w http.ResponseWriter, err error) {
	var validation *service.ValidationError
	if errors.As(err, &validation) {
		payload := struct {
			Error struct {
				Code       string               `json:"code"`
				Message    string               `json:"message"`
				Violations []validate.Violation `json:"violations,omitempty"`
			} `json:"error"`
		}{}
		payload.Error.Code = validation.Code
		payload.Error.Message = validation.Message
		payload.Error.Violations = validation.Violations
		writeJSON(w, http.StatusUnprocessableEntity, payload)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "record not found")
	case errors.Is(err, store.ErrVersionConflict):
		writeError(w, http.StatusConflict, "version_conflict", "this build changed elsewhere")
	case errors.Is(err, store.ErrCannotDeleteCurrent):
		writeError(w, http.StatusConflict, "cannot_delete_current", "the current build cannot be deleted")
	case errors.Is(err, domain.ErrUnsupportedClass):
		writeError(w, http.StatusBadRequest, "unsupported_class", "unknown class")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeValidationError(w http.ResponseWriter, violations []validate.Violation) {
	payload := struct {
		Error struct {
			Code       string               `json:"code"`
			Message    string               `json:"message"`
			Violations []validate.Violation `json:"violations"`
		} `json:"error"`
	}{}
	payload.Error.Code = "invalid_build"
	payload.Error.Message = "The build did not pass validation."
	payload.Error.Violations = violations
	writeJSON(w, http.StatusUnprocessableEntity, payload)
}

// decodeJSON rejects unknown fields, oversized bodies and trailing content.
func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds 1 MiB")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body: "+err.Error())
		}
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_request", "multiple JSON values in body")
		return errors.New("multiple JSON values in body")
	}
	return nil
}
