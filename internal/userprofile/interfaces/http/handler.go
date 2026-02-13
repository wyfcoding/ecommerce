package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/userprofile/application"
	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
)

type UserProfileHandler struct {
	commandService *application.ProfileCommandService
	queryService   *application.ProfileQueryService
}

func NewUserProfileHandler(
	commandService *application.ProfileCommandService,
	queryService *application.ProfileQueryService,
) *UserProfileHandler {
	return &UserProfileHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

func (h *UserProfileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/user-profiles", h.handleProfiles)
	mux.HandleFunc("/api/v1/user-profiles/", h.handleProfileByID)
	mux.HandleFunc("/api/v1/user-profiles/by-tag", h.GetUsersByTag)
	mux.HandleFunc("/api/v1/user-profiles/by-segment/", h.GetUsersBySegment)
}

func (h *UserProfileHandler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateProfile(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserProfileHandler) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user-profiles/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getProfile(w, r, userID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch parts[1] {
	case "tags":
		h.handleTags(w, r, userID, parts)
	case "behaviors":
		h.handleBehaviors(w, r, userID)
	case "behavior-features":
		if r.Method == http.MethodGet {
			h.getBehaviorFeatures(w, r, userID)
		}
	case "preferences":
		if r.Method == http.MethodGet {
			h.getPreferences(w, r, userID)
		}
	case "consumption":
		if r.Method == http.MethodGet {
			h.getConsumptionProfile(w, r, userID)
		}
	case "recalculate":
		if r.Method == http.MethodPost {
			h.recalculateProfile(w, r, userID)
		}
	case "summary":
		if r.Method == http.MethodGet {
			h.getProfileSummary(w, r, userID)
		}
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *UserProfileHandler) handleTags(w http.ResponseWriter, r *http.Request, userID uint64, parts []string) {
	if len(parts) == 2 {
		switch r.Method {
		case http.MethodGet:
			h.getTags(w, r, userID)
		case http.MethodPost:
			h.addTag(w, r, userID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) >= 3 && r.Method == http.MethodDelete {
		h.removeTag(w, r, userID, parts[2])
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (h *UserProfileHandler) handleBehaviors(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method == http.MethodPost {
		h.recordBehavior(w, r, userID)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint64 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	profile, err := h.commandService.CreateProfile(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *UserProfileHandler) getProfile(w http.ResponseWriter, r *http.Request, userID uint64) {
	profile, err := h.queryService.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *UserProfileHandler) getTags(w http.ResponseWriter, r *http.Request, userID uint64) {
	tags, err := h.queryService.GetTags(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *UserProfileHandler) addTag(w http.ResponseWriter, r *http.Request, userID uint64) {
	var req struct {
		TagKey     string  `json:"tag_key"`
		TagValue   string  `json:"tag_value"`
		Category   int8    `json:"category"`
		Source     int8    `json:"source"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.commandService.AddTag(r.Context(), userID, req.TagKey, req.TagValue, domain.TagCategory(req.Category), domain.TagSource(req.Source), req.Confidence); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) removeTag(w http.ResponseWriter, r *http.Request, userID uint64, tagKey string) {
	if err := h.commandService.RemoveTag(r.Context(), userID, tagKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) recordBehavior(w http.ResponseWriter, r *http.Request, userID uint64) {
	var req struct {
		BehaviorType string `json:"behavior_type"`
		TargetType   string `json:"target_type"`
		TargetID     uint64 `json:"target_id"`
		Value        string `json:"value"`
		Duration     int64  `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.commandService.RecordBehavior(r.Context(), userID, domain.BehaviorType(req.BehaviorType), req.TargetType, req.TargetID, req.Value, req.Duration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) getBehaviorFeatures(w http.ResponseWriter, r *http.Request, userID uint64) {
	features, err := h.queryService.GetBehaviorFeatures(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

func (h *UserProfileHandler) getPreferences(w http.ResponseWriter, r *http.Request, userID uint64) {
	preferences, err := h.queryService.GetPreferences(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferences)
}

func (h *UserProfileHandler) getConsumptionProfile(w http.ResponseWriter, r *http.Request, userID uint64) {
	consumption, err := h.queryService.GetConsumptionProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(consumption)
}

func (h *UserProfileHandler) recalculateProfile(w http.ResponseWriter, r *http.Request, userID uint64) {
	if err := h.commandService.RecalculateProfile(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) getProfileSummary(w http.ResponseWriter, r *http.Request, userID uint64) {
	summary, err := h.queryService.GetProfileSummary(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *UserProfileHandler) GetUsersByTag(w http.ResponseWriter, r *http.Request) {
	tagKey := r.URL.Query().Get("tag_key")
	tagValue := r.URL.Query().Get("tag_value")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	profiles, err := h.queryService.GetUsersByTag(r.Context(), tagKey, tagValue, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func (h *UserProfileHandler) GetUsersBySegment(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user-profiles/by-segment/")
	segmentNo := path

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	profiles, err := h.queryService.GetSegmentUsers(r.Context(), segmentNo, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}
